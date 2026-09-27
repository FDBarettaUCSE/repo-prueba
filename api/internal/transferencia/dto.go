package transferencia

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SolicitarTransferenciaDTO contiene los datos requeridos para iniciar una solicitud de transferencia
type SolicitarTransferenciaDTO struct {
	DepositoOrigenID  string `json:"deposito_origen_id" binding:"required,len=24,hexadecimal"`
	DepositoDestinoID string `json:"deposito_destino_id" binding:"required,len=24,hexadecimal"`
	ProductoID        string `json:"producto_id" binding:"required,len=24,hexadecimal"`
	Cantidad          int32  `json:"cantidad" binding:"required,gt=0"`
	Motivo            string `json:"motivo" binding:"required"`
}

// ResolverTransferenciaDTO contiene la acción del responsable de destino (aprobar o rechazar)
type ResolverTransferenciaDTO struct {
	Accion string `json:"accion" binding:"required,oneof=aprobar rechazar"`
	Motivo string `json:"motivo,omitempty"`
}

// CompletarTransferenciaDTO contiene datos opcionales al confirmar la llegada de mercadería
type CompletarTransferenciaDTO struct {
	Observaciones string `json:"observaciones,omitempty"`
}

// FiltroTransferenciasDTO agrupa parámetros de búsqueda de transferencias
type FiltroTransferenciasDTO struct {
	DepositoID string `form:"deposito_id"`
	Estado     string `form:"estado"`
}

// TransferenciaDTO representa la transferencia formateada para la respuesta al cliente
type TransferenciaDTO struct {
	ID                   string              `json:"id"`
	DepositoOrigenID     string              `json:"deposito_origen_id"`
	DepositoDestinoID    string              `json:"deposito_destino_id"`
	ProductoID           string              `json:"producto_id"`
	Cantidad             int32               `json:"cantidad"`
	Estado               EstadoTransferencia `json:"estado"`
	UsuarioSolicitanteID string              `json:"usuario_solicitante_id"`
	UsuarioResolucionID  string              `json:"usuario_resolucion_id,omitempty"`
	FechaSolicitud       string              `json:"fecha_solicitud"`
	FechaResolucion      string              `json:"fecha_resolucion,omitempty"`
	Motivo               string              `json:"motivo,omitempty"`
}

// ConvertirADTO transforma la entidad Transferencia a su DTO de salida
func (t Transferencia) ConvertirADTO() TransferenciaDTO {
	dto := TransferenciaDTO{
		ID:                   t.ID.Hex(),
		DepositoOrigenID:     t.DepositoOrigenID.Hex(),
		DepositoDestinoID:    t.DepositoDestinoID.Hex(),
		ProductoID:           t.ProductoID.Hex(),
		Cantidad:             t.Cantidad,
		Estado:               t.Estado,
		UsuarioSolicitanteID: t.UsuarioSolicitanteID.Hex(),
		FechaSolicitud:       t.FechaSolicitud.Format(time.RFC3339),
		Motivo:               t.Motivo,
	}
	if t.UsuarioResolucionID != nil && !t.UsuarioResolucionID.IsZero() {
		dto.UsuarioResolucionID = t.UsuarioResolucionID.Hex()
	}
	if t.FechaResolucion != nil && !t.FechaResolucion.IsZero() {
		dto.FechaResolucion = t.FechaResolucion.Format(time.RFC3339)
	}
	return dto
}

// ConvertirAModelo transforma el DTO de solicitud al modelo Transferencia
func (dto SolicitarTransferenciaDTO) ConvertirAModelo(solicitanteID string) (Transferencia, error) {
	origOID, err := bson.ObjectIDFromHex(dto.DepositoOrigenID)
	if err != nil {
		return Transferencia{}, err
	}
	destOID, err := bson.ObjectIDFromHex(dto.DepositoDestinoID)
	if err != nil {
		return Transferencia{}, err
	}
	prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return Transferencia{}, err
	}
	usrOID, err := bson.ObjectIDFromHex(solicitanteID)
	if err != nil {
		return Transferencia{}, err
	}

	return Transferencia{
		DepositoOrigenID:     origOID,
		DepositoDestinoID:    destOID,
		ProductoID:           prodOID,
		Cantidad:             dto.Cantidad,
		Estado:               TransferenciaSolicitada,
		UsuarioSolicitanteID: usrOID,
		FechaSolicitud:       time.Now(),
		Motivo:               dto.Motivo,
	}, nil
}
