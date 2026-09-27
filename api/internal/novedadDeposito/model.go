package novedaddeposito

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type NovedadDeposito struct {
	ID         bson.ObjectID       `bson:"_id,omitempty"`
	DepositoID bson.ObjectID       `bson:"deposito_id"`
	Texto      string              `bson:"texto"`
	FechaEnvio time.Time           `bson:"fecha_envio"`
	Leido      bool                `bson:"leido"`
	Auditoria  auditoria.Auditoria `bson:"auditoria"`
}

type RespuestaNovedad struct {
	ID                bson.ObjectID       `bson:"_id,omitempty"`
	NovedadDepositoId bson.ObjectID       `bson:"deposito_id"`
	Auditoria         auditoria.Auditoria `bson:"auditoria"`
}
