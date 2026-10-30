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
