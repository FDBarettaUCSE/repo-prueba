package reporte

import (
	ordendecompra "api/internal/ordenDeCompra"
	"api/internal/transferencia"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ReporteRepositorio define las consultas utilizadas por el módulo de
// reportes. Los filtros, agrupaciones y la paginación se ejecutan en MongoDB.
// El Service se encarga de las reglas de negocio, el alcance por depósito,
// el tamaño fijo de página y la conversión a DTO.
type ReporteRepositorio interface {
	InventarioPorDeposito(
		ctx context.Context,
		filtro FiltroInventario,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoInventarioDeposito, int64, error)

	InventarioPorCategoria(
		ctx context.Context,
		filtro FiltroInventario,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoInventarioCategoria, int64, error)

	Movimientos(
		ctx context.Context,
		filtro FiltroReporteMovimientos,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoMovimiento, int64, error)

	AlertasBajoMinimo(
		ctx context.Context,
		filtro FiltroAlertas,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoAlertaMinimo, int64, error)

	AlertasPorVencer(
		ctx context.Context,
		filtro FiltroAlertas,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoAlertaVencimiento, int64, error)

	OrdenesPendientes(
		ctx context.Context,
		filtro FiltroPendientes,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoOrdenesDeposito, int64, error)

	TransferenciasPendientes(
		ctx context.Context,
		filtro FiltroPendientes,
		pagina int,
		tamanioPagina int,
	) ([]ResultadoTransferenciasDeposito, int64, error)

	TotalesInventario(
		ctx context.Context,
		filtro FiltroInventario,
	) (int32, float64, error)
}

type RepositorioMongo struct {
	coleccionDepositoProducto *mongo.Collection
	coleccionProducto         *mongo.Collection
	coleccionDeposito         *mongo.Collection
	coleccionCategoria        *mongo.Collection
	coleccionMovimiento       *mongo.Collection
	coleccionOrdenCompra      *mongo.Collection
	coleccionTransferencia    *mongo.Collection
}

func NuevoRepositorioMongo(
	coleccionDepositoProducto *mongo.Collection,
	coleccionProducto *mongo.Collection,
	coleccionDeposito *mongo.Collection,
	coleccionCategoria *mongo.Collection,
	coleccionMovimiento *mongo.Collection,
	coleccionOrdenCompra *mongo.Collection,
	coleccionTransferencia *mongo.Collection,
) *RepositorioMongo {
	return &RepositorioMongo{
		coleccionDepositoProducto: coleccionDepositoProducto,
		coleccionProducto:         coleccionProducto,
		coleccionDeposito:         coleccionDeposito,
		coleccionCategoria:        coleccionCategoria,
		coleccionMovimiento:       coleccionMovimiento,
		coleccionOrdenCompra:      coleccionOrdenCompra,
		coleccionTransferencia:    coleccionTransferencia,
	}
}

var _ ReporteRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) InventarioPorDeposito(
	ctx context.Context,
	filtro FiltroInventario,
	pagina int,
	tamanioPagina int,
) ([]ResultadoInventarioDeposito, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	match := bson.D{}

	if filtro.DepositoID != nil {
		match = append(match, bson.E{
			Key:   "deposito_id",
			Value: *filtro.DepositoID,
		})
	}

	pipeline := mongo.Pipeline{
		// 1. Filtrar por depósito directamente sobre deposito_producto.
		{
			{Key: "$match", Value: match},
		},

		// 2. Obtener el producto.
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionProducto.Name()},
				{Key: "localField", Value: "producto_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "producto"},
			}},
		},
		{
			{Key: "$unwind", Value: "$producto"},
		},

		// 3. No incluir productos eliminados.
		{
			{Key: "$match", Value: bson.D{
				{Key: "producto.fecha_eliminacion", Value: nil},
			}},
		},
	}

	// 4. Filtrar por categoría si corresponde.
	if filtro.CategoriaID != nil {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "producto.categoria", Value: *filtro.CategoriaID},
			}},
		})
	}

	pipeline = append(pipeline,

		// 5. Obtener el depósito.
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionDeposito.Name()},
				{Key: "localField", Value: "deposito_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "deposito"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: "$deposito"},
		},

		// 6. No incluir depósitos eliminados.
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "deposito.fecha_eliminacion", Value: nil},
			}},
		},

		// 7. Agrupar por depósito.
		bson.D{
			{Key: "$group", Value: bson.D{
				{
					Key: "_id",
					Value: bson.D{
						{Key: "deposito_id", Value: "$deposito_id"},
						{Key: "deposito_nombre", Value: "$deposito.nombre"},
					},
				},
				{
					Key: "stock",
					Value: bson.D{
						{Key: "$sum", Value: "$stock"},
					},
				},
				{
					Key: "valorizado",
					Value: bson.D{
						{
							Key: "$sum",
							Value: bson.D{
								{Key: "$multiply", Value: bson.A{
									"$stock",
									"$producto.costo_unitario",
								}},
							},
						},
					},
				},
			}},
		},

		// 8. Proyección.
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "deposito_id", Value: "$_id.deposito_id"},
				{Key: "deposito_nombre", Value: "$_id.deposito_nombre"},
				{Key: "stock", Value: 1},
				{Key: "valorizado", Value: 1},
			}},
		},

		// 9. Orden estable.
		bson.D{
			{Key: "$sort", Value: bson.D{
				{Key: "deposito_nombre", Value: 1},
				{Key: "deposito_id", Value: 1},
			}},
		},

		// 10. Página + total.
		bson.D{
			{Key: "$facet", Value: bson.D{
				{
					Key: "items",
					Value: bson.A{
						bson.D{
							{Key: "$skip", Value: skip},
						},
						bson.D{
							{Key: "$limit", Value: int64(tamanioPagina)},
						},
					},
				},
				{
					Key: "total",
					Value: bson.A{
						bson.D{
							{Key: "$count", Value: "cantidad"},
						},
					},
				},
			}},
		},
	)

	cursor, err := r.coleccionDepositoProducto.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoInventarioDeposito `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoInventarioDeposito{}, 0, nil
	}

	resultado := resultados[0]

	var total int64
	if len(resultado.Total) > 0 {
		total = resultado.Total[0].Cantidad
	}

	if resultado.Items == nil {
		resultado.Items = []ResultadoInventarioDeposito{}
	}

	return resultado.Items, total, nil
}

func (r *RepositorioMongo) InventarioPorCategoria(
	ctx context.Context,
	filtro FiltroInventario,
	pagina int,
	tamanioPagina int,
) ([]ResultadoInventarioCategoria, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	match := bson.D{}

	if filtro.DepositoID != nil {
		match = append(match, bson.E{
			Key:   "deposito_id",
			Value: *filtro.DepositoID,
		})
	}

	pipeline := mongo.Pipeline{
		// 1. Filtrar por depósito antes de los joins.
		{
			{Key: "$match", Value: match},
		},

		// 2. Obtener el producto.
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionProducto.Name()},
				{Key: "localField", Value: "producto_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "producto"},
			}},
		},
		{
			{Key: "$unwind", Value: "$producto"},
		},

		// 3. No considerar productos eliminados.
		{
			{Key: "$match", Value: bson.D{
				{Key: "producto.fecha_eliminacion", Value: nil},
			}},
		},
	}

	// 4. Filtrar por categoría si corresponde.
	if filtro.CategoriaID != nil {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "producto.categoria", Value: *filtro.CategoriaID},
			}},
		})
	}

	pipeline = append(pipeline,

		// 5. Obtener la categoría.
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionCategoria.Name()},
				{Key: "localField", Value: "producto.categoria"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "categoria"},
			}},
		},
		bson.D{
			{Key: "$unwind", Value: "$categoria"},
		},

		// 6. No considerar categorías eliminadas.
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "categoria.fecha_eliminacion", Value: nil},
			}},
		},

		// 7. Agrupar por categoría.
		bson.D{
			{Key: "$group", Value: bson.D{
				{
					Key: "_id",
					Value: bson.D{
						{Key: "categoria_id", Value: "$producto.categoria"},
						{Key: "categoria_nombre", Value: "$categoria.nombre"},
					},
				},
				{
					Key: "stock",
					Value: bson.D{
						{Key: "$sum", Value: "$stock"},
					},
				},
				{
					Key: "valorizado",
					Value: bson.D{
						{
							Key: "$sum",
							Value: bson.D{
								{Key: "$multiply", Value: bson.A{
									"$stock",
									"$producto.costo_unitario",
								}},
							},
						},
					},
				},
			}},
		},

		// 8. Proyección.
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "categoria_id", Value: "$_id.categoria_id"},
				{Key: "categoria_nombre", Value: "$_id.categoria_nombre"},
				{Key: "stock", Value: 1},
				{Key: "valorizado", Value: 1},
			}},
		},

		// 9. Orden estable.
		bson.D{
			{Key: "$sort", Value: bson.D{
				{Key: "categoria_nombre", Value: 1},
				{Key: "categoria_id", Value: 1},
			}},
		},

		// 10. Página + total.
		bson.D{
			{Key: "$facet", Value: bson.D{
				{
					Key: "items",
					Value: bson.A{
						bson.D{
							{Key: "$skip", Value: skip},
						},
						bson.D{
							{Key: "$limit", Value: int64(tamanioPagina)},
						},
					},
				},
				{
					Key: "total",
					Value: bson.A{
						bson.D{
							{Key: "$count", Value: "cantidad"},
						},
					},
				},
			}},
		},
	)

	cursor, err := r.coleccionDepositoProducto.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoInventarioCategoria `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoInventarioCategoria{}, 0, nil
	}

	resultado := resultados[0]

	var total int64
	if len(resultado.Total) > 0 {
		total = resultado.Total[0].Cantidad
	}

	if resultado.Items == nil {
		resultado.Items = []ResultadoInventarioCategoria{}
	}

	return resultado.Items, total, nil
}

