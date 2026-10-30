package movimientostock

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MovimientoStockRepositorio interface {
	Crear(ctx context.Context, m MovimientoStock) (MovimientoStock, error)
	Listar(ctx context.Context, filtro FiltroMovimientosDTO, depositoPermitido string) ([]MovimientoStock, int64, error)
	EncontrarPorID(ctx context.Context, id string) (MovimientoStock, error)
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(col *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: col}
}

var _ MovimientoStockRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) Crear(ctx context.Context, m MovimientoStock) (MovimientoStock, error) {
	res, err := r.coleccion.InsertOne(ctx, m)
	if err != nil {
		return MovimientoStock{}, err
	}
	m.ID = res.InsertedID.(bson.ObjectID)
	return m, nil
}

func (r *RepositorioMongo) EncontrarPorID(ctx context.Context, id string) (MovimientoStock, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return MovimientoStock{}, ErrIDMovimientoInvalido
	}

	var m MovimientoStock
	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&m); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return MovimientoStock{}, ErrMovimientoInexistente
		}
		return MovimientoStock{}, err
	}
	return m, nil
}

func (r *RepositorioMongo) Listar(ctx context.Context, filtro FiltroMovimientosDTO, depositoPermitido string) ([]MovimientoStock, int64, error) {
	query := bson.M{}

	if depositoPermitido != "" {
		if dOID, err := bson.ObjectIDFromHex(depositoPermitido); err == nil {
			query["deposito_id"] = dOID
		}
	} else if filtro.DepositoID != "" {
		if dOID, err := bson.ObjectIDFromHex(filtro.DepositoID); err == nil {
			query["deposito_id"] = dOID
		}
	}

	if filtro.ProductoID != "" {
		if pOID, err := bson.ObjectIDFromHex(filtro.ProductoID); err == nil {
			query["producto_id"] = pOID
		}
	}

	if filtro.Tipo != "" {
		query["tipo"] = filtro.Tipo
	}

	if filtro.FechaDesde != "" || filtro.FechaHasta != "" {
		filtroFecha := bson.M{}
		if filtro.FechaDesde != "" {
			if tDesde, err := time.Parse("2006-01-02", filtro.FechaDesde); err == nil {
				filtroFecha["$gte"] = tDesde
			}
		}
		if filtro.FechaHasta != "" {
			if tHasta, err := time.Parse("2006-01-02", filtro.FechaHasta); err == nil {
				filtroFecha["$lte"] = tHasta.Add(24*time.Hour - time.Nanosecond)
			}
		}
		if len(filtroFecha) > 0 {
			query["fecha_hora"] = filtroFecha
		}
	}

	total, err := r.coleccion.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Orden descendente por fecha_hora para ver los más recientes primero con paginación
	opts := options.Find().SetSort(bson.D{{Key: "fecha_hora", Value: -1}})

	if filtro.TamanioPagina >= 0 {
		pagina := filtro.Pagina
		if pagina < 1 {
			pagina = 1
		}
		tamanioPagina := filtro.TamanioPagina
		if tamanioPagina < 1 {
			tamanioPagina = 20
		}
		opts.SetSkip(int64((pagina - 1) * tamanioPagina)).
			SetLimit(int64(tamanioPagina))
	}

	cur, err := r.coleccion.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var movimientos []MovimientoStock
	if err := cur.All(ctx, &movimientos); err != nil {
		return nil, 0, err
	}
	return movimientos, total, nil
}
