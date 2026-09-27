package transferencia

import (
	"context"
	"errors"
	"time"

	"api/internal/auditoria"
	depositoproducto "api/internal/depositoProducto"
	movimientostock "api/internal/movimientoStock"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrMismoDeposito           = errors.New("el depósito de origen y destino no pueden ser el mismo")
	ErrNoAutorizadoDeposito    = errors.New("no tenés autorización para operar sobre este depósito")
	ErrStockInsuficiente       = errors.New("el depósito de origen no posee stock suficiente")
	ErrEstadoInvalido          = errors.New("la transferencia no se encuentra en el estado requerido para esta acción")
	ErrSoloGerenteDestinoAdmin = errors.New("solo el gerente del depósito destino o un administrador pueden resolver esta transferencia")
)

type Service struct {
	repo              Repositorio
	depositoStockRepo depositoproducto.DepositoProductoRepositorio
	movimientoRepo    movimientostock.MovimientoStockRepositorio
}

func NuevoService(repo Repositorio, stockRepo depositoproducto.DepositoProductoRepositorio, movRepo movimientostock.MovimientoStockRepositorio) *Service {
	return &Service{
		repo:              repo,
		depositoStockRepo: stockRepo,
		movimientoRepo:    movRepo,
	}
}

// Solicitar crea la solicitud de transferencia validando origen y stock preventivo
func (s *Service) Solicitar(ctx context.Context, usuarioID, rol, depUsuario string, dto SolicitarTransferenciaDTO) (Transferencia, error) {
	if dto.DepositoOrigenID == dto.DepositoDestinoID {
		return Transferencia{}, ErrMismoDeposito
	}

	// Operario y Gerente solo pueden solicitar desde su propio depósito
	if (rol == "operario" || rol == "gerente") && depUsuario != dto.DepositoOrigenID {
		return Transferencia{}, ErrNoAutorizadoDeposito
	}

	origOID, err := bson.ObjectIDFromHex(dto.DepositoOrigenID)
	if err != nil {
		return Transferencia{}, err
	}
	prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return Transferencia{}, err
	}

	// Verificar stock preventivo en origen
	dp, err := s.depositoStockRepo.EncontrarPorDepositoYProducto(ctx, origOID, prodOID)
	if err != nil || dp.Stock < dto.Cantidad {
		return Transferencia{}, ErrStockInsuficiente
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return Transferencia{}, err
	}

	t, err := dto.ConvertirAModelo(usuarioID)
	if err != nil {
		return Transferencia{}, err
	}
	t.Auditoria = auditoria.Auditoria{
		CreadoPor: uOID,
		CreadoEn:  time.Now(),
	}

	return s.repo.Crear(ctx, t)
}

// Aprobar descuenta el stock de origen y cambia el estado a 'aprobada'
func (s *Service) Aprobar(ctx context.Context, idTransferencia, usuarioID, rol, depUsuario string) error {
	t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
	if err != nil {
		return err
	}
	if t.Estado != TransferenciaSolicitada {
		return ErrEstadoInvalido
	}

	// Solo el Gerente del depósito destino o el Administrador pueden aprobar
	if rol != "administrador" && (rol != "gerente" || depUsuario != t.DepositoDestinoID.Hex()) {
		return ErrSoloGerenteDestinoAdmin
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return err
	}

	// 1. Descontar stock del depósito de origen (delta negativo)
	_, err = s.depositoStockRepo.AjustarStock(ctx, t.DepositoOrigenID, t.ProductoID, -t.Cantidad, uOID)
	if err != nil {
		return ErrStockInsuficiente
	}

	// 2. Registrar movimiento inmutable de salida por transferencia
	mov := movimientostock.MovimientoStock{
		ProductoID:           t.ProductoID,
		DepositoID:           t.DepositoOrigenID,
		Cantidad:             t.Cantidad,
		Tipo:                 movimientostock.Transferencia,
		UsuarioResponsableID: uOID,
		FechaHora:            time.Now(),
		Motivo:               "Transferencia interdepósito aprobada hacia: " + t.DepositoDestinoID.Hex(),
		Auditoria: auditoria.Auditoria{
			CreadoPor: uOID,
			CreadoEn:  time.Now(),
		},
	}
	if _, err = s.movimientoRepo.Crear(ctx, mov); err != nil {
		return err
	}

	// 3. Actualizar estado de la transferencia
	ahora := time.Now()
	return s.repo.ActualizarEstado(ctx, idTransferencia, TransferenciaAprobada, &uOID, &ahora, usuarioID, "")
}

