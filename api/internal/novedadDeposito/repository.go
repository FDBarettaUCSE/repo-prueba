package novedaddeposito

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// NovedadDepositoRepositorio persiste las novedades. Igual que en
// movimientoStock, no hay Actualizar genérico: solo las dos transiciones de
// estado válidas (MarcarLeida, Responder) tienen su propio método, para que
// no se pueda pisar por accidente el Texto o la FechaEnvio originales de una
// novedad ya enviada.
type NovedadDepositoRepositorio interface {
	Crear(ctx context.Context, n NovedadDeposito) (NovedadDeposito, error)
	EncontrarPorId(ctx context.Context, id string) (NovedadDeposito, error)
	ListarPorDeposito(ctx context.Context, depositoID bson.ObjectID, soloNoLeidas bool, pagina, tamañoPagina int) ([]NovedadDeposito, int64, error)
	ContarNoLeidas(ctx context.Context, depositoID bson.ObjectID) (int64, error)
	MarcarLeida(ctx context.Context, id string) (NovedadDeposito, error)
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
	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&n); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NovedadDeposito{}, errors.New("novedad no encontrada")
		}
		return NovedadDeposito{}, err
	}
	return n, nil
}

// ListarPorDeposito ordena por fecha_envio descendente (la más reciente
// primero) y pagina con skip/limit. El total que devuelve es el del filtro
// aplicado (con o sin "solo no leídas"), no el total general del depósito —
// para eso está ContarNoLeidas, aparte.
func (r *RepositorioMongo) ListarPorDeposito(ctx context.Context, depositoID bson.ObjectID, soloNoLeidas bool, pagina, tamañoPagina int) ([]NovedadDeposito, int64, error) {
	consulta := bson.M{"deposito_id": depositoID}
	if soloNoLeidas {
		consulta["leido"] = false
	}

	total, err := r.coleccion.CountDocuments(ctx, consulta)
	if err != nil {
		return nil, 0, err
	}

	if pagina < 1 {
		pagina = 1
	}
	if tamañoPagina < 1 {
		tamañoPagina = 20
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "fecha_envio", Value: -1}}).
		SetSkip(int64((pagina - 1) * tamañoPagina)).
		SetLimit(int64(tamañoPagina))

	cursor, err := r.coleccion.Find(ctx, consulta, opts)
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

func (r *RepositorioMongo) ContarNoLeidas(ctx context.Context, depositoID bson.ObjectID) (int64, error) {
	return r.coleccion.CountDocuments(ctx, bson.M{"deposito_id": depositoID, "leido": false})
}

func (r *RepositorioMongo) MarcarLeida(ctx context.Context, id string) (NovedadDeposito, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return NovedadDeposito{}, err
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var n NovedadDeposito
	err = r.coleccion.FindOneAndUpdate(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{"leido": true}}, opts).Decode(&n)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NovedadDeposito{}, errors.New("novedad no encontrada")
		}
		return NovedadDeposito{}, err
	}
	return n, nil
}

// Responder graba la respuesta y marca la novedad como leída en la MISMA
// actualización: responder implica haberla leído, así que no tendría
// sentido que quedara "no leída" después de contestarla.
func (r *RepositorioMongo) Responder(ctx context.Context, id string, respuesta RespuestaNovedad) (NovedadDeposito, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return NovedadDeposito{}, err
	}

	update := bson.M{
		"$set": bson.M{
			"leido":     true,
			"respuesta": respuesta,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var n NovedadDeposito
	err = r.coleccion.FindOneAndUpdate(ctx, bson.M{"_id": oid}, update, opts).Decode(&n)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return NovedadDeposito{}, errors.New("novedad no encontrada")
		}
		return NovedadDeposito{}, err
	}
	return n, nil
}
