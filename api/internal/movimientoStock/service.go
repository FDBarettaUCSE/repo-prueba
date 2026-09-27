package movimientostock

import (
	"context"
	"errors"
	"time"

	"api/internal/auditoria"
	depositoproducto "api/internal/depositoProducto"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrStockInsuficiente       = errors.New("el stock disponible es insuficiente para realizar esta operación")
	ErrNoAutorizadoDeposito    = errors.New("no tenés autorización para operar sobre este depósito")
	ErrTipoMovimientoInvalido = errors.New("tipo de movimiento no admitido para registro manual")
)

type Service struct {
	repo              MovimientoStockRepositorio
	depositoStockRepo depositoproducto.DepositoProductoRepositorio
}

func NuevoService(repo MovimientoStockRepositorio, stockRepo depositoproducto.DepositoProductoRepositorio) *Service {
	return &Service{
		repo:              repo,
		depositoStockRepo: stockRepo,
	}
}

// Registrar aplica el cambio en el inventario y crea el registro inmutable de auditoría
func (s *Service) Registrar(ctx context.Context, usuarioID, rol, depUsuario string, dto RegistrarMovimientoDTO) (MovimientoStock, error) {
	// Operarios y Gerentes solo pueden operar en su depósito asignado
	if (rol == "operario" || rol == "gerente") && depUsuario != dto.DepositoID {
		return MovimientoStock{}, ErrNoAutorizadoDeposito
	}

	depoOID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return MovimientoStock{}, err
	}
	prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return MovimientoStock{}, err
	}
	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return MovimientoStock{}, err
	}

	var delta int32
	switch dto.Tipo {
	case Ingreso:
		delta = dto.Cantidad
	case Egreso:
		delta = -dto.Cantidad
	case Ajuste:
		// Para ajuste directo, el delta aplica según el sentido de la corrección
		delta = dto.Cantidad
	default:
		return MovimientoStock{}, ErrTipoMovimientoInvalido
	}

	// 1. Modificar el stock atómicamente en depositoProducto
	_, err = s.depositoStockRepo.AjustarStock(ctx, depoOID, prodOID, delta, uOID)
	if err != nil {
		if errors.Is(err, depositoproducto.ErrStockInsuficiente) {
			return MovimientoStock{}, ErrStockInsuficiente
		}
		return MovimientoStock{}, err
	}

	// 2. Registrar el movimiento inmutable
	mov := MovimientoStock{
		ProductoID:           prodOID,
		DepositoID:           depoOID,
		Cantidad:             dto.Cantidad,
		Tipo:                 dto.Tipo,
		UsuarioResponsableID: uOID,
		FechaHora:            time.Now(),
		Motivo:               dto.Motivo,
		Auditoria: auditoria.Auditoria{
			CreadoPor: uOID,
			CreadoEn:  time.Now(),
		},
	}

	return s.repo.Crear(ctx, mov)
}

func (s *Service) Listar(ctx context.Context, rol, depUsuario string, filtro FiltroMovimientosDTO) ([]MovimientoStock, error) {
	depositoPermitido := ""
	if rol == "operario" || rol == "gerente" {
		depositoPermitido = depUsuario
	}
	return s.repo.Listar(ctx, filtro, depositoPermitido)
}

func (s *Service) ObtenerPorID(ctx context.Context, id, rol, depUsuario string) (MovimientoStock, error) {
	m, err := s.repo.EncontrarPorID(ctx, id)
	if err != nil {
		return MovimientoStock{}, err
	}

	if (rol == "operario" || rol == "gerente") && depUsuario != m.DepositoID.Hex() {
		return MovimientoStock{}, ErrNoAutorizadoDeposito
	}

	return m, nil
}