func (r *RepositorioMongo) Movimientos(
	ctx context.Context,
	filtro FiltroReporteMovimientos,
	pagina int,
	tamanioPagina int,
) ([]ResultadoMovimiento, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	match := bson.D{}

	if filtro.DepositoID != nil {
		match = append(match, bson.E{
			Key:   "deposito_id",
			Value: *filtro.DepositoID,
		})
	}

	if filtro.Tipo != nil {
		match = append(match, bson.E{
			Key:   "tipo",
			Value: *filtro.Tipo,
		})
	}

	if filtro.FechaDesde != nil || filtro.FechaHasta != nil {
		rangoFecha := bson.D{}

		if filtro.FechaDesde != nil {
			rangoFecha = append(rangoFecha, bson.E{
				Key:   "$gte",
				Value: *filtro.FechaDesde,
			})
		}

		if filtro.FechaHasta != nil {
			// FechaHasta llega desde Service como el día siguiente
			// al solicitado, por lo que debe ser límite exclusivo.
			rangoFecha = append(rangoFecha, bson.E{
				Key:   "$lt",
				Value: *filtro.FechaHasta,
			})
		}

		match = append(match, bson.E{
			Key:   "fecha_hora",
			Value: rangoFecha,
		})
	}

	pipeline := mongo.Pipeline{
		// 1. Aplicar filtros de movimientos.
		{
			{Key: "$match", Value: match},
		},

		// 2. Buscar depósito.
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionDeposito.Name()},
				{Key: "localField", Value: "deposito_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "deposito"},
			}},
		},
		{
			{Key: "$unwind", Value: "$deposito"},
		},

		// 3. Ignorar depósitos eliminados.
		{
			{Key: "$match", Value: bson.D{
				{Key: "deposito.fecha_eliminacion", Value: nil},
			}},
		},

		// 4. Obtener fecha en horario argentino.
		{
			{Key: "$set", Value: bson.D{
				{Key: "fecha_reporte", Value: bson.D{
					{Key: "$dateToString", Value: bson.D{
						{Key: "format", Value: "%Y-%m-%d"},
						{Key: "date", Value: "$fecha_hora"},
						{Key: "timezone", Value: "America/Argentina/Buenos_Aires"},
					}},
				}},
			}},
		},

		// 5. Agrupar por depósito, fecha y tipo.
		{
			{Key: "$group", Value: bson.D{
				{
					Key: "_id",
					Value: bson.D{
						{Key: "deposito_id", Value: "$deposito_id"},
						{Key: "deposito_nombre", Value: "$deposito.nombre"},
						{Key: "fecha", Value: "$fecha_reporte"},
						{Key: "tipo", Value: "$tipo"},
					},
				},
				{
					Key: "cantidad_eventos",
					Value: bson.D{
						{Key: "$sum", Value: 1},
					},
				},
				{
					Key: "unidades_netas",
					Value: bson.D{
						{Key: "$sum", Value: "$cantidad"},
					},
				},
			}},
		},

		// 6. Proyección.
		{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "deposito_id", Value: "$_id.deposito_id"},
				{Key: "deposito_nombre", Value: "$_id.deposito_nombre"},
				{Key: "fecha", Value: "$_id.fecha"},
				{Key: "tipo", Value: "$_id.tipo"},
				{Key: "cantidad_eventos", Value: 1},
				{Key: "unidades_netas", Value: 1},
			}},
		},

		// 7. Orden estable.
		{
			{Key: "$sort", Value: bson.D{
				{Key: "fecha", Value: -1},
				{Key: "deposito_nombre", Value: 1},
				{Key: "deposito_id", Value: 1},
				{Key: "tipo", Value: 1},
			}},
		},

		// 8. Página + total.
		{
			{Key: "$facet", Value: bson.D{
				{
					Key: "items",
					Value: bson.A{
						bson.D{
							{Key: "$skip", Value: skip},
						},
						bson.D{
							{Key: "$limit", Value: int64(tamanioPagina)},
						},
					},
				},
				{
					Key: "total",
					Value: bson.A{
						bson.D{
							{Key: "$count", Value: "cantidad"},
						},
					},
				},
			}},
		},
	}

	cursor, err := r.coleccionMovimiento.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoMovimiento `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoMovimiento{}, 0, nil
	}

	total := int64(0)

	if len(resultados[0].Total) > 0 {
		total = resultados[0].Total[0].Cantidad
	}

	items := resultados[0].Items

	if items == nil {
		items = []ResultadoMovimiento{}
	}

	return items, total, nil
}

func (r *RepositorioMongo) AlertasBajoMinimo(
	ctx context.Context,
	filtro FiltroAlertas,
	pagina int,
	tamanioPagina int,
) ([]ResultadoAlertaMinimo, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	pipeline := mongo.Pipeline{}

	// 1. Filtrar por depósito si corresponde.
	if filtro.DepositoID != nil {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "deposito_id", Value: *filtro.DepositoID},
			}},
		})
	}

	// 2. Obtener producto.
	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: r.coleccionProducto.Name()},
			{Key: "localField", Value: "producto_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "producto"},
		}},
	})

	pipeline = append(pipeline, bson.D{
		{Key: "$unwind", Value: "$producto"},
	})

	// 3. Ignorar productos eliminados.
	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "producto.fecha_eliminacion", Value: nil},
		}},
	})

	// 4. Obtener depósito.
	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: r.coleccionDeposito.Name()},
			{Key: "localField", Value: "deposito_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "deposito"},
		}},
	})

	pipeline = append(pipeline, bson.D{
		{Key: "$unwind", Value: "$deposito"},
	})

	// 5. Ignorar depósitos eliminados.
	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "deposito.fecha_eliminacion", Value: nil},
		}},
	})

	// 6. Determinar stock mínimo efectivo.
	pipeline = append(pipeline, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "stock_minimo_efectivo", Value: bson.D{
				{Key: "$ifNull", Value: bson.A{
					"$stock_minimo",
					"$producto.stock_minimo",
				}},
			}},
		}},
	})

	// 7. Buscar productos por debajo del mínimo.
	pipeline = append(pipeline, bson.D{
		{Key: "$match", Value: bson.D{
			{Key: "$expr", Value: bson.D{
				{Key: "$lt", Value: bson.A{
					"$stock",
					"$stock_minimo_efectivo",
				}},
			}},
		}},
	})

	// 8. Proyección.
	pipeline = append(pipeline, bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "deposito_id", Value: "$deposito_id"},
			{Key: "deposito_nombre", Value: "$deposito.nombre"},
			{Key: "producto_id", Value: "$producto_id"},
			{Key: "producto_nombre", Value: "$producto.nombre"},
			{Key: "stock", Value: "$stock"},
			{Key: "stock_minimo", Value: "$stock_minimo_efectivo"},
		}},
	})

	// 9. Orden estable.
	pipeline = append(pipeline, bson.D{
		{Key: "$sort", Value: bson.D{
			{Key: "deposito_nombre", Value: 1},
			{Key: "deposito_id", Value: 1},
			{Key: "producto_nombre", Value: 1},
			{Key: "producto_id", Value: 1},
		}},
	})

	// 10. Página + total.
	pipeline = append(pipeline, bson.D{
		{Key: "$facet", Value: bson.D{
			{
				Key: "items",
				Value: bson.A{
					bson.D{
						{Key: "$skip", Value: skip},
					},
					bson.D{
						{Key: "$limit", Value: int64(tamanioPagina)},
					},
				},
			},
			{
				Key: "total",
				Value: bson.A{
					bson.D{
						{Key: "$count", Value: "cantidad"},
					},
				},
			},
		}},
	})

	cursor, err := r.coleccionDepositoProducto.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoAlertaMinimo `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoAlertaMinimo{}, 0, nil
	}

	total := int64(0)

	if len(resultados[0].Total) > 0 {
		total = resultados[0].Total[0].Cantidad
	}

	items := resultados[0].Items

	if items == nil {
		items = []ResultadoAlertaMinimo{}
	}

	return items, total, nil
}

