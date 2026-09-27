package alertastock

import (
	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AlertaStock struct {
	ID          bson.ObjectID       `bson:"_id,omitempty"`
	ProductoID  bson.ObjectID       `bson:"producto_id"`
	DepositoID  bson.ObjectID       `bson:"deposito_id"`
	StockMinimo int                 `bson:"stock_minimo"`
	StockActual int                 `bson:"stock_actual"`
	Resuelta    bool                `bson:"resuelta"`
	Auditoria   auditoria.Auditoria `bson:"auditoria"`
}
