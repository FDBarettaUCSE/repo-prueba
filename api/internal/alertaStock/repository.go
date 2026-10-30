package alertastock

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// AlertaStockRepositorio persiste las alertas de stock. No tiene un ABM
// clásico porque las alertas se generan automáticamente al detectar stock
// bajo mínimo (ver Service.EscanearYGenerarAlertas) y solo admiten una
// transición de estado: de pendiente (resuelta=false) a resuelta
// (resuelta=true) — mismo criterio de inmutabilidad que usa
// movimientoStock, donde no se editan ni eliminan registros ya creados.
type AlertaStockRepositorio interface {
	Crear(ctx context.Context, alerta AlertaStock) (AlertaStock, error)
	EncontrarPorID(ctx context.Context, id string) (AlertaStock, error)
	BuscarActiva(ctx context.Context, depositoID, productoID bson.ObjectID, tipo TipoAlerta) (AlertaStock, error)
	Listar(ctx context.Context, depositoID *bson.ObjectID, soloPendientes bool, pagina, tamañoPagina int) ([]AlertaStock, int64, error)
	ContarPendientes(ctx context.Context, depositoID *bson.ObjectID) (int64, error)
	MarcarResuelta(ctx context.Context, id string, usuarioID bson.ObjectID) (AlertaStock, error)
	ActualizarStock(ctx context.Context, id string, tockActual, stockMinimo int32, usuarioID bson.ObjectID) error
	ActualizarVencimiento(ctx context.Context, id string, fechaVencimiento time.Time, stockActual int32, usuarioID bson.ObjectID) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ AlertaStockRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) Crear(ctx context.Context, alerta AlertaStock) (AlertaStock, error) {
	resultado, err := r.coleccion.InsertOne(ctx, alerta)
	if err != nil {
		return AlertaStock{}, err
	}
	alerta.ID = resultado.InsertedID.(bson.ObjectID)
	return alerta, nil
}

// EncontrarPorID recibe el id como string (hex) — mismo criterio que
// deposito.DepositoRepositorio, producto.ProductoRepositorio y
// movimientoStock.MovimientoStockRepositorio. La conversión a ObjectID
// queda encapsulada en el repositorio.
func (r *RepositorioMongo) EncontrarPorID(ctx context.Context, id string) (AlertaStock, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return AlertaStock{}, err
	}

	var alerta AlertaStock
	err = r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&alerta)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return AlertaStock{}, errors.New("alerta de stock no encontrada")
		}
		return AlertaStock{}, err
	}
	return alerta, nil
}

// BuscarActiva busca si ya existe una alerta NO resuelta para el mismo
// depósito y producto. Se usa antes de crear una nueva alerta para evitar
// duplicados — si ya hay una pendiente para ese par (depósito, producto),
// no tiene sentido generar otra.
func (r *RepositorioMongo) BuscarActiva(ctx context.Context, depositoID, productoID bson.ObjectID, tipo TipoAlerta) (AlertaStock, error) {

	var alerta AlertaStock

	filtro := bson.M{
		"deposito_id": depositoID,
		"producto_id": productoID,
		"tipo":        tipo,
		"resuelta":    false,
	}

	err := r.coleccion.FindOne(ctx, filtro).Decode(&alerta)
	if err != nil {
		return AlertaStock{}, err
	}

	return alerta, nil
}

func (r *RepositorioMongo) Listar(
	ctx context.Context,
	depositoID *bson.ObjectID,
	soloPendientes bool,
	pagina, tamañoPagina int,
) ([]AlertaStock, int64, error) {

	consulta := bson.M{}

	if depositoID != nil {
		consulta["deposito_id"] = *depositoID
	}

	if soloPendientes {
		consulta["resuelta"] = false
	}

	total, err := r.coleccion.CountDocuments(ctx, consulta)
	if err != nil {
		return nil, 0, err
	}

	if pagina < 1 {
		pagina = 1
	}
	if tamañoPagina < 1 {
		tamañoPagina = 20
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: consulta}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "productos",
			"localField":   "producto_id",
			"foreignField": "_id",
			"as":           "producto",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$producto",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "depositos",
			"localField":   "deposito_id",
			"foreignField": "_id",
			"as":           "deposito",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$deposito",
			"preserveNullAndEmptyArrays": true,
		}}},
		{{Key: "$set", Value: bson.M{
			"producto_nombre": bson.M{"$ifNull": bson.A{"$producto.nombre", ""}},
			"deposito_nombre": bson.M{"$ifNull": bson.A{"$deposito.nombre", ""}},
		}}},
		{{Key: "$sort", Value: bson.D{
			{Key: "auditoria.creado_en", Value: -1},
		}}},
		{{Key: "$skip", Value: int64((pagina - 1) * tamañoPagina)}},
		{{Key: "$limit", Value: int64(tamañoPagina)}},
	}

	cursor, err := r.coleccion.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	alertas := make([]AlertaStock, 0)
	if err := cursor.All(ctx, &alertas); err != nil {
		return nil, 0, err
	}

	return alertas, total, nil
}

func (r *RepositorioMongo) ContarPendientes(ctx context.Context, depositoID *bson.ObjectID) (int64, error) {
	consulta := bson.M{"resuelta": false}
	if depositoID != nil {
		consulta["deposito_id"] = *depositoID
	}
	return r.coleccion.CountDocuments(ctx, consulta)
}

// MarcarResuelta recibe el id como string (hex) y actualiza resuelta=true
// junto con los campos de auditoría (quién la resolvió y cuándo) en la MISMA
// operación atómica — mismo patrón que novedadDeposito.RepositorioMongo.
// MarcarLeida.
func (r *RepositorioMongo) MarcarResuelta(ctx context.Context, id string, usuarioID bson.ObjectID) (AlertaStock, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return AlertaStock{}, err
	}

	ahora := time.Now()
	update := bson.M{
		"$set": bson.M{
			"resuelta":                 true,
			"auditoria.modificado_por": usuarioID,
			"auditoria.actualizado_en": ahora,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var alerta AlertaStock
	err = r.coleccion.FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&alerta)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return AlertaStock{}, errors.New("alerta no encontrada")
		}
		return AlertaStock{}, err
	}
	return alerta, nil
}

func (r *RepositorioMongo) ActualizarStock(ctx context.Context, id string, stockActual, stockMinimo int32, usuarioID bson.ObjectID) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{"$set": bson.M{
		"stock_actual":             stockActual,
		"stock_minimo":             stockMinimo,
		"auditoria.modificado_por": usuarioID,
		"auditoria.actualizado_en": time.Now(),
	}}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid, "resuelta": false},
		update,
	)
	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("alerta activa no encontrada")
	}

	return nil
}

func (r *RepositorioMongo) ActualizarVencimiento(
	ctx context.Context,
	id string,
	fechaVencimiento time.Time,
	stockActual int32,
	usuarioID bson.ObjectID,
) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	ahora := time.Now()

	update := bson.M{"$set": bson.M{
		"fecha_vencimiento":        fechaVencimiento,
		"stock_actual":             stockActual,
		"auditoria.modificado_por": usuarioID,
		"auditoria.actualizado_en": ahora,
	}}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid, "resuelta": false},
		update,
	)
	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("alerta activa de vencimiento no encontrada")
	}

	return nil
}
