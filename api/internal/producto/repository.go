package producto

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProductoRepositorio interface {
	EncontrarPorId(ctx context.Context, id string) (Producto, error)
	EncontrarTodos(ctx context.Context) ([]Producto, error)
	CrearProducto(ctx context.Context, d Producto) (Producto, error)
	ActualizarProducto(ctx context.Context, id string, d Producto) (Producto, error)
	ActualizarStockMinimo(ctx context.Context, id string, stockMinimo int32, idAdmin string) error
	EliminarProducto(ctx context.Context, id, idAdmin string) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ ProductoRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (Producto, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Producto{}, err
	}

	var producto Producto

	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&producto); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Producto{}, errors.New("producto no encontrado")
		}
		return Producto{}, err
	}
	return producto, nil
}

func (r *RepositorioMongo) CrearProducto(ctx context.Context, p Producto) (Producto, error) {
	result, err := r.coleccion.InsertOne(ctx, p)
	if err != nil {
		return Producto{}, err
	}

	p.ID = result.InsertedID.(bson.ObjectID)
	return p, nil
}

func (r *RepositorioMongo) EncontrarTodos(ctx context.Context) ([]Producto, error) {
	cursor, err := r.coleccion.Find(
		ctx,
		bson.M{"fecha_eliminacion": nil},
	)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	var productos []Producto

	if err := cursor.All(ctx, &productos); err != nil {
		return nil, err
	}

	return productos, nil
}

func (r *RepositorioMongo) ActualizarProducto(ctx context.Context, id string, p Producto) (Producto, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Producto{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"nombre":                   p.Nombre,
			"descripcion":              p.Descripcion,
			"categoria":                p.Categoria,
			"unidad_medida":            p.UnidadMedida,
			"costo_unitario":           p.CostoUnitario,
			"fecha_vencimiento":        p.FechaVencimiento,
			"auditoria.modificado_por": p.Auditoria.ModificadoPor,
			"auditoria.actualizado_en": p.Auditoria.ActualizadoEn,
		},
	}

	resultado, err := r.coleccion.UpdateOne(
		ctx,
		bson.M{"_id": oid},
		update,
	)

	if err != nil {
		return Producto{}, err
	}

	if resultado.MatchedCount == 0 {
		return Producto{}, errors.New("producto no encontrado")
	}

	p.ID = oid

	return p, nil
}

func (r *RepositorioMongo) EliminarProducto(ctx context.Context, id, idAdmin string) error {

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
		return errors.New("producto no encontrado")
	}

	return nil
}

func (r *RepositorioMongo) ActualizarStockMinimo(
	ctx context.Context,
	id string,
	stockMinimo int32,
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
			"$set": bson.M{
				"stock_minimo":             stockMinimo,
				"auditoria.modificado_por": oidAdmin,
				"auditoria.actualizado_en": ahora,
			},
		},
	)

	if err != nil {
		return err
	}

	if resultado.MatchedCount == 0 {
		return errors.New("producto no encontrado")
	}

	return nil
}