func (r *RepositorioMongo) AlertasPorVencer(
	ctx context.Context,
	filtro FiltroAlertas,
	pagina int,
	tamanioPagina int,
) ([]ResultadoAlertaVencimiento, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	matchDepositoProducto := bson.D{}

	if filtro.DepositoID != nil {
		matchDepositoProducto = bson.D{
			{Key: "deposito_id", Value: *filtro.DepositoID},
		}
	}

	ahora := time.Now()
	limite := ahora.AddDate(0, 0, filtro.DiasVencimiento)

	pipeline := mongo.Pipeline{
		// 1. Filtrar depósito si corresponde.
		{
			{Key: "$match", Value: matchDepositoProducto},
		},

		// 2. Buscar producto.
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionProducto.Name()},
				{Key: "localField", Value: "producto_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "producto"},
			}},
		},
		{
			{Key: "$unwind", Value: "$producto"},
		},

		// 3. Producto activo.
		{
			{Key: "$match", Value: bson.D{
				{Key: "producto.fecha_eliminacion", Value: nil},
			}},
		},

		// 4. Tiene stock y fecha de vencimiento.
		{
			{Key: "$match", Value: bson.D{
				{Key: "stock", Value: bson.D{
					{Key: "$gt", Value: 0},
				}},
				{Key: "producto.fecha_vencimiento", Value: bson.D{
					{Key: "$ne", Value: nil},
					{Key: "$gte", Value: ahora},
					{Key: "$lte", Value: limite},
				}}}},
		},

		// 5. Buscar depósito.
		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionDeposito.Name()},
				{Key: "localField", Value: "deposito_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "deposito"},
			}},
		},
		{
			{Key: "$unwind", Value: "$deposito"},
		},

		// 6. Depósito activo.
		{
			{Key: "$match", Value: bson.D{
				{Key: "deposito.fecha_eliminacion", Value: nil},
			}},
		},

		// 7. Calcular días para vencer.
		{
			{Key: "$set", Value: bson.D{
				{Key: "dias_para_vencer", Value: bson.D{
					{Key: "$dateDiff", Value: bson.D{
						{Key: "startDate", Value: ahora},
						{Key: "endDate", Value: "$producto.fecha_vencimiento"},
						{Key: "unit", Value: "day"},
					}},
				}},
			}},
		},

		// 8. Proyección.
		{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "deposito_id", Value: "$deposito._id"},
				{Key: "deposito_nombre", Value: "$deposito.nombre"},
				{Key: "producto_id", Value: "$producto._id"},
				{Key: "producto_nombre", Value: "$producto.nombre"},
				{Key: "stock", Value: 1},
				{Key: "fecha_vencimiento", Value: "$producto.fecha_vencimiento"},
				{Key: "dias_para_vencer", Value: 1},
			}},
		},

		// 9. Orden estable.
		{
			{Key: "$sort", Value: bson.D{
				{Key: "fecha_vencimiento", Value: 1},
				{Key: "deposito_nombre", Value: 1},
				{Key: "deposito_id", Value: 1},
				{Key: "producto_nombre", Value: 1},
				{Key: "producto_id", Value: 1},
			}},
		},

		// 10. Página + total.
		{
			{Key: "$facet", Value: bson.D{
				{
					Key: "items",
					Value: bson.A{
						bson.D{
							{Key: "$skip", Value: skip},
						},
						bson.D{
							{Key: "$limit", Value: int64(tamanioPagina)},
						},
					},
				},
				{
					Key: "total",
					Value: bson.A{
						bson.D{
							{Key: "$count", Value: "cantidad"},
						},
					},
				},
			}},
		},
	}

	cursor, err := r.coleccionDepositoProducto.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoAlertaVencimiento `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoAlertaVencimiento{}, 0, nil
	}

	items := resultados[0].Items

	if items == nil {
		items = []ResultadoAlertaVencimiento{}
	}

	var total int64

	if len(resultados[0].Total) > 0 {
		total = resultados[0].Total[0].Cantidad
	}

	return items, total, nil
}

func (r *RepositorioMongo) OrdenesPendientes(
	ctx context.Context,
	filtro FiltroPendientes,
	pagina int,
	tamanioPagina int,
) ([]ResultadoOrdenesDeposito, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	match := bson.D{
		{
			Key:   "fecha_eliminacion",
			Value: nil,
		},
	}

	if filtro.DepositoID != nil {
		match = append(match, bson.E{
			Key:   "deposito_id",
			Value: *filtro.DepositoID,
		})
	}

	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: match},
		},

		{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionDeposito.Name()},
				{Key: "localField", Value: "deposito_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "deposito"},
			}},
		},

		{
			{Key: "$unwind", Value: "$deposito"},
		},

		{
			{Key: "$match", Value: bson.D{
				{Key: "deposito.fecha_eliminacion", Value: nil},
			}},
		},

		{
			{Key: "$group", Value: bson.D{
				{
					Key: "_id",
					Value: bson.D{
						{Key: "deposito_id", Value: "$deposito._id"},
						{Key: "deposito_nombre", Value: "$deposito.nombre"},
					},
				},

				{
					Key: "borrador",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(ordendecompra.OrdenBorrador),
									}},
								},
								1,
								0,
							}},
						}},
					},
				},

				{
					Key: "confirmada",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(ordendecompra.OrdenConfirmada),
									}},
								},
								1,
								0,
							}},
						}},
					},
				},

				{
					Key: "parcial",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(ordendecompra.OrdenParcial),
									}},
								},
								1,
								0,
							}},
						}},
					},
				},

				{
					Key: "completada",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(ordendecompra.OrdenCompletada),
									}},
								},
								1,
								0,
							}},
						}},
					},
				},
			}},
		},

		{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "deposito_id", Value: "$_id.deposito_id"},
				{Key: "deposito_nombre", Value: "$_id.deposito_nombre"},
				{Key: "borrador", Value: 1},
				{Key: "confirmada", Value: 1},
				{Key: "parcial", Value: 1},
				{Key: "completada", Value: 1},
			}},
		},

		{
			{Key: "$sort", Value: bson.D{
				{Key: "deposito_nombre", Value: 1},
				{Key: "deposito_id", Value: 1},
			}},
		},

		{
			{Key: "$facet", Value: bson.D{
				{
					Key: "items",
					Value: bson.A{
						bson.D{
							{Key: "$skip", Value: skip},
						},
						bson.D{
							{Key: "$limit", Value: int64(tamanioPagina)},
						},
					},
				},
				{
					Key: "total",
					Value: bson.A{
						bson.D{
							{Key: "$count", Value: "cantidad"},
						},
					},
				},
			}},
		},
	}

	cursor, err := r.coleccionOrdenCompra.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoOrdenesDeposito `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoOrdenesDeposito{}, 0, nil
	}

	items := resultados[0].Items

	if items == nil {
		items = []ResultadoOrdenesDeposito{}
	}

	var total int64

	if len(resultados[0].Total) > 0 {
		total = resultados[0].Total[0].Cantidad
	}

	return items, total, nil
}

