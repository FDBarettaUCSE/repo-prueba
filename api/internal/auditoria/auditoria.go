package auditoria

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Auditoria struct {
	CreadoPor     bson.ObjectID  `json:"creadoPor" bson:"creado_por"`
	CreadoEn      time.Time      `json:"creadoEn" bson:"creado_en"`
	ModificadoPor *bson.ObjectID `json:"modificadoPor,omitempty" bson:"modificado_por,omitempty"`
	ActualizadoEn *time.Time     `json:"actualizadoEn,omitempty" bson:"actualizado_en,omitempty"`
}

func NuevaAuditoria(creadoPor, modificadoPor string, creadoEn, actualizadoEn time.Time) (Auditoria, error) {

	idCreado, err := bson.ObjectIDFromHex(creadoPor)
	if err != nil {
		return Auditoria{}, err // Error real: el ID del creador debe ser válido
	}

	var idModificado *bson.ObjectID
	var fechaActualizacion *time.Time

	// Solo convertimos el ID de modificación si el string NO está vacío
	if modificadoPor != "" {
		idHex, err := bson.ObjectIDFromHex(modificadoPor)
		if err != nil {
			return Auditoria{}, err
		}
		idModificado = &idHex
		fechaActualizacion = &actualizadoEn
	}

	return Auditoria{
		CreadoPor:     idCreado,
		CreadoEn:      creadoEn,
		ModificadoPor: idModificado,       // Será nil si nadie lo modificó
		ActualizadoEn: fechaActualizacion, // Será nil si nadie lo modificó
	}, nil
}
