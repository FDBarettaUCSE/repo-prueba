package movimientostock

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// RegistrarMovimientoDTO contiene los datos requeridos para registrar un ingreso, egreso o ajuste
type RegistrarMovimientoDTO struct {
	ProductoID string         `json:"producto_id" binding:"required,len=24,hexadecimal"`
	DepositoID string         `json:"deposito_id" binding:"required,len=24,hexadecimal"`
	Cantidad   int32          `json:"cantidad" binding:"required,gt=0"`
	Tipo       TipoMovimiento `json:"tipo" binding:"required,oneof=ingreso egreso ajuste"`
	Motivo     string         `json:"motivo" binding:"required"`
}

// FiltroMovimientosDTO agrupa los parámetros de búsqueda del historial de movimientos
type FiltroMovimientosDTO struct {
	DepositoID string `form:"deposito_id"`
	ProductoID string `form:"producto_id"`
	Tipo       string `form:"tipo"`
	FechaDesde string `form:"fecha_desde"`
	FechaHasta string `form:"fecha_hasta"`
}

// MovimientoStockDTO representa el registro inmutable de auditoría enviado al cliente
type MovimientoStockDTO struct {
	ID                   string         `json:"id"`
	ProductoID           string         `json:"producto_id"`
	DepositoID           string         `json:"deposito_id"`
	Cantidad             int32          `json:"cantidad"`
	Tipo                 TipoMovimiento `json:"tipo"`
	UsuarioResponsableID string         `json:"usuario_responsable_id"`
	FechaHora            string         `json:"fecha_hora"`
	Motivo               string         `json:"motivo"`
}

// ConvertirADTO transforma la entidad MovimientoStock a su DTO de salida
func (m MovimientoStock) ConvertirADTO() MovimientoStockDTO {
	return MovimientoStockDTO{
		ID:                   m.ID.Hex(),
		ProductoID:           m.ProductoID.Hex(),
		DepositoID:           m.DepositoID.Hex(),
		Cantidad:             m.Cantidad,
		Tipo:                 m.Tipo,
		UsuarioResponsableID: m.UsuarioResponsableID.Hex(),
		FechaHora:            m.FechaHora.Format(time.RFC3339),
		Motivo:               m.Motivo,
	}
}

// ConvertirAModelo transforma el DTO de registro a la entidad MovimientoStock
func (dto RegistrarMovimientoDTO) ConvertirAModelo(usuarioID string) (MovimientoStock, error) {
	prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return MovimientoStock{}, err
	}
	depoOID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return MovimientoStock{}, err
	}
	usrOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return MovimientoStock{}, err
	}

	return MovimientoStock{
		ProductoID:           prodOID,
		DepositoID:           depoOID,
		Cantidad:             dto.Cantidad,
		Tipo:                 dto.Tipo,
		UsuarioResponsableID: usrOID,
		FechaHora:            time.Now(),
		Motivo:               dto.Motivo,
	}, nil
}
