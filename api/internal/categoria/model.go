package categoria

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Categoria struct {
	ID               bson.ObjectID       `bson:"_id, omitempty"`
	Nombre           string              `bson:"nombre"`
	FechaEliminacion *time.Time          `bson:"fecha_eliminacion,omitempty"`
	Auditoria        auditoria.Auditoria `bson:"auditoria"`
}

func MapeoACategoriaDTO(c Categoria) CategoriaDTO {
	return CategoriaDTO{
		ID:     c.ID.Hex(),
		Nombre: c.Nombre,
	}
}