// Rechazar rechaza la solicitud sin afectar stock
func (s *Service) Rechazar(ctx context.Context, idTransferencia, usuarioID, rol, depUsuario, motivo string) error {
	t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
	if err != nil {
		return err
	}
	if t.Estado != TransferenciaSolicitada {
		return ErrEstadoInvalido
	}

	if rol != "administrador" && (rol != "gerente" || depUsuario != t.DepositoDestinoID.Hex()) {
		return ErrSoloGerenteDestinoAdmin
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return err
	}
	ahora := time.Now()
	return s.repo.ActualizarEstado(ctx, idTransferencia, TransferenciaRechazada, &uOID, &ahora, usuarioID, motivo)
}

// Completar acredita el stock en destino y finaliza el ciclo
func (s *Service) Completar(ctx context.Context, idTransferencia, usuarioID, rol, depUsuario string) error {
	t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
	if err != nil {
		return err
	}
	if t.Estado != TransferenciaAprobada {
		return ErrEstadoInvalido
	}

	// La recepción la confirma el personal del depósito destino o admin
	if rol != "administrador" && depUsuario != t.DepositoDestinoID.Hex() {
		return ErrNoAutorizadoDeposito
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return err
	}

	// 1. Acreditar stock en el depósito de destino (delta positivo)
	_, err = s.depositoStockRepo.AjustarStock(ctx, t.DepositoDestinoID, t.ProductoID, t.Cantidad, uOID)
	if err != nil {
		return err
	}

	// 2. Registrar movimiento inmutable de entrada por transferencia
	mov := movimientostock.MovimientoStock{
		ProductoID:           t.ProductoID,
		DepositoID:           t.DepositoDestinoID,
		Cantidad:             t.Cantidad,
		Tipo:                 movimientostock.Transferencia,
		UsuarioResponsableID: uOID,
		FechaHora:            time.Now(),
		Motivo:               "Recepción de mercadería transferida desde: " + t.DepositoOrigenID.Hex(),
		Auditoria: auditoria.Auditoria{
			CreadoPor: uOID,
			CreadoEn:  time.Now(),
		},
	}
	if _, err = s.movimientoRepo.Crear(ctx, mov); err != nil {
		return err
	}

	// 3. Marcar completada
	return s.repo.ActualizarEstado(ctx, idTransferencia, TransferenciaCompletada, nil, nil, usuarioID, "")
}

func (s *Service) Listar(ctx context.Context, rol, depUsuario string, filtro FiltroTransferenciasDTO) ([]Transferencia, error) {
	depositoPermitido := ""
	if rol == "operario" || rol == "gerente" {
		depositoPermitido = depUsuario
	}
	return s.repo.Listar(ctx, filtro, depositoPermitido)
}

func (s *Service) ObtenerPorID(ctx context.Context, id, rol, depUsuario string) (Transferencia, error) {
	t, err := s.repo.EncontrarPorID(ctx, id)
	if err != nil {
		return Transferencia{}, err
	}

	if (rol == "operario" || rol == "gerente") && depUsuario != t.DepositoOrigenID.Hex() && depUsuario != t.DepositoDestinoID.Hex() {
		return Transferencia{}, ErrNoAutorizadoDeposito
	}

	return t, nil
}
