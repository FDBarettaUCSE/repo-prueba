package ordendecompra

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type OrdenCompraRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (OrdenCompra, error)
	EncontrarTodos(ctx context.Context) ([]OrdenCompra, error)
	EncontrarPorDeposito(ctx context.Context, depositoID bson.ObjectID) ([]OrdenCompra, error)
	Listar(ctx context.Context, filtro FiltroOrdenesCompraDTO, depositoPermitido string) ([]OrdenCompra, int64, error)
	CrearOrdenCompra(ctx context.Context, orden OrdenCompra) (OrdenCompra, error)
	ActualizarOrdenCompra(ctx context.Context, id string, orden OrdenCompra) (OrdenCompra, error)
	EliminarOrdenCompra(ctx context.Context, id string) error
	ActualizarEstadoOrdenCompra(ctx context.Context, id string, estado EstadoOrdenCompra, idAdmin string) error
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
		return OrdenCompra{}, ErrIDOrdenCompraInvalido
	}

	var orden OrdenCompra

	if err := r.coleccion.FindOne(
		ctx,
		bson.M{
			"_id": oid,
		},
	).Decode(&orden); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return OrdenCompra{}, ErrOrdenCompraInexistente
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

func (r *RepositorioMongo) EncontrarPorDeposito(
	ctx context.Context,
	depositoID bson.ObjectID,
) ([]OrdenCompra, error) {

	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{"deposito_id": depositoID},
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
		return OrdenCompra{}, ErrOrdenCompraInexistente
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
		return ErrIDOrdenCompraInvalido
	}

	resultado, err := r.coleccion.DeleteOne(
		ctx,
		bson.M{"_id": oid},
	)

	if err != nil {
		return err
	}

	if resultado.DeletedCount == 0 {
		return ErrOrdenCompraInexistente
	}

	return nil
}

func (r *RepositorioMongo) ActualizarEstadoOrdenCompra(
	ctx context.Context,
	id string,
	estado EstadoOrdenCompra,
	idAdmin string,
) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrIDOrdenCompraInvalido
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
			"$set": bson.M{
				"estado":                   estado,
				"auditoria.modificado_por": oidAdmin,
				"auditoria.actualizado_en": ahora,
			},
		},
	)

	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return ErrOrdenCompraInexistente
	}

	return nil
}

func (r *RepositorioMongo) Listar(
	ctx context.Context,
	filtro FiltroOrdenesCompraDTO,
	depositoPermitido string,
) ([]OrdenCompra, int64, error) {
	query := bson.M{}

	if depositoPermitido != "" {
		if depOID, err := bson.ObjectIDFromHex(depositoPermitido); err == nil {
			query["deposito_id"] = depOID
		}
	} else if filtro.DepositoID != "" {
		if depOID, err := bson.ObjectIDFromHex(filtro.DepositoID); err == nil {
			query["deposito_id"] = depOID
		}
	}

	if filtro.Estado != "" {
		query["estado"] = filtro.Estado
	}

	if filtro.ProveedorID != "" {
		if provOID, err := bson.ObjectIDFromHex(filtro.ProveedorID); err == nil {
			query["proveedor_id"] = provOID
		}
	}

	total, err := r.coleccion.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "auditoria.creado_en", Value: -1}})

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

	cursor, err := r.coleccion.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var ordenes []OrdenCompra
	if err := cursor.All(ctx, &ordenes); err != nil {
		return nil, 0, err
	}

	return ordenes, total, nil
}

