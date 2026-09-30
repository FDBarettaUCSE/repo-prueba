package categoria

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type CategoriaRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (Categoria, error)
	EncontrarPorIds(ctx context.Context, id []string) ([]Categoria, error)
	EncontrarTodos(ctx context.Context) ([]Categoria, error)
	CrearCategoria(ctx context.Context, d Categoria) (Categoria, error)
	ActualizarCategoria(ctx context.Context, id string, d Categoria) (Categoria, error)
	EliminarCategoria(ctx context.Context, id, idAdmin string) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ CategoriaRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (Categoria, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Categoria{}, err
	}

	var categoria Categoria

	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid, "fecha_eliminacion": nil}).Decode(&categoria); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Categoria{}, errors.New("libro no encontrado")
		}
		return Categoria{}, err
	}
	return categoria, nil
}

func (r *RepositorioMongo) EncontrarPorIds(ctx context.Context, ids []string) ([]Categoria, error) {

	if len(ids) == 0 {
		return []Categoria{}, nil
	}

	oids := make([]bson.ObjectID, 0, len(ids))

	for _, id := range ids {
		oid, err := bson.ObjectIDFromHex(id)
		if err != nil {
			return nil, err
		}

		oids = append(oids, oid)
	}

	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{
			"_id": bson.M{
				"$in": oids,
			},
			"fecha_eliminacion": nil,
		},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var categorias []Categoria

	if err := cursor.All(ctx, &categorias); err != nil {
		return nil, err
	}

	return categorias, nil
}

func (r *RepositorioMongo) CrearCategoria(ctx context.Context, c Categoria) (Categoria, error) {
	result, err := r.coleccion.InsertOne(ctx, c)
	if err != nil {
		return Categoria{}, err
	}

	c.ID = result.InsertedID.(bson.ObjectID)
	return c, nil
}

func (r *RepositorioMongo) EncontrarTodos(ctx context.Context) ([]Categoria, error) {
	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{"fecha_eliminacion": nil},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var categorias []Categoria

	if err := cursor.All(ctx, &categorias); err != nil {
		return nil, err
	}

	return categorias, nil
}

func (r *RepositorioMongo) ActualizarCategoria(ctx context.Context, id string, c Categoria) (Categoria, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Categoria{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"nombre":                   c.Nombre,
			"auditoria.modificado_por": c.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": c.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		update,
	)

	if err != nil {
		return Categoria{}, err
	}

	if resultado.MatchedCount == 0 {
		return Categoria{}, errors.New("depósito no encontrado")
	}

	c.ID = oid

	return c, nil
}

func (r *RepositorioMongo) EliminarCategoria(ctx context.Context, id, idAdmin string) error {

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
		return errors.New("categoría no encontrada")
	}

	return nil
}