func (r *RepositorioMongo) TransferenciasPendientes(
	ctx context.Context,
	filtro FiltroPendientes,
	pagina int,
	tamanioPagina int,
) ([]ResultadoTransferenciasDeposito, int64, error) {

	skip := int64((pagina - 1) * tamanioPagina)

	pipeline := mongo.Pipeline{
		{
			{Key: "$match", Value: bson.D{
				{Key: "estado", Value: bson.D{
					{Key: "$in", Value: bson.A{
						string(transferencia.TransferenciaSolicitada),
						string(transferencia.TransferenciaAprobada),
					}},
				}},
			}},
		},

		{
			{Key: "$project", Value: bson.D{
				{Key: "participaciones", Value: bson.A{
					bson.D{
						{Key: "deposito_id", Value: "$deposito_origen_id"},
						{Key: "tipo", Value: "solicitada_origen"},
						{Key: "cantidad", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(transferencia.TransferenciaSolicitada),
									}},
								},
								1,
								0,
							}},
						}},
					},

					bson.D{
						{Key: "deposito_id", Value: "$deposito_destino_id"},
						{Key: "tipo", Value: "pendiente_aprobacion"},
						{Key: "cantidad", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(transferencia.TransferenciaSolicitada),
									}},
								},
								1,
								0,
							}},
						}},
					},

					bson.D{
						{Key: "deposito_id", Value: "$deposito_destino_id"},
						{Key: "tipo", Value: "aprobada_recibir"},
						{Key: "cantidad", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$estado",
										string(transferencia.TransferenciaAprobada),
									}},
								},
								1,
								0,
							}},
						}},
					},
				}},
			}},
		},

		{
			{Key: "$unwind", Value: "$participaciones"},
		},

		{
			{Key: "$match", Value: bson.D{
				{Key: "participaciones.cantidad", Value: bson.D{
					{Key: "$gt", Value: 0},
				}},
			}},
		},
	}

	if filtro.DepositoID != nil {
		pipeline = append(pipeline, bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "participaciones.deposito_id", Value: *filtro.DepositoID},
			}},
		})
	}

	pipeline = append(pipeline,

		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: r.coleccionDeposito.Name()},
				{Key: "localField", Value: "participaciones.deposito_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "deposito"},
			}},
		},

		bson.D{
			{Key: "$unwind", Value: "$deposito"},
		},

		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "deposito.fecha_eliminacion", Value: nil},
			}},
		},

		bson.D{
			{Key: "$group", Value: bson.D{
				{
					Key: "_id",
					Value: bson.D{
						{Key: "deposito_id", Value: "$deposito._id"},
						{Key: "deposito_nombre", Value: "$deposito.nombre"},
					},
				},

				{
					Key: "solicitadas_como_origen",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$participaciones.tipo",
										"solicitada_origen",
									}},
								},
								1,
								0,
							}},
						}},
					},
				},

				{
					Key: "pendientes_de_aprobar",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$participaciones.tipo",
										"pendiente_aprobacion",
									}},
								},
								1,
								0,
							}},
						}},
					},
				},

				{
					Key: "aprobadas_por_recibir",
					Value: bson.D{
						{Key: "$sum", Value: bson.D{
							{Key: "$cond", Value: bson.A{
								bson.D{
									{Key: "$eq", Value: bson.A{
										"$participaciones.tipo",
										"aprobada_recibir",
									}},
								},
								1,
								0,
							}},
						}},
					},
				},
			}},
		},

		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "deposito_id", Value: "$_id.deposito_id"},
				{Key: "deposito_nombre", Value: "$_id.deposito_nombre"},
				{Key: "solicitadas_como_origen", Value: 1},
				{Key: "pendientes_de_aprobar", Value: 1},
				{Key: "aprobadas_por_recibir", Value: 1},
			}},
		},

		bson.D{
			{Key: "$sort", Value: bson.D{
				{Key: "deposito_nombre", Value: 1},
				{Key: "deposito_id", Value: 1},
			}},
		},

		bson.D{
			{Key: "$facet", Value: bson.D{
				{
					Key: "items",
					Value: bson.A{
						bson.D{
							{Key: "$skip", Value: skip},
						},
						bson.D{
							{Key: "$limit", Value: int64(tamanioPagina)},
						},
					},
				},
				{
					Key: "total",
					Value: bson.A{
						bson.D{
							{Key: "$count", Value: "cantidad"},
						},
					},
				},
			}},
		},
	)

	cursor, err := r.coleccionTransferencia.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultados []struct {
		Items []ResultadoTransferenciasDeposito `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultados); err != nil {
		return nil, 0, err
	}

	if len(resultados) == 0 {
		return []ResultadoTransferenciasDeposito{}, 0, nil
	}

	items := resultados[0].Items

	if items == nil {
		items = []ResultadoTransferenciasDeposito{}
	}

	var total int64

	if len(resultados[0].Total) > 0 {
		total = resultados[0].Total[0].Cantidad
	}

	return items, total, nil
}

func (r *RepositorioMongo) TotalesInventario(
	ctx context.Context,
	filtro FiltroInventario,
) (int32, float64, error) {

	matchDepositoProducto := bson.D{}

	if filtro.DepositoID != nil {
		matchDepositoProducto = bson.D{
			{
				Key:   "deposito_id",
				Value: *filtro.DepositoID,
			},
		}
	}

	pipeline := mongo.Pipeline{
		// 1. Filtrar por depósito si corresponde.
		{
			{
				Key:   "$match",
				Value: matchDepositoProducto,
			},
		},

		// 2. Buscar producto activo.
		{
			{
				Key: "$lookup",
				Value: bson.D{
					{
						Key:   "from",
						Value: r.coleccionProducto.Name(),
					},
					{
						Key:   "localField",
						Value: "producto_id",
					},
					{
						Key:   "foreignField",
						Value: "_id",
					},
					{
						Key:   "as",
						Value: "producto",
					},
					{
						Key: "pipeline",
						Value: mongo.Pipeline{
							{
								{
									Key: "$match",
									Value: bson.D{
										{
											Key:   "fecha_eliminacion",
											Value: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		},

		// 3. Descartar productos inexistentes/eliminados.
		{
			{
				Key:   "$unwind",
				Value: "$producto",
			},
		},
	}

	// 4. Filtrar por categoría si corresponde.
	if filtro.CategoriaID != nil {
		pipeline = append(
			pipeline,
			bson.D{
				{
					Key: "$match",
					Value: bson.D{
						{
							Key:   "producto.categoria",
							Value: *filtro.CategoriaID,
						},
					},
				},
			},
		)
	}

	pipeline = append(
		pipeline,

		// 5. Buscar depósito activo.
		bson.D{
			{
				Key: "$lookup",
				Value: bson.D{
					{
						Key:   "from",
						Value: r.coleccionDeposito.Name(),
					},
					{
						Key:   "localField",
						Value: "deposito_id",
					},
					{
						Key:   "foreignField",
						Value: "_id",
					},
					{
						Key:   "as",
						Value: "deposito",
					},
					{
						Key: "pipeline",
						Value: mongo.Pipeline{
							{
								{
									Key: "$match",
									Value: bson.D{
										{
											Key:   "fecha_eliminacion",
											Value: nil,
										},
									},
								},
							},
						},
					},
				},
			},
		},

		// 6. Descartar depósitos inexistentes/eliminados.
		bson.D{
			{
				Key:   "$unwind",
				Value: "$deposito",
			},
		},

		// 7. Calcular el valorizado de cada registro.
		bson.D{
			{
				Key: "$project",
				Value: bson.D{
					{
						Key:   "stock",
						Value: "$stock",
					},
					{
						Key: "valorizado",
						Value: bson.D{
							{
								Key: "$multiply",
								Value: bson.A{
									"$stock",
									"$producto.costo_unitario",
								},
							},
						},
					},
				},
			},
		},

		// 8. Obtener los totales globales.
		bson.D{
			{
				Key: "$group",
				Value: bson.D{
					{
						Key:   "_id",
						Value: nil,
					},
					{
						Key: "stock_total",
						Value: bson.D{
							{
								Key:   "$sum",
								Value: "$stock",
							},
						},
					},
					{
						Key: "valorizado_total",
						Value: bson.D{
							{
								Key:   "$sum",
								Value: "$valorizado",
							},
						},
					},
				},
			},
		},
	)

	cursor, err := r.coleccionDepositoProducto.Aggregate(
		ctx,
		pipeline,
	)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	var resultado struct {
		StockTotal      int32   `bson:"stock_total"`
		ValorizadoTotal float64 `bson:"valorizado_total"`
	}

	if cursor.Next(ctx) {
		if err := cursor.Decode(&resultado); err != nil {
			return 0, 0, err
		}

		return resultado.StockTotal, resultado.ValorizadoTotal, nil
	}

	if err := cursor.Err(); err != nil {
		return 0, 0, err
	}

	return 0, 0, nil
}
