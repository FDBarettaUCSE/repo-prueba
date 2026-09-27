package usuario

import (
	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Rol string

const (
	Administrador Rol = "administrador"
	Gerente       Rol = "gerente"
	Operario      Rol = "operario"
	Auditor       Rol = "auditor"
)

type Usuario struct {
	ID                 bson.ObjectID       `bson:"_id,omitempty"`
	Correo             string              `bson:"correo"`
	ContraseñaHasheada string              `json:"-" bson:"contraseña_hasheada"`
	Rol                Rol                 `bson:"rol"`
	Activado           bool                `bson:"activado"`
	DepositoAsociado   bson.ObjectID       `bson:"deposito_asociado,omitempty"`
	Auditoria          auditoria.Auditoria `bson:"auditoria"`
}
