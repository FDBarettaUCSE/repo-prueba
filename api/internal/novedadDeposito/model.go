package novedaddeposito

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RespuestaNovedad es la respuesta de UN operario del depósito a una
// novedad. El enunciado aclara que "no se requiere un hilo de conversación
// bidireccional" — no hay que poder responderle a la respuesta, ni el
// gerente/auditor le contesta de vuelta al operario. Igual se guarda una
// lista (una entrada por cada respuesta) para que la respuesta de un
// operario no pise la de otro del mismo depósito.
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
//
// El estado de lectura es POR OPERARIO: LeidoPor guarda los usuarios que ya
// la leyeron, así el contador de no leídas de cada operario es propio y la
// lectura de uno no la marca como leída para el resto del depósito.
type NovedadDeposito struct {
	ID         bson.ObjectID       `bson:"_id,omitempty"`
	DepositoID bson.ObjectID       `bson:"deposito_id"`
	Texto      string              `bson:"texto"`
	FechaEnvio time.Time           `bson:"fecha_envio"`
	LeidoPor   []bson.ObjectID     `bson:"leido_por"`
	Respuestas []RespuestaNovedad  `bson:"respuestas"`
	Auditoria  auditoria.Auditoria `bson:"auditoria"` // CreadoPor = quién la envió (gerente, auditor o administrador)
}

// LeidaPor indica si el usuario ya leyó la novedad.
func (n NovedadDeposito) LeidaPor(usuarioID bson.ObjectID) bool {
	for _, id := range n.LeidoPor {
		if id == usuarioID {
			return true
		}
	}
	return false
}
