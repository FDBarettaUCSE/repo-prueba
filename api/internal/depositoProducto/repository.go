package depositoproducto

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ErrStockInsuficiente se devuelve cuando AjustarStock intentaría dejar el
// stock de un producto en un depósito por debajo de cero. Es un error de
// paquete (no uno genérico de Mongo) para que el service pueda distinguirlo
// y responder 409 en vez de 500 (ver movimientoStock.Service, que es quien
// más va a chequear este error puntual).
var ErrStockInsuficiente = errors.New("el stock disponible es insuficiente para esta operación")

// FilaStockDeposito es el resultado de la agregación paginada de stock por
// depósito en MongoDB.
type FilaStockDeposito struct {
	ProductoID          bson.ObjectID `bson:"producto_id"`
	ProductoNombre      string        `bson:"producto_nombre"`
	CategoriaID         bson.ObjectID `bson:"categoria_id"`
	CategoriaNombre     string        `bson:"categoria_nombre"`
	UnidadMedida        string        `bson:"unidad_medida"`
	CostoUnitario       float64       `bson:"costo_unitario"`
	Stock               int32         `bson:"stock"`
	StockMinimo         int32         `bson:"stock_minimo"`
	StockMinimoOverride *int32        `bson:"stock_minimo_override,omitempty"`
	BajoMinimo          bool          `bson:"bajo_minimo"`
}

// DepositoProductoRepositorio persiste el stock de cada producto en cada
// depósito. No tiene un ABM clásico (crear/editar/eliminar) porque la fila no
// la crea un usuario a mano: nace sola la primera vez que hace falta (ver
// AsegurarExiste) y a partir de ahí solo se ajusta con $inc (ver
// AjustarStock) — nunca con un UpdateOne que pise el valor entero, así dos
// movimientos concurrentes sobre el mismo producto no se pisan entre sí.
type DepositoProductoRepositorio interface {
	EncontrarPorDepositoYProducto(ctx context.Context, depositoID, productoID bson.ObjectID) (DepositoProducto, error)
	//Este método es general y se usa más que nada para otros módulos. No filtra.
	EncontrarPorDeposito(ctx context.Context, depositoID bson.ObjectID) ([]DepositoProducto, error)
	EncontrarPorProducto(ctx context.Context, productoID bson.ObjectID) ([]DepositoProducto, error)
	EncontrarTodos(ctx context.Context) ([]DepositoProducto, error)
	AsegurarExiste(ctx context.Context, depositoID, productoID, usuarioID bson.ObjectID) (DepositoProducto, error)
	AjustarStock(ctx context.Context, depositoID, productoID bson.ObjectID, delta int32, usuarioID bson.ObjectID) (DepositoProducto, error)
	ConfigurarStockMinimo(ctx context.Context, depositoID, productoID bson.ObjectID, stockMinimo *int32, usuarioID bson.ObjectID) (DepositoProducto, error)
	//A diferencia de EncontrarPorDeposito, este metodo si pagina y filtra.
	ListarStockDeposito(ctx context.Context, epositoID bson.ObjectID, filtro ListarStockFiltroRepositorio) ([]StockDepositoFila, int64, error)
}

type RepositorioMongo struct {
	coleccion          *mongo.Collection
	coleccionProducto  *mongo.Collection
	coleccionCategoria *mongo.Collection
}

func NuevoRepositorioMongo(
	coleccion *mongo.Collection,
	coleccionProdutos ...*mongo.Collection,
) *RepositorioMongo {
	r := &RepositorioMongo{coleccion: coleccion}
	if len(coleccionProdutos) > 0 {
		r.coleccionProducto = coleccionProdutos[0]
	}
	if len(coleccionProdutos) > 1 {
		r.coleccionCategoria = coleccionProdutos[1]
	}
	return r
}

var _ DepositoProductoRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorDepositoYProducto(ctx context.Context, depositoID, productoID bson.ObjectID) (DepositoProducto, error) {
	var dp DepositoProducto

	filtro := bson.M{"deposito_id": depositoID, "producto_id": productoID}
	if err := r.coleccion.FindOne(ctx, filtro).Decode(&dp); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// No es un error real: significa que todavía no hubo ningún
			// movimiento de este producto en este depósito. El caller decide
			// si eso es "stock 0" (lectura) o si corresponde AsegurarExiste
			// (escritura).
			return DepositoProducto{}, mongo.ErrNoDocuments
		}
		return DepositoProducto{}, err
	}
	return dp, nil
}

