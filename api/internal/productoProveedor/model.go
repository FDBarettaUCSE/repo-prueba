package productoproveedor

import (
	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductoProveedor struct {
	ProductoID         bson.ObjectID       `bson:"producto_id"`
	ProveedorID        bson.ObjectID       `bson:"proveedor_id"`
	CostoUnidad        float64             `bson:"precio_unitario"`
	TiempoEstimadoDias int                 `bson:"tiempo_estimado_dias"`
	Auditoria          auditoria.Auditoria `bson:"auditoria"`
}
