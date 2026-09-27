package ordendecompra

import (
	"api/internal/auditoria"
	productoproveedor "api/internal/productoProveedor"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type EstadoOrden string

const (
	OrdenBorrador   EstadoOrden = "borrador"
	OrdenConfirmada EstadoOrden = "confirmada"
	OrdenParcial    EstadoOrden = "parcial"
	OrdenCompletada EstadoOrden = "completada"
)

type ItemOrden struct {
	ProductoProveedor productoproveedor.ProductoProveedor `bson:"producto_proveedor"`
	Cantidad          int                                 `bson:"cantidad"`
}

type OrdenDeCompra struct {
	ID          bson.ObjectID       `bson:"_id,omitempty"`
	DepositoID  bson.ObjectID       `bson:"deposito_id"`
	Items       []ItemOrden         `bson:"items"`
	EstadoOrden EstadoOrden         `bson:"estado_orden"`
	Auditoria   auditoria.Auditoria `bson:"auditoria"`
}
