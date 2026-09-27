package movimientostock

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TipoMovimiento string

const (
	Ingreso       TipoMovimiento = "ingreso"
	Egreso        TipoMovimiento = "egreso"
	Ajuste        TipoMovimiento = "ajuste"
	Transferencia TipoMovimiento = "transferencia"
)

type MovimientoStock struct {
	ID                   bson.ObjectID       `bson:"_id,omitempty"`
	ProductoID           bson.ObjectID       `bson:"producto_id"`
	DepositoID           bson.ObjectID       `bson:"deposito_id"`
	Cantidad             int32               `bson:"cantidad"`
	Tipo                 TipoMovimiento      `bson:"tipo"`
	UsuarioResponsableID bson.ObjectID       `bson:"usuario_responsable_id"`
	FechaHora            time.Time           `bson:"fecha_hora"`
	Motivo               string              `bson:"motivo"`
	Auditoria            auditoria.Auditoria `bson:"auditoria"`
}
