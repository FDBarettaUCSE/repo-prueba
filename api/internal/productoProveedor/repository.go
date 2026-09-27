package productoproveedor

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ProductoProveedorRepositorio persiste el vínculo N:M entre producto y
// proveedor. No tiene ID propio: la clave lógica es el par (ProductoID,
// ProveedorID) — no puede haber dos vínculos distintos para el mismo par.
type ProductoProveedorRepositorio interface {
	EncontrarUno(ctx context.Context, productoID, proveedorID bson.ObjectID) (ProductoProveedor, error)
	EncontrarPorProveedor(ctx context.Context, proveedorID bson.ObjectID) ([]ProductoProveedor, error)
	EncontrarPorProducto(ctx context.Context, productoID bson.ObjectID) ([]ProductoProveedor, error)
	Guardar(ctx context.Context, pp ProductoProveedor) error
	Eliminar(ctx context.Context, productoID, proveedorID bson.ObjectID) error
	EliminarTodosDeProveedor(ctx context.Context, proveedorID bson.ObjectID) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ ProductoProveedorRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarUno(ctx context.Context, productoID, proveedorID bson.ObjectID) (ProductoProveedor, error) {
	var pp ProductoProveedor

	filtro := bson.M{"producto_id": productoID, "proveedor_id": proveedorID}
	if err := r.coleccion.FindOne(ctx, filtro).Decode(&pp); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ProductoProveedor{}, errors.New("este proveedor no suministra ese producto")
		}
		return ProductoProveedor{}, err
	}
	return pp, nil
}

func (r *RepositorioMongo) EncontrarPorProveedor(ctx context.Context, proveedorID bson.ObjectID) ([]ProductoProveedor, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{"proveedor_id": proveedorID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var resultado []ProductoProveedor
	if err := cursor.All(ctx, &resultado); err != nil {
		return nil, err
	}
	return resultado, nil
}

func (r *RepositorioMongo) EncontrarPorProducto(ctx context.Context, productoID bson.ObjectID) ([]ProductoProveedor, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{"producto_id": productoID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var resultado []ProductoProveedor
	if err := cursor.All(ctx, &resultado); err != nil {
		return nil, err
	}
	return resultado, nil
}

// Guardar hace upsert por (ProductoID, ProveedorID): si el proveedor ya
// suministraba ese producto, actualiza costo/tiempo de entrega; si no,
// crea el vínculo. Así "agregar" y "actualizar" el costo son la misma
// operación desde el repositorio (el service decide con qué mensaje
// responder en cada caso).
func (r *RepositorioMongo) Guardar(ctx context.Context, pp ProductoProveedor) error {
	filtro := bson.M{"producto_id": pp.ProductoID, "proveedor_id": pp.ProveedorID}
	update := bson.M{
		"$set": bson.M{
			"precio_unitario":      pp.CostoUnidad,
			"tiempo_estimado_dias": pp.TiempoEstimadoDias,
			"producto_id":          pp.ProductoID,
			"proveedor_id":         pp.ProveedorID,
		},
	}
	_, err := r.coleccion.UpdateOne(ctx, filtro, update, options.UpdateOne().SetUpsert(true))
	return err
}

func (r *RepositorioMongo) Eliminar(ctx context.Context, productoID, proveedorID bson.ObjectID) error {
	filtro := bson.M{"producto_id": productoID, "proveedor_id": proveedorID}
	resultado, err := r.coleccion.DeleteOne(ctx, filtro)
	if err != nil {
		return err
	}
	if resultado.DeletedCount == 0 {
		return errors.New("este proveedor no suministra ese producto")
	}
	return nil
}

// EliminarTodosDeProveedor se usa al eliminar (soft-delete) un proveedor
// entero: no tendría sentido dejar vínculos sueltos a un proveedor dado de baja.
func (r *RepositorioMongo) EliminarTodosDeProveedor(ctx context.Context, proveedorID bson.ObjectID) error {
	_, err := r.coleccion.DeleteMany(ctx, bson.M{"proveedor_id": proveedorID})
	return err
}
