package productoproveedor

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrProductoSuministradoInexistente = errors.New("este proveedor no suministra ese producto")
	ErrProductoYaSuministrado          = errors.New("el producto ya está asociado a este proveedor")
)

// ProductoProveedorRepositorio persiste el vínculo N:M entre producto y
// proveedor. No tiene ID propio: la clave lógica es el par (ProductoID,
// ProveedorID) — no puede haber dos vínculos distintos para el mismo par.
type ProductoProveedorRepositorio interface {
	EncontrarUno(ctx context.Context, productoID, proveedorID bson.ObjectID) (ProductoProveedor, error)
	EncontrarPorProveedor(ctx context.Context, proveedorID bson.ObjectID) ([]ProductoProveedor, error)
	EncontrarPorProducto(ctx context.Context, productoID bson.ObjectID) ([]ProductoProveedor, error)
	Crear(ctx context.Context, pp ProductoProveedor) error
	Actualizar(ctx context.Context, pp ProductoProveedor) error
	Eliminar(ctx context.Context, productoID, proveedorID, adminID bson.ObjectID) error
	EliminarTodosDeProveedor(ctx context.Context, proveedorID bson.ObjectID, adminID bson.ObjectID) error
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
			return ProductoProveedor{}, ErrProductoSuministradoInexistente
		}
		return ProductoProveedor{}, err
	}
	return pp, nil
}

func (r *RepositorioMongo) EncontrarPorProveedor(ctx context.Context, proveedorID bson.ObjectID) ([]ProductoProveedor, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{
		"proveedor_id":      proveedorID,
		"fecha_eliminacion": nil,
	})
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

// La funcion guardar previamente colocada fue reemplazada por dos funciones distintas:
// Crear: para crear nuevos objetos y Actualizar: para actualizar viejos objetos.
func (r *RepositorioMongo) Crear(
	ctx context.Context,
	pp ProductoProveedor,
) error {
	filtro := bson.M{
		"producto_id":       pp.ProductoID,
		"proveedor_id":      pp.ProveedorID,
		"fecha_eliminacion": nil,
	}
	if err := r.coleccion.FindOne(ctx, filtro).Err(); err == nil {
		return ErrProductoYaSuministrado
	} else if !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}

	_, err := r.coleccion.InsertOne(ctx, pp)
	return err
}

func (r *RepositorioMongo) Actualizar(
	ctx context.Context,
	pp ProductoProveedor,
) error {
	filtro := bson.M{
		"producto_id":       pp.ProductoID,
		"proveedor_id":      pp.ProveedorID,
		"fecha_eliminacion": nil,
	}

	update := bson.M{
		"$set": bson.M{
			"precio_unitario":          pp.CostoUnidad,
			"tiempo_estimado_dias":     pp.TiempoEstimadoDias,
			"auditoria.modificado_por": pp.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": pp.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(ctx, filtro, update)
	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return ErrProductoSuministradoInexistente
	}

	return nil
}

func (r *RepositorioMongo) Eliminar(ctx context.Context, productoID, proveedorID, adminID bson.ObjectID) error {
	ahora := time.Now()

	filtro := bson.M{
		"producto_id":       productoID,
		"proveedor_id":      proveedorID,
		"fecha_eliminacion": nil,
	}

	actualizacion := bson.M{
		"$set": bson.M{
			"fecha_eliminacion":        ahora,
			"auditoria.modificado_por": adminID,
			"auditoria.actualizado_en": ahora,
		},
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		filtro,
		actualizacion,
	)
	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return ErrProductoSuministradoInexistente
	}

	return nil
}

// EliminarTodosDeProveedor se usa al eliminar (soft-delete) un proveedor
// entero: no tendría sentido dejar vínculos sueltos a un proveedor dado de baja.
func (r *RepositorioMongo) EliminarTodosDeProveedor(
	ctx context.Context,
	proveedorID bson.ObjectID,
	adminID bson.ObjectID,
) error {
	ahora := time.Now()

	filtro := bson.M{
		"proveedor_id":      proveedorID,
		"fecha_eliminacion": nil,
	}

	actualizacion := bson.M{
		"$set": bson.M{
			"fecha_eliminacion":        ahora,
			"auditoria.modificado_por": adminID,
			"auditoria.actualizado_en": ahora,
		},
	}

	_, err := r.coleccion.UpdateMany(
		ctx,
		filtro,
		actualizacion,
	)

	return err
}
