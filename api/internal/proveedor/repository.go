package proveedor

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProveedorRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (Proveedor, error)
	EncontrarTodos(ctx context.Context) ([]Proveedor, error)
	EncontrarPorCUIT(ctx context.Context, cuit string) (Proveedor, error)
	CrearProveedor(ctx context.Context, p Proveedor) (Proveedor, error)
	ActualizarProveedor(ctx context.Context, id string, p Proveedor) (Proveedor, error)
	EliminarProveedor(ctx context.Context, id, idAdmin string) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ ProveedorRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (Proveedor, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Proveedor{}, err
	}

	var proveedor Proveedor
	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&proveedor); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Proveedor{}, errors.New("proveedor no encontrado")
		}
		return Proveedor{}, err
	}
	return proveedor, nil
}

func (r *RepositorioMongo) EncontrarPorCUIT(ctx context.Context, cuit string) (Proveedor, error) {
	var proveedor Proveedor
	if err := r.coleccion.FindOne(ctx, bson.M{"cuit": cuit}).Decode(&proveedor); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Proveedor{}, errors.New("no existe un proveedor con ese CUIT")
		}
		return Proveedor{}, err
	}
	return proveedor, nil
}

func (r *RepositorioMongo) CrearProveedor(ctx context.Context, p Proveedor) (Proveedor, error) {
	result, err := r.coleccion.InsertOne(ctx, p)
	if err != nil {
		return Proveedor{}, err
	}

	p.ID = result.InsertedID.(bson.ObjectID)
	return p, nil
}

func (r *RepositorioMongo) EncontrarTodos(ctx context.Context) ([]Proveedor, error) {
	cursor, err := r.coleccion.Find(ctx, bson.M{"fecha_eliminacion": nil})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var proveedores []Proveedor
	if err := cursor.All(ctx, &proveedores); err != nil {
		return nil, err
	}
	return proveedores, nil
}

func (r *RepositorioMongo) ActualizarProveedor(ctx context.Context, id string, p Proveedor) (Proveedor, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Proveedor{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"nombre":                   p.Nombre,
			"cuit":                     p.CUIT,
			"email":                    p.Email,
			"telefono":                 p.Telefono,
			"auditoria.modificado_por": p.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": p.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, update)
	if err != nil {
		return Proveedor{}, err
	}
	if resultado.MatchedCount == 0 {
		return Proveedor{}, errors.New("proveedor no encontrado")
	}

	p.ID = oid
	return p, nil
}

func (r *RepositorioMongo) EliminarProveedor(ctx context.Context, id, idAdmin string) error {
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
		bson.M{"_id": oid, "fecha_eliminacion": nil},
		bson.M{
			"$set": bson.M{
				"fecha_eliminacion":        ahora,
				"auditoria.modificado_por": oidAdmin,
				"auditoria.actualizado_en": ahora,
			},
		},
	)
	if err != nil {
		return err
	}
	if resultado.MatchedCount == 0 {
		return errors.New("proveedor no encontrado")
	}
	return nil
}
