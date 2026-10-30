package novedaddeposito

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NovedadDepositoRepositorio persiste las novedades.
type NovedadDepositoRepositorio interface {
	Crear(ctx context.Context, n NovedadDeposito) (NovedadDeposito, error)
	EncontrarPorId(ctx context.Context, id string) (NovedadDeposito, error)
	ListarPorDeposito(ctx context.Context, depositoID bson.ObjectID, lector *bson.ObjectID, soloNoLeidas bool, pagina, tamanioPagina int) ([]NovedadDeposito, int64, error)
	ContarNoLeidas(ctx context.Context, depositoID bson.ObjectID, lector *bson.ObjectID) (int64, error)
	MarcarLeida(ctx context.Context, id string, usuarioID bson.ObjectID) (NovedadDeposito, error)
	Responder(ctx context.Context, id string, respuesta RespuestaNovedad) (NovedadDeposito, error)
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(coleccion *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: coleccion}
}

var _ NovedadDepositoRepositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) Crear(ctx context.Context, n NovedadDeposito) (NovedadDeposito, error) {
	resultado, err := r.coleccion.InsertOne(ctx, n)
	if err != nil {
		return NovedadDeposito{}, err
	}

	n.ID = resultado.InsertedID.(bson.ObjectID)
	return n, nil
}

func (r *RepositorioMongo) EncontrarPorId(ctx context.Context, id string) (NovedadDeposito, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return NovedadDeposito{}, err
	}

	var n NovedadDeposito

	if err := r.coleccion.FindOne(
		ctx,
		bson.M{"_id": oid},
	).Decode(&n); err != nil {

		if errors.Is(err, mongo.ErrNoDocuments) {
			return NovedadDeposito{}, ErrNovedadNoEncontrada
		}

		return NovedadDeposito{}, err
	}

	return n, nil
}

// aplicarFiltroNoLeidas restringe la consulta a las novedades no leídas:
// si hay un lector (operario), las que no contienen su ID en leido_por;
// si lector es nil (gerente/auditor), las que ningún operario leyó todavía.
func aplicarFiltroNoLeidas(consulta bson.M, lector *bson.ObjectID) {
	if lector != nil {
		consulta["leido_por"] = bson.M{"$ne": *lector}
	} else {
		consulta["leido_por.0"] = bson.M{"$exists": false}
	}
}

// ListarPorDeposito ordena por fecha_envio descendente
// (la más reciente primero) y pagina con skip/limit.
func (r *RepositorioMongo) ListarPorDeposito(
	ctx context.Context,
	depositoID bson.ObjectID,
	lector *bson.ObjectID,
	soloNoLeidas bool,
	pagina,
	tamanioPagina int,
) ([]NovedadDeposito, int64, error) {

	consulta := bson.M{
		"deposito_id": depositoID,
	}

	if soloNoLeidas {
		aplicarFiltroNoLeidas(consulta, lector)
	}

	total, err := r.coleccion.CountDocuments(ctx, consulta)
	if err != nil {
		return nil, 0, err
	}

	if pagina < 1 {
		pagina = 1
	}

	if tamanioPagina < 1 {
		tamanioPagina = 20
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "fecha_envio", Value: -1}}).
		SetSkip(int64((pagina - 1) * tamanioPagina)).
		SetLimit(int64(tamanioPagina))

	cursor, err := r.coleccion.Find(
		ctx,
		consulta,
		opts,
	)
	if err != nil {
		return nil, 0, err
	}

	defer cursor.Close(ctx)

	var novedades []NovedadDeposito

	if err := cursor.All(ctx, &novedades); err != nil {
		return nil, 0, err
	}

	return novedades, total, nil
}

func (r *RepositorioMongo) ContarNoLeidas(
	ctx context.Context,
	depositoID bson.ObjectID,
	lector *bson.ObjectID,
) (int64, error) {

	consulta := bson.M{
		"deposito_id": depositoID,
	}
	aplicarFiltroNoLeidas(consulta, lector)

	return r.coleccion.CountDocuments(ctx, consulta)
}

func (r *RepositorioMongo) MarcarLeida(
	ctx context.Context,
	id string,
	usuarioID bson.ObjectID,
) (NovedadDeposito, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return NovedadDeposito{}, err
	}

	ahora := time.Now()
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var n NovedadDeposito

	err = r.coleccion.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		bson.M{
			"$addToSet": bson.M{
				"leido_por": usuarioID,
			},
			"$set": bson.M{
				"auditoria.modificado_por": usuarioID,
				"auditoria.actualizado_en": ahora,
			},
		},
		opts,
	).Decode(&n)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NovedadDeposito{}, ErrNovedadNoEncontrada
		}

		return NovedadDeposito{}, err
	}

	return n, nil
}

// Responder agrega la respuesta del operario a la lista de respuestas y
// registra al usuario en leido_por en la misma actualización.
func (r *RepositorioMongo) Responder(
	ctx context.Context,
	id string,
	respuesta RespuestaNovedad,
) (NovedadDeposito, error) {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return NovedadDeposito{}, err
	}

	update := bson.M{
		"$addToSet": bson.M{
			"leido_por": respuesta.UsuarioID,
		},
		"$push": bson.M{
			"respuestas": respuesta,
		},
		"$set": bson.M{
			"auditoria.modificado_por": respuesta.UsuarioID,
			"auditoria.actualizado_en": respuesta.FechaRespuesta,
		},
	}

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var n NovedadDeposito

	err = r.coleccion.FindOneAndUpdate(
		ctx,
		bson.M{"_id": oid},
		update,
		opts,
	).Decode(&n)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NovedadDeposito{}, ErrNovedadNoEncontrada
		}

		return NovedadDeposito{}, err
	}

	return n, nil
}
