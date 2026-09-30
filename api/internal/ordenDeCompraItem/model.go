package ordendecompraitem

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrdenCompraItem struct {
	ID                 bson.ObjectID       `bson:"_id,omitempty"`
	OrdenCompraID      bson.ObjectID       `bson:"orden_compra_id"`
	ProductoID         bson.ObjectID       `bson:"producto_id"`
	CantidadSolicitada int32               `bson:"cantidad_solicitada"`
	CantidadRecibida   int32               `bson:"cantidad_recibida"`
	CostoUnitario      float64             `bson:"costo_unitario"`
	FechaEliminacion   *time.Time          `bson:"fecha_eliminacion,omitempty"`
	Auditoria          auditoria.Auditoria `bson:"auditoria"`
}

func MapeoAOrdenCompraItemDTO(item *OrdenCompraItem) OrdenCompraItemDTO {
	return OrdenCompraItemDTO{
		ID:                 item.ID.Hex(),
		OrdenCompraID:      item.OrdenCompraID.Hex(),
		ProductoID:         item.ProductoID.Hex(),
		CantidadSolicitada: item.CantidadSolicitada,
		CantidadRecibida:   item.CantidadRecibida,
		CostoUnitario:      item.CostoUnitario,
	}
}
