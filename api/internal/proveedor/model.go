package proveedor

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Proveedor struct {
	ID               bson.ObjectID       `bson:"_id,omitempty"`
	Nombre           string              `bson:"nombre"`
	CUIT             string              `bson:"cuit"`
	Email            string              `bson:"email"`
	Telefono         string              `bson:"telefono"`
	FechaEliminacion *time.Time          `bson:"fecha_eliminacion,omitempty"`
	Auditoria        auditoria.Auditoria `bson:"auditoria"`
}

func MapeoAProveedorDTO(p Proveedor) ProveedorDTO {
	return ProveedorDTO{
		ID:       p.ID.Hex(),
		Nombre:   p.Nombre,
		CUIT:     p.CUIT,
		Email:    p.Email,
		Telefono: p.Telefono,
	}
}
