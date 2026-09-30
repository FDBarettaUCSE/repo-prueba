package deposito

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type DepositoRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (Deposito, error)
	EncontrarTodos(ctx context.Context) ([]Deposito, error)
	CrearDeposito(ctx context.Context, d Deposito) (Deposito, error)
	ActualizarDeposito(ctx context.Context, id string, d Deposito) (Deposito, error)
	EliminarDeposito(ctx context.Context, id, idAdmin string) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ DepositoRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (Deposito, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Deposito{}, err
	}

	var deposito Deposito

	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid, "fecha_eliminacion": nil}).Decode(&deposito); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Deposito{}, errors.New("libro no encontrado")
		}
		return Deposito{}, err
	}
	return deposito, nil
}

func (r *RepositorioMongo) CrearDeposito(ctx context.Context, d Deposito) (Deposito, error) {
	result, err := r.coleccion.InsertOne(ctx, d)
	if err != nil {
		return Deposito{}, err
	}

	d.ID = result.InsertedID.(bson.ObjectID)
	return d, nil
}

func (r *RepositorioMongo) EncontrarTodos(ctx context.Context) ([]Deposito, error) {
	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{"fecha_eliminacion": nil},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var depositos []Deposito

	if err := cursor.All(ctx, &depositos); err != nil {
		return nil, err
	}

	return depositos, nil
}

func (r *RepositorioMongo) ActualizarDeposito(ctx context.Context, id string, d Deposito) (Deposito, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Deposito{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"nombre":                   d.Nombre,
			"localidad":                d.Localidad,
			"provincia":                d.Provincia,
			"auditoria.modificado_por": d.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": d.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		update,
	)

	if err != nil {
		return Deposito{}, err
	}

	if resultado.MatchedCount == 0 {
		return Deposito{}, errors.New("depósito no encontrado")
	}

	d.ID = oid

	return d, nil
}

func (r *RepositorioMongo) EliminarDeposito(ctx context.Context, id, idAdmin string) error {

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
		bson.M{
			"_id":               oid,
			"fecha_eliminacion": nil,
		},
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
		return errors.New("depósito no encontrado")
	}

	return nil
}
