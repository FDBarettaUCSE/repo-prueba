package ordendecompra

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrdenCompraRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (OrdenCompra, error)
	EncontrarTodos(ctx context.Context) ([]OrdenCompra, error)
	CrearOrdenCompra(ctx context.Context, orden OrdenCompra) (OrdenCompra, error)
	ActualizarOrdenCompra(ctx context.Context, id string, orden OrdenCompra) (OrdenCompra, error)
	EliminarOrdenCompra(ctx context.Context, id string) error
	ActualizarEstadoOrdenCompra(ctx context.Context, id string, estado EstadoOrdenCompra) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) OrdenCompraRepositorio {
	return &RepositorioMongo{
		coleccion: coleccion,
	}
}

var _ OrdenCompraRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorId(
	ctx context.Context,
	id string,
) (OrdenCompra, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return OrdenCompra{}, err
	}

	var orden OrdenCompra

	if err := r.coleccion.FindOne(
		ctx,
		bson.M{
			"_id": oid,
		},
	).Decode(&orden); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return OrdenCompra{}, errors.New("orden de compra no encontrada")
		}

		return OrdenCompra{}, err
	}

	return orden, nil
}

func (r *RepositorioMongo) EncontrarTodos(
	ctx context.Context,
) ([]OrdenCompra, error) {

	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var ordenes []OrdenCompra

	if err := cursor.All(ctx, &ordenes); err != nil {
		return nil, err
	}

	return ordenes, nil
}

func (r *RepositorioMongo) CrearOrdenCompra(
	ctx context.Context,
	orden OrdenCompra,
) (OrdenCompra, error) {

	result, err := r.coleccion.InsertOne(ctx, orden)
	if err != nil {
		return OrdenCompra{}, err
	}

	orden.ID = result.InsertedID.(bson.ObjectID)

	return orden, nil
}

func (r *RepositorioMongo) ActualizarOrdenCompra(
	ctx context.Context,
	id string,
	orden OrdenCompra,
) (OrdenCompra, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return OrdenCompra{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"proveedor_id":             orden.ProveedorID,
			"deposito_id":              orden.DepositoID,
			"estado":                   orden.Estado,
			"auditoria.modificado_por": orden.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": orden.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		update,
	)

	if err != nil {
		return OrdenCompra{}, err
	}

	if resultado.MatchedCount == 0 {
		return OrdenCompra{}, errors.New("orden de compra no encontrada")
	}

	orden.ID = oid

	return orden, nil
}

func (r *RepositorioMongo) EliminarOrdenCompra(
	ctx context.Context,
	id string,
) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	resultado, err := r.coleccion.DeleteOne(
		ctx,
		bson.M{"_id": oid},
	)

	if err != nil {
		return err
	}

	if resultado.DeletedCount == 0 {
		return errors.New("orden de compra no encontrada")
	}

	return nil
}

func (r *RepositorioMongo) ActualizarEstadoOrdenCompra(
	ctx context.Context,
	id string,
	estado EstadoOrdenCompra,
) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{
			"$set": bson.M{
				"estado": estado,
			},
		},
	)

	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("orden de compra no encontrada")
	}

	return nil
}
