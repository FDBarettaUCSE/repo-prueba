package deposito

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Deposito struct {
	ID               bson.ObjectID       `bson:"_id,omitempty"`
	Nombre           string              `bson:"nombre"`
	Localidad        string              `bson:"localidad"`
	Provincia        string              `bson:"provincia"`
	FechaEliminacion *time.Time          `bson:"fecha_eliminacion,omitempty"`
	Auditoria        auditoria.Auditoria `bson:"auditoria"`
}

func MapeoADepositoDTO(d Deposito) DepositoDTO {
	return DepositoDTO{
		ID:        d.ID.Hex(),
		Nombre:    d.Nombre,
		Localidad: d.Localidad,
		Provincia: d.Provincia,
	}
}
