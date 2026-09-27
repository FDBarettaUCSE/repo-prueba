package depositoproducto

import (
	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DepositoProducto es la fila de stock de UN producto en UN depósito
// puntual (ver Deposito.RegistrarRutas y Producto: cada depósito mantiene su
// propio stock, independiente del de los demás). No hay una fila para cada
// combinación posible desde el principio: se crea recién con el primer
// movimiento de stock que la necesite (ver AsegurarExiste en repository.go).
//
// StockMinimo es un *int32 (puntero) a propósito: nil significa "este
// depósito NO tiene una configuración particular, usa el stock mínimo
// general del producto" (ver Producto.StockMinimo). Un valor no-nil,
// incluido *0, es un override explícito que reemplaza al general para este
// depósito puntual — esa es la única forma de distinguir "no configurado"
// de "configurado en cero".
type DepositoProducto struct {
	ID          bson.ObjectID       `bson:"_id,omitempty"`
	DepositoID  bson.ObjectID       `bson:"deposito_id"`
	ProductoID  bson.ObjectID       `bson:"producto_id"`
	Stock       int32               `bson:"stock"`
	StockMinimo *int32              `bson:"stock_minimo,omitempty"`
	Auditoria   auditoria.Auditoria `bson:"auditoria"`
}
