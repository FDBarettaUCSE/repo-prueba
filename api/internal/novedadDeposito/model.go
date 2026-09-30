package novedaddeposito

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RespuestaNovedad es la respuesta de UN operario del depósito a una
// novedad. El enunciado aclara que "no se requiere un hilo de conversación
// bidireccional" — no hay que poder responderle a la respuesta, ni el
// gerente/auditor le contesta de vuelta al operario — así que alcanza con
// una respuesta embebida (no una subcolección con su propio historial).
type RespuestaNovedad struct {
	Texto          string        `bson:"texto"`
	UsuarioID      bson.ObjectID `bson:"usuario_id"`
	FechaRespuesta time.Time     `bson:"fecha_respuesta"`
}

// NovedadDeposito es un mensaje dirigido a los operarios de UN depósito
// puntual. Cuando el Auditor la envía a "varios depósitos, o a todos" (ver
// enunciado), eso se resuelve en el service creando UN documento
// independiente por depósito destino (fan-out) — así cada depósito tiene su
// propio estado de lectura/respuesta, sin pisarse entre sí.
type NovedadDeposito struct {
	ID         bson.ObjectID       `bson:"_id,omitempty"`
	DepositoID bson.ObjectID       `bson:"deposito_id"`
	Texto      string              `bson:"texto"`
	FechaEnvio time.Time           `bson:"fecha_envio"`
	Leido      bool                `bson:"leido"`
	Respuesta  *RespuestaNovedad   `bson:"respuesta,omitempty"`
	Auditoria  auditoria.Auditoria `bson:"auditoria"` // CreadoPor = quién la envió (gerente, auditor o administrador)
}
