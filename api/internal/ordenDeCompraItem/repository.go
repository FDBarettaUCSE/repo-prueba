package ordendecompraitem

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrdenCompraItemRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (OrdenCompraItem, error)
	EncontrarPorOrdenCompra(ctx context.Context, ordenCompraID string) ([]OrdenCompraItem, error)
	EncontrarTodos(ctx context.Context) ([]OrdenCompraItem, error)
	CrearOrdenCompraItem(ctx context.Context, item OrdenCompraItem) (OrdenCompraItem, error)
	ActualizarOrdenCompraItem(ctx context.Context, id string, item OrdenCompraItem) (OrdenCompraItem, error)
	EliminarOrdenCompraItem(ctx context.Context, id string) error
	EliminarPorOrdenCompra(ctx context.Context, ordenCompraID string) error
	RegistrarRecepcion(ctx context.Context, id string, cantidad int32, idAdmin string) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{
		coleccion: coleccion,
	}
}

var _ OrdenCompraItemRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorId(
	ctx context.Context,
	id string,
) (OrdenCompraItem, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	var item OrdenCompraItem

	if err := r.coleccion.FindOne(
		ctx,
		bson.M{"_id": oid},
	).Decode(&item); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return OrdenCompraItem{}, errors.New(
				"item de orden de compra no encontrado",
			)
		}

		return OrdenCompraItem{}, err
	}

	return item, nil
}

func (r *RepositorioMongo) EncontrarPorOrdenCompra(
	ctx context.Context,
	ordenCompraID string,
) ([]OrdenCompraItem, error) {

	oid, err := bson.ObjectIDFromHex(ordenCompraID)
	if err != nil {
		return nil, err
	}

	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{"orden_compra_id": oid},
	)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var items []OrdenCompraItem

	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *RepositorioMongo) EncontrarTodos(
	ctx context.Context,
) ([]OrdenCompraItem, error) {

	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{},
	)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var items []OrdenCompraItem

	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *RepositorioMongo) CrearOrdenCompraItem(
	ctx context.Context,
	item OrdenCompraItem,
) (OrdenCompraItem, error) {

	result, err := r.coleccion.InsertOne(
		ctx,
		item,
	)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	item.ID = result.InsertedID.(bson.ObjectID)

	return item, nil
}

func (r *RepositorioMongo) ActualizarOrdenCompraItem(
	ctx context.Context,
	id string,
	item OrdenCompraItem,
) (OrdenCompraItem, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"producto_id":              item.ProductoID,
			"cantidad_solicitada":      item.CantidadSolicitada,
			"cantidad_recibida":        item.CantidadRecibida,
			"costo_unitario":           item.CostoUnitario,
			"auditoria.modificado_por": item.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": item.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		update,
	)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	if resultado.MatchedCount == 0 {
		return OrdenCompraItem{}, errors.New(
			"item de orden de compra no encontrado",
		)
	}

	item.ID = oid

	return item, nil
}

func (r *RepositorioMongo) EliminarOrdenCompraItem(
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
		return errors.New(
			"item de orden de compra no encontrado",
		)
	}

	return nil
}

func (r *RepositorioMongo) EliminarPorOrdenCompra(ctx context.Context, ordenCompraID string) error {

	oid, err := bson.ObjectIDFromHex(ordenCompraID)
	if err != nil {
		return err
	}

	_, err = r.coleccion.DeleteMany(
		ctx,
		bson.M{"orden_compra_id": oid},
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *RepositorioMongo) RegistrarRecepcion(
	ctx context.Context,
	id string,
	cantidad int32,
	idAdmin string,
) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	ahora := time.Now()

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		bson.M{
			"$inc": bson.M{
				"cantidad_recibida": cantidad,
			},
			"$set": bson.M{
				"auditoria.modificado_por": oidAdmin,
				"auditoria.actualizado_en": ahora,
			},
		},
	)

	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("item de orden de compra no encontrado")
	}

	return nil
}
