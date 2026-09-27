package depositoproducto

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ErrStockInsuficiente se devuelve cuando AjustarStock intentaría dejar el
// stock de un producto en un depósito por debajo de cero. Es un error de
// paquete (no uno genérico de Mongo) para que el service pueda distinguirlo
// y responder 400 en vez de 500 (ver movimientoStock.Service, que es quien
// más va a chequear este error puntual).
var ErrStockInsuficiente = errors.New("el stock disponible es insuficiente para esta operación")

// DepositoProductoRepositorio persiste el stock de cada producto en cada
// depósito. No tiene un ABM clásico (crear/editar/eliminar) porque la fila no
// la crea un usuario a mano: nace sola la primera vez que hace falta (ver
// AsegurarExiste) y a partir de ahí solo se ajusta con $inc (ver
// AjustarStock) — nunca con un UpdateOne que pise el valor entero, así dos
// movimientos concurrentes sobre el mismo producto no se pisan entre sí.
type DepositoProductoRepositorio interface {
	EncontrarPorDepositoYProducto(ctx context.Context, depositoID, productoID bson.ObjectID) (DepositoProducto, error)
	EncontrarPorDeposito(ctx context.Context, depositoID bson.ObjectID) ([]DepositoProducto, error)
	EncontrarPorProducto(ctx context.Context, productoID bson.ObjectID) ([]DepositoProducto, error)
	EncontrarTodos(ctx context.Context) ([]DepositoProducto, error)
	AsegurarExiste(ctx context.Context, depositoID, productoID, usuarioID bson.ObjectID) (DepositoProducto, error)
	AjustarStock(ctx context.Context, depositoID, productoID bson.ObjectID, delta int32, usuarioID bson.ObjectID) (DepositoProducto, error)
	ConfigurarStockMinimo(ctx context.Context, depositoID, productoID bson.ObjectID, stockMinimo *int32, usuarioID bson.ObjectID) (DepositoProducto, error)
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
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
// particular del depósito. También hace upsert: un gerente tiene que poder
// configurar el mínimo de un producto en su depósito ANTES de que exista
// cualquier movimiento de stock para ese producto.
func (r *RepositorioMongo) ConfigurarStockMinimo(ctx context.Context, depositoID, productoID bson.ObjectID, stockMinimo *int32, usuarioID bson.ObjectID) (DepositoProducto, error) {
	filtro := bson.M{"deposito_id": depositoID, "producto_id": productoID}
	ahora := time.Now()

	set := bson.M{
		"auditoria.modificado_por": usuarioID,
		"auditoria.actualizado_en": ahora,
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"deposito_id":          depositoID,
			"producto_id":          productoID,
			"stock":                int32(0),
			"auditoria.creado_por": usuarioID,
			"auditoria.creado_en":  ahora,
		},
	}

	if stockMinimo == nil {
		update["$unset"] = bson.M{"stock_minimo": ""}
		update["$set"] = set
	} else {
		set["stock_minimo"] = *stockMinimo
		update["$set"] = set
	}

	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var dp DepositoProducto
	if err := r.coleccion.FindOneAndUpdate(ctx, filtro, update, opts).Decode(&dp); err != nil {
		return DepositoProducto{}, err
	}
	return dp, nil
}
