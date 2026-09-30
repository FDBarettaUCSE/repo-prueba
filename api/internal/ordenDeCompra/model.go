package ordendecompra

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EstadoOrdenCompra string

const (
	OrdenBorrador   EstadoOrdenCompra = "borrador"
	OrdenConfirmada EstadoOrdenCompra = "confirmada"
	OrdenParcial    EstadoOrdenCompra = "parcial"
	OrdenCompletada EstadoOrdenCompra = "completada"
	OrdenDescartada EstadoOrdenCompra = "descartada"
)

type OrdenCompra struct {
	ID               bson.ObjectID       `bson:"_id,omitempty"`
	ProveedorID      bson.ObjectID       `bson:"proveedor_id"`
	DepositoID       bson.ObjectID       `bson:"deposito_id"`
	Estado           EstadoOrdenCompra   `bson:"estado"`
	FechaEliminacion *time.Time          `bson:"fecha_eliminacion,omitempty"`
	Auditoria        auditoria.Auditoria `bson:"auditoria"`
}

func MapeoAOrdenCompraDTO(o OrdenCompra) OrdenCompraDTO {
	return OrdenCompraDTO{
		ID:          o.ID.Hex(),
		ProveedorID: o.ProveedorID.Hex(),
		DepositoID:  o.DepositoID.Hex(),
		Estado:      o.Estado,
	}
}

// Para la recepcion de productos para completar las ordenes de compra
type Recepcion struct {
	Items []RecepcionItem
}

type RecepcionItem struct {
	ItemID   bson.ObjectID
	Cantidad int32
}
