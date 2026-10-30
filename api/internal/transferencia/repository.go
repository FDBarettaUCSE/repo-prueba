package transferencia

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repositorio interface {
	Crear(ctx context.Context, t Transferencia) (Transferencia, error)
	EncontrarPorID(ctx context.Context, id string) (Transferencia, error)
	Listar(ctx context.Context, filtro FiltroTransferenciasDTO, depositoPermitido string, usuarioID string) ([]Transferencia, int64, error)
	ActualizarEstado(ctx context.Context, id string, estadoEsperado EstadoTransferencia, nuevoEstado EstadoTransferencia, resolucionID *bson.ObjectID, fechaResolucion *time.Time, adminID string, motivo string) error
}

type RepositorioMongo struct {
	coleccion *mongo.Collection
}

func NuevoRepositorioMongo(col *mongo.Collection) *RepositorioMongo {
	return &RepositorioMongo{coleccion: col}
}

var _ Repositorio = (*RepositorioMongo)(nil)

func (r *RepositorioMongo) Crear(ctx context.Context, t Transferencia) (Transferencia, error) {
	res, err := r.coleccion.InsertOne(ctx, t)
	if err != nil {
		return Transferencia{}, err
	}
	t.ID = res.InsertedID.(bson.ObjectID)
	return t, nil
}

func (r *RepositorioMongo) EncontrarPorID(ctx context.Context, id string) (Transferencia, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return Transferencia{}, err
	}
	var t Transferencia
	if err := r.coleccion.FindOne(ctx, bson.M{"_id": oid}).Decode(&t); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Transferencia{}, ErrTransferenciaInexistente
		}
		return Transferencia{}, err
	}
	return t, nil
}

func (r *RepositorioMongo) Listar(
	ctx context.Context,
	filtro FiltroTransferenciasDTO,
	depositoPermitido string,
	usuarioID string,
) ([]Transferencia, int64, error) {

	query := bson.M{}

	// Filtrar por depósito cuando el usuario tiene uno asignado.
	if depositoPermitido != "" {
		depOID, err := bson.ObjectIDFromHex(depositoPermitido)
		if err != nil {
			return nil, 0, err
		}

		query["$or"] = []bson.M{
			{"deposito_origen_id": depOID},
			{"deposito_destino_id": depOID},
		}
	} else if filtro.DepositoID != "" {
		depOID, err := bson.ObjectIDFromHex(filtro.DepositoID)
		if err != nil {
			return nil, 0, err
		}

		query["$or"] = []bson.M{
			{"deposito_origen_id": depOID},
			{"deposito_destino_id": depOID},
		}
	}

	// Si se recibe un usuarioID, filtrar por solicitante.
	if usuarioID != "" {
		usuarioOID, err := bson.ObjectIDFromHex(usuarioID)
		if err != nil {
			return nil, 0, err
		}

		query["usuario_solicitante_id"] = usuarioOID
	}

	// Filtrar por estado si corresponde.
	if filtro.Estado != "" {
		query["estado"] = filtro.Estado
	}

	// Contar los resultados que cumplen todos los filtros.
	total, err := r.coleccion.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Ordenar por fecha de solicitud descendente.
	opts := options.Find().SetSort(
		bson.D{{Key: "fecha_solicitud", Value: -1}},
	)

	// Aplicar paginación.
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

	cur, err := r.coleccion.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cur.Close(ctx)

	var transferencias []Transferencia

	if err := cur.All(ctx, &transferencias); err != nil {
		return nil, 0, err
	}

	return transferencias, total, nil
}

// ActualizarEstado realiza una transición de estado atómica filtrando por
// {_id, estado: estadoEsperado}. Si ninguna fila coincide (por ejemplo, otra
// petición concurrente ya cambió el estado), devuelve ErrEstadoInvalido.
func (r *RepositorioMongo) ActualizarEstado(
	ctx context.Context,
	id string,
	estadoEsperado EstadoTransferencia,
	nuevoEstado EstadoTransferencia,
	resolucionID *bson.ObjectID,
	fechaResolucion *time.Time,
	usuarioModificadorID string,
	motivo string,
) error {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	uModOID, err := bson.ObjectIDFromHex(usuarioModificadorID)
	if err != nil {
		return err
	}

	setFields := bson.M{
		"estado":                   nuevoEstado,
		"auditoria.modificado_por": uModOID,
		"auditoria.actualizado_en": time.Now(),
	}
	if resolucionID != nil {
		setFields["usuario_resolucion_id"] = resolucionID
	}
	if fechaResolucion != nil {
		setFields["fecha_resolucion"] = fechaResolucion
	}
	if motivo != "" {
		setFields["motivo"] = motivo
	}

	filtro := bson.M{
		"_id":    oid,
		"estado": estadoEsperado,
	}

	res, err := r.coleccion.UpdateOne(ctx, filtro, bson.M{"$set": setFields})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return ErrEstadoInvalido
	}
	return nil
}
