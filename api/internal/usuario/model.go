package usuario

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Rol string

const (
	Administrador Rol = "administrador"
	Gerente       Rol = "gerente"
	Operario      Rol = "operario"
	Auditor       Rol = "auditor"
)

// EventoAuditoriaUsuario deja registro inmutable de cada operación sensible
// realizada sobre un usuario (alta, cambio de rol/depósito, cambio de
// contraseña, activación y desactivación), preservando el historial completo
// en lugar de conocer únicamente la última modificación.
type EventoAuditoriaUsuario struct {
	Accion       string        `json:"accion" bson:"accion"`
	Detalle      string        `json:"detalle,omitempty" bson:"detalle,omitempty"`
	RealizadoPor bson.ObjectID `json:"realizado_por" bson:"realizado_por"`
	Fecha        time.Time     `json:"fecha" bson:"fecha"`
}

type Usuario struct {
	ID                  bson.ObjectID            `bson:"_id,omitempty"`
	Correo              string                   `bson:"correo"`
	ContraseniaHasheada string                   `json:"-" bson:"contrasenia_hasheada"`
	Rol                 Rol                      `bson:"rol"`
	Activado            bool                     `bson:"activado"`
	DepositoAsociado    bson.ObjectID            `bson:"deposito_asociado,omitempty"`
	Auditoria           auditoria.Auditoria      `bson:"auditoria"`
	HistorialAuditoria  []EventoAuditoriaUsuario `bson:"historial_auditoria,omitempty"`
}
