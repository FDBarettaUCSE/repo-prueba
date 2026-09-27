package transferencia

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EstadoTransferencia string

const (
	TransferenciaSolicitada EstadoTransferencia = "solicitada"
	TransferenciaAprobada   EstadoTransferencia = "aprobada"
	TransferenciaRechazada  EstadoTransferencia = "rechazada"
	TransferenciaCompletada EstadoTransferencia = "completada"
)

type Transferencia struct {
	ID                   bson.ObjectID       `bson:"_id,omitempty"`
	DepositoOrigenID     bson.ObjectID       `bson:"deposito_origen_id"`
	DepositoDestinoID    bson.ObjectID       `bson:"deposito_destino_id"`
	ProductoID           bson.ObjectID       `bson:"producto_id"`
	Cantidad             int32               `bson:"cantidad"`
	Estado               EstadoTransferencia `bson:"estado"`
	UsuarioSolicitanteID bson.ObjectID       `bson:"usuario_solicitante_id"`
	UsuarioResolucionID  *bson.ObjectID      `bson:"usuario_resolucion_id,omitempty"`
	FechaSolicitud       time.Time           `bson:"fecha_solicitud"`
	FechaResolucion      *time.Time          `bson:"fecha_resolucion,omitempty"`
	Motivo               string              `bson:"motivo,omitempty"`
	Auditoria            auditoria.Auditoria `bson:"auditoria"`
}