func (r *RepositorioMongo) EncontrarPorDeposito(ctx context.Context, depositoID bson.ObjectID) ([]DepositoProducto, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{"deposito_id": depositoID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var resultado []DepositoProducto
	if err := cursor.All(ctx, &resultado); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *RepositorioMongo) EncontrarPorProducto(ctx context.Context, productoID bson.ObjectID) ([]DepositoProducto, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{"producto_id": productoID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var resultado []DepositoProducto
	if err := cursor.All(ctx, &resultado); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *RepositorioMongo) EncontrarTodos(ctx context.Context) ([]DepositoProducto, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var resultado []DepositoProducto
	if err := cursor.All(ctx, &resultado); err != nil {
		return nil, err
	}
	return resultado, nil
}

// AsegurarExiste hace upsert por (DepositoID, ProductoID): si la fila ya
// existía la devuelve tal cual, si no, la crea con stock 0. La usa
// movimientoStock ANTES de todo $inc, para no pisar con $setOnInsert un
// stock que otro movimiento concurrente ya haya modificado.
func (r *RepositorioMongo) AsegurarExiste(ctx context.Context, depositoID, productoID, usuarioID bson.ObjectID) (DepositoProducto, error) {
	filtro := bson.M{"deposito_id": depositoID, "producto_id": productoID}
	ahora := time.Now()

	update := bson.M{
		"$setOnInsert": bson.M{
			"deposito_id":          depositoID,
			"producto_id":          productoID,
			"stock":                int32(0),
			"auditoria.creado_por": usuarioID,
			"auditoria.creado_en":  ahora,
		},
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var dp DepositoProducto
	if err := r.coleccion.FindOneAndUpdate(ctx, filtro, update, opts).Decode(&dp); err != nil {
		return DepositoProducto{}, err
	}
	return dp, nil
}

// AjustarStock suma delta (puede ser negativo) al stock actual de forma
// atómica. Si delta es negativo, el filtro exige "stock >= -delta" — así el
// $inc y el chequeo de "no negativo" ocurren en la MISMA operación de Mongo,
// sin la ventana de carrera de un "leer, validar en Go, y recién ahí
// escribir" (dos egresos simultáneos podrían leer el mismo stock "antes" y
// los dos creerse habilitados). Si no matchea ninguna fila, quiere decir que
// el stock disponible era menor al que se intentó descontar.
func (r *RepositorioMongo) AjustarStock(ctx context.Context, depositoID, productoID bson.ObjectID, delta int32, usuarioID bson.ObjectID) (DepositoProducto, error) {
	filtro := bson.M{"deposito_id": depositoID, "producto_id": productoID}
	if delta < 0 {
		filtro["stock"] = bson.M{"$gte": -delta}
	}

	ahora := time.Now()
	update := bson.M{
		"$inc": bson.M{"stock": delta},
		"$set": bson.M{
			"auditoria.modificado_por": usuarioID,
			"auditoria.actualizado_en": ahora,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var dp DepositoProducto
	err := r.coleccion.FindOneAndUpdate(ctx, filtro, update, opts).Decode(&dp)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			if delta < 0 {
				return DepositoProducto{}, ErrStockInsuficiente
			}
			return DepositoProducto{}, errors.New("no existe stock registrado para ese producto en ese depósito")
		}
		return DepositoProducto{}, err
	}
	return dp, nil
}

// ConfigurarStockMinimo fija (o, con stockMinimo nil, quita) el override
// particular del depósito. También hace upsert: el administrador tiene que
// poder configurar el mínimo de un producto en un depósito ANTES de que
// exista cualquier movimiento de stock para ese producto.
func (r *RepositorioMongo) ConfigurarStockMinimo(ctx context.Context, depositoID, productoID bson.ObjectID, stockMinimo *int32, usuarioID bson.ObjectID) (DepositoProducto, error) {
	filtro := bson.M{"deposito_id": depositoID, "producto_id": productoID}
	ahora := time.Now()

	update := bson.M{
		"$setOnInsert": bson.M{
			"deposito_id":          depositoID,
			"producto_id":          productoID,
			"stock":                int32(0),
			"auditoria.creado_por": usuarioID,
			"auditoria.creado_en":  ahora,
		},
		"$set": bson.M{
			"auditoria.modificado_por": usuarioID,
			"auditoria.actualizado_en": ahora,
		},
	}

	if stockMinimo == nil {
		update["$unset"] = bson.M{
			"stock_minimo": "",
		}
	} else {
		update["$set"].(bson.M)["stock_minimo"] = *stockMinimo
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var dp DepositoProducto
	if err := r.coleccion.FindOneAndUpdate(ctx, filtro, update, opts).Decode(&dp); err != nil {
		return DepositoProducto{}, err
	}
	return dp, nil
}

func (r *RepositorioMongo) ListarStockDeposito(ctx context.Context, depositoID bson.ObjectID, filtro ListarStockFiltroRepositorio) ([]StockDepositoFila, int64, error) {
	pagina := filtro.Pagina
	tamanio := filtro.TamanioPagina

	coleccionProd := r.coleccionProducto
	if coleccionProd == nil {
		coleccionProd = r.coleccion.Database().Collection("productos")
	}

	// 1. Partir de todos los productos activos del catálogo.
	matchProductos := bson.D{
		{Key: "fecha_eliminacion", Value: nil},
	}

	if filtro.CategoriaID != "" {
		categoriaID, err := bson.ObjectIDFromHex(filtro.CategoriaID)
		if err != nil {
			return nil, 0, fmt.Errorf("categoría inválida: %w", err)
		}
		matchProductos = append(matchProductos, bson.E{Key: "categoria", Value: categoriaID})
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: matchProductos}},

		// 2. Buscar el registro de stock para este depósito (si existe).
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "deposito_producto"},
			{Key: "localField", Value: "_id"},
			{Key: "foreignField", Value: "producto_id"},
			{Key: "as", Value: "dp"},
			{Key: "pipeline", Value: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "deposito_id", Value: depositoID},
				}}},
			}},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$dp"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},

		// 3. Buscar la categoría del producto.
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "categorias"},
			{Key: "localField", Value: "categoria"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "cat"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$cat"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}}},

		// 4. Preparar los datos y calcular el stock (0 si no hay fila) y el mínimo efectivo.
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "producto_id", Value: "$_id"},
			{Key: "producto_nombre", Value: "$nombre"},
			{Key: "categoria_id", Value: "$categoria"},
			{Key: "categoria_nombre", Value: bson.D{{
				Key:   "$ifNull",
				Value: bson.A{"$cat.nombre", ""},
			}}},
			{Key: "stock", Value: bson.D{{
				Key:   "$ifNull",
				Value: bson.A{"$dp.stock", int32(0)},
			}}},
			{Key: "stock_minimo_override", Value: "$dp.stock_minimo"},
			{Key: "stock_minimo_efectivo", Value: bson.D{{
				Key:   "$ifNull",
				Value: bson.A{"$dp.stock_minimo", "$stock_minimo"},
			}}},
		}}},

		// 5. Calcular la valorización y si está bajo el mínimo.
		bson.D{{Key: "$set", Value: bson.D{
			{Key: "valorizado", Value: bson.D{{
				Key:   "$multiply",
				Value: bson.A{"$stock", "$costo_unitario"},
			}}},
			{Key: "bajo_minimo", Value: bson.D{{
				Key:   "$lt",
				Value: bson.A{"$stock", "$stock_minimo_efectivo"},
			}}},
		}}},
	}

	if filtro.SoloBajoMinimo {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: bson.D{
			{Key: "bajo_minimo", Value: true},
		}}})
	}

	// 6. Obtener el total y la página en una misma consulta.
	pipeline = append(pipeline, bson.D{{Key: "$facet", Value: bson.D{
		{Key: "items", Value: bson.A{
			bson.D{{Key: "$sort", Value: bson.D{
				{Key: "producto_nombre", Value: 1},
			}}},
			bson.D{{Key: "$skip", Value: int64((pagina - 1) * tamanio)}},
			bson.D{{Key: "$limit", Value: int64(tamanio)}},
			bson.D{{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "producto_id", Value: 1},
				{Key: "producto_nombre", Value: 1},
				{Key: "categoria_id", Value: 1},
				{Key: "categoria_nombre", Value: 1},
				{Key: "unidad_medida", Value: 1},
				{Key: "stock", Value: 1},
				{Key: "stock_minimo_efectivo", Value: 1},
				{Key: "stock_minimo_override", Value: 1},
				{Key: "costo_unitario", Value: 1},
				{Key: "valorizado", Value: 1},
				{Key: "bajo_minimo", Value: 1},
			}}},
		}},
		{Key: "total", Value: bson.A{
			bson.D{{Key: "$count", Value: "cantidad"}},
		}},
	}}})

	cursor, err := coleccionProd.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var resultado []struct {
		Items []StockDepositoFila `bson:"items"`
		Total []struct {
			Cantidad int64 `bson:"cantidad"`
		} `bson:"total"`
	}

	if err := cursor.All(ctx, &resultado); err != nil {
		return nil, 0, err
	}

	filas := make([]StockDepositoFila, 0)
	var total int64

	if len(resultado) > 0 {
		filas = resultado[0].Items

		if len(resultado[0].Total) > 0 {
			total = resultado[0].Total[0].Cantidad
		}
	}

	return filas, total, nil
}
