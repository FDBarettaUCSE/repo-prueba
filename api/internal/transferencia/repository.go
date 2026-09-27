package transferencia

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repositorio interface {
	Crear(ctx context.Context, t Transferencia) (Transferencia, error)
	EncontrarPorID(ctx context.Context, id string) (Transferencia, error)
	Listar(ctx context.Context, filtro FiltroTransferenciasDTO, depositoPermitido string) ([]Transferencia, error)
	ActualizarEstado(ctx context.Context, id string, nuevoEstado EstadoTransferencia, resolucionID *bson.ObjectID, fechaResolucion *time.Time, adminID string, motivo string) error
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
			return Transferencia{}, errors.New("transferencia no encontrada")
		}
		return Transferencia{}, err
	}
	return t, nil
}

func (r *RepositorioMongo) Listar(ctx context.Context, filtro FiltroTransferenciasDTO, depositoPermitido string) ([]Transferencia, error) {
	query := bson.M{}

	// Si el usuario está restringido a un depósito (gerente/operario)
	if depositoPermitido != "" {
		depOID, err := bson.ObjectIDFromHex(depositoPermitido)
		if err == nil {
			query["$or"] = []bson.M{
				{"deposito_origen_id": depOID},
				{"deposito_destino_id": depOID},
			}
		}
	} else if filtro.DepositoID != "" {
		depOID, err := bson.ObjectIDFromHex(filtro.DepositoID)
		if err == nil {
			query["$or"] = []bson.M{
				{"deposito_origen_id": depOID},
				{"deposito_destino_id": depOID},
			}
		}
	}

	if filtro.Estado != "" {
		query["estado"] = filtro.Estado
	}

	cur, err := r.coleccion.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var transferencias []Transferencia
	if err := cur.All(ctx, &transferencias); err != nil {
		return nil, err
	}
	return transferencias, nil
}

func (r *RepositorioMongo) ActualizarEstado(ctx context.Context, id string, nuevoEstado EstadoTransferencia, resolucionID *bson.ObjectID, fechaResolucion *time.Time, usuarioModificadorID string, motivo string) error {
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

	res, err := r.coleccion.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": setFields})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return errors.New("transferencia no encontrada")
	}
	return nil
}
