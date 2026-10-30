package alertastock

import (
	"time"

	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TipoAlerta string

const (
	AlertaStockBajo   TipoAlerta = "stock_bajo"
	AlertaVencimiento TipoAlerta = "vencimiento"
)

// AlertaStock representa una alerta generada automáticamente por el sistema.
// Puede corresponder a stock por debajo del mínimo o a un producto próximo
// a vencer.
type AlertaStock struct {
	ID               bson.ObjectID       `bson:"_id,omitempty"`
	Tipo             TipoAlerta          `bson:"tipo"`
	ProductoID       bson.ObjectID       `bson:"producto_id"`
	DepositoID       bson.ObjectID       `bson:"deposito_id"`
	StockMinimo      int32               `bson:"stock_minimo"`
	StockActual      int32               `bson:"stock_actual"`
	FechaVencimiento *time.Time          `bson:"fecha_vencimiento,omitempty"`
	Resuelta         bool                `bson:"resuelta"`
	Auditoria        auditoria.Auditoria `bson:"auditoria"`

	ProductoNombre string `bson:"producto_nombre,omitempty"`
	DepositoNombre string `bson:"deposito_nombre,omitempty"`
}

// ConvertirADTO transforma una AlertaStock a su DTO enriquecido. Recibe los
// nombres de producto y depósito como parámetros porque el modelo solo guarda
// IDs — la resolución a nombres es responsabilidad del service que tiene
// acceso a los repositorios de producto y depósito.
func ConvertirADTO(a AlertaStock) AlertaStockDTO {
	tipo := a.Tipo
	if tipo == "" {
		tipo = AlertaStockBajo
	}

	dto := AlertaStockDTO{
		ID:             a.ID.Hex(),
		Tipo:           tipo,
		ProductoID:     a.ProductoID.Hex(),
		ProductoNombre: a.ProductoNombre,
		DepositoID:     a.DepositoID.Hex(),
		DepositoNombre: a.DepositoNombre,
		StockMinimo:    a.StockMinimo,
		StockActual:    a.StockActual,
		Resuelta:       a.Resuelta,
		FechaCreacion:  a.Auditoria.CreadoEn.Format(time.RFC3339),
	}

	if a.FechaVencimiento != nil {
		fecha := a.FechaVencimiento.Format(time.RFC3339)
		dto.FechaVencimiento = &fecha
	}

	if a.Resuelta && a.Auditoria.ActualizadoEn != nil {
		fecha := a.Auditoria.ActualizadoEn.Format(time.RFC3339)
		dto.FechaResolucion = &fecha
	}

	if a.Resuelta && a.Auditoria.ModificadoPor != nil {
		mod := a.Auditoria.ModificadoPor.Hex()
		dto.ModificadoPor = &mod
	}

	return dto
}
