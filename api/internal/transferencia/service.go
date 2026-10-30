package transferencia

import (
	alertastock "api/internal/alertaStock"
	"api/internal/auditoria"
	"api/internal/deposito"
	depositoproducto "api/internal/depositoProducto"
	movimientostock "api/internal/movimientoStock"
	"api/internal/producto"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	evaluacionstock "api/internal/evaluacionStock"
)

const MaxTamanioPagina = 100

var (
	ErrMismoDeposito            = errors.New("el depósito de origen y destino no pueden ser el mismo")
	ErrNoAutorizadoDeposito     = errors.New("no tenés autorización para operar sobre este depósito")
	ErrStockInsuficiente        = errors.New("el depósito de origen no posee stock suficiente")
	ErrEstadoInvalido           = errors.New("la transferencia no se encuentra en el estado requerido para esta acción")
	ErrTransferenciaInexistente = errors.New("la transferencia solicitada no existe")
	ErrDepositoInexistente      = errors.New("el depósito indicado no existe")
)

type Service struct {
	repo                    Repositorio
	depositoStockRepo       depositoproducto.DepositoProductoRepositorio
	depositoRepo            deposito.DepositoRepositorio
	productoRepo            producto.ProductoRepositorio
	alertaStockRepo         alertastock.AlertaStockRepositorio
	evaluadorStock          evaluacionstock.EvaluadorStock
	movimientoRepo          movimientostock.MovimientoStockRepositorio
	generarSugerenciasStock func(context.Context, string, string)
}

func NuevoService(
	repo Repositorio,
	stockRepo depositoproducto.DepositoProductoRepositorio,
	depositoRepo deposito.DepositoRepositorio,
	productoRepo producto.ProductoRepositorio,
	alertaRepo alertastock.AlertaStockRepositorio,
	evaluadorStock evaluacionstock.EvaluadorStock,
	movRepo movimientostock.MovimientoStockRepositorio,
	generarSugerenciasStock func(context.Context, string, string),
) *Service {
	return &Service{
		repo:                    repo,
		depositoStockRepo:       stockRepo,
		depositoRepo:            depositoRepo,
		productoRepo:            productoRepo,
		alertaStockRepo:         alertaRepo,
		evaluadorStock:          evaluadorStock,
		movimientoRepo:          movRepo,
		generarSugerenciasStock: generarSugerenciasStock,
	}
}

// Solicitar crea la solicitud de transferencia validando origen, destino y stock preventivo.
func (s *Service) Solicitar(
	ctx context.Context,
	usuarioID,
	rol,
	depUsuario string,
	dto SolicitarTransferenciaDTO,
) (Transferencia, error) {

	if dto.DepositoOrigenID == dto.DepositoDestinoID {
		return Transferencia{}, ErrMismoDeposito
	}

	// Sólo se pueden solicitar transferencias desde el depósito de origen.
	// En él, solo pueden hacerlo los gerentes u operarios.
	if depUsuario != dto.DepositoOrigenID {
		return Transferencia{}, ErrNoAutorizadoDeposito
	}

	// Validar que tanto el depósito de origen como el de destino existan
	// y no estén dados de baja antes de crear la solicitud.
	if _, err := s.depositoRepo.EncontrarPorId(ctx, dto.DepositoOrigenID); err != nil {
		if errors.Is(err, deposito.ErrDepositoNoEncontrado) {
			return Transferencia{}, ErrDepositoInexistente
		}
		return Transferencia{}, err
	}

	if _, err := s.depositoRepo.EncontrarPorId(ctx, dto.DepositoDestinoID); err != nil {
		if errors.Is(err, deposito.ErrDepositoNoEncontrado) {
			return Transferencia{}, ErrDepositoInexistente
		}
		return Transferencia{}, err
	}

	origOID, err := bson.ObjectIDFromHex(dto.DepositoOrigenID)
	if err != nil {
		return Transferencia{}, err
	}

	prodOID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return Transferencia{}, err
	}

	// Verificar stock preventivo en origen.
	dp, err := s.depositoStockRepo.EncontrarPorDepositoYProducto(
		ctx,
		origOID,
		prodOID,
	)

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

// Aprobar cambia atómicamente el estado a 'aprobada' y descuenta el stock de origen.
func (s *Service) Aprobar(
	ctx context.Context,
	idTransferencia,
	usuarioID,
	rol,
	depUsuario string,
) error {

	t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
	if err != nil {
		return ErrTransferenciaInexistente
	}

	if t.Estado != TransferenciaSolicitada {
		return ErrEstadoInvalido
	}

	// El gerente solamente puede aprobar transferencias destinadas a su depósito.
	// El administrador puede aprobar transferencias de cualquier depósito.
	if rol != "administrador" &&
		depUsuario != t.DepositoDestinoID.Hex() {
		return ErrNoAutorizadoDeposito
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return err
	}

	// Obtener el nombre del depósito de destino.
	depositoDestino, err := s.depositoRepo.EncontrarPorId(
		ctx,
		t.DepositoDestinoID.Hex(),
	)
	if err != nil {
		return err
	}

	ahora := time.Now()

	if err := s.repo.ActualizarEstado(
		ctx,
		idTransferencia,
		TransferenciaSolicitada,
		TransferenciaAprobada,
		&uOID,
		&ahora,
		usuarioID,
		"",
	); err != nil {
		return err
	}

	// Descontar stock del depósito de origen.
	_, err = s.depositoStockRepo.AjustarStock(
		ctx,
		t.DepositoOrigenID,
		t.ProductoID,
		-t.Cantidad,
		uOID,
	)
	if err != nil {
		s.revertirEstado(
			ctx,
			idTransferencia,
			TransferenciaAprobada,
			TransferenciaSolicitada,
			usuarioID,
		)
		return ErrStockInsuficiente
	}

	// Registrar movimiento de salida por transferencia.
	mov := movimientostock.MovimientoStock{
		ProductoID:           t.ProductoID,
		DepositoID:           t.DepositoOrigenID,
		Cantidad:             -t.Cantidad,
		Tipo:                 movimientostock.Transferencia,
		UsuarioResponsableID: uOID,
		FechaHora:            time.Now(),
		Motivo: "Transferencia interdepósito aprobada hacia: " +
			depositoDestino.Nombre,
		Auditoria: auditoria.Auditoria{
			CreadoPor: uOID,
			CreadoEn:  time.Now(),
		},
	}

	if _, err = s.movimientoRepo.Crear(ctx, mov); err != nil {
		s.revertirStock(
			ctx,
			t.DepositoOrigenID,
			t.ProductoID,
			t.Cantidad,
			uOID,
		)
		s.revertirEstado(
			ctx,
			idTransferencia,
			TransferenciaAprobada,
			TransferenciaSolicitada,
			usuarioID,
		)
		return err
	}

	// Evaluar alertas del depósito de origen.
	stockBajo, err := s.evaluarAlertas(
		ctx,
		t.DepositoOrigenID,
		t.ProductoID,
		uOID,
	)
	if err != nil {
		return err
	}

	// Generar sugerencias de compra si el stock quedó por debajo del mínimo.
	if stockBajo && s.generarSugerenciasStock != nil {
		s.generarSugerenciasStock(
			ctx,
			t.DepositoOrigenID.Hex(),
			usuarioID,
		)
	}

	return nil
}

// revertirStock deshace un AjustarStock previo aplicando el delta inverso.
// Se usa como compensación cuando falla un paso posterior de la operación
// (no hay transacciones). Si la propia reversión falla, se loguea como
// crítico porque stock y movimientos quedan inconsistentes.
func (s *Service) revertirStock(
	ctx context.Context,
	depositoID, productoID bson.ObjectID,
	delta int32,
	usuarioOID bson.ObjectID,
) {
	if _, err := s.depositoStockRepo.AjustarStock(ctx, depositoID, productoID, delta, usuarioOID); err != nil {
		// Se conserva el registro del error crítico de compensación.
		// Si tampoco se desea registrar estos errores, se puede quitar este bloque.
		panic(err)
	}
}

// revertirEstado restaura el estado anterior de una transferencia cuando un
// paso posterior al guard de estado falla.
func (s *Service) revertirEstado(
	ctx context.Context,
	idTransferencia string,
	estadoActual, estadoAnterior EstadoTransferencia,
	usuarioID string,
) {
	if err := s.repo.ActualizarEstado(
		ctx,
		idTransferencia,
		estadoActual,
		estadoAnterior,
		nil,
		nil,
		usuarioID,
		"",
	); err != nil {
		panic(err)
	}
}

// Rechazar rechaza la solicitud sin afectar stock.
func (s *Service) Rechazar(
	ctx context.Context,
	idTransferencia,
	usuarioID,
	rol,
	depUsuario,
	motivo string,
) error {

	t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
	if err != nil {
		return ErrTransferenciaInexistente
	}

	if t.Estado != TransferenciaSolicitada {
		return ErrEstadoInvalido
	}

	if rol != "administrador" &&
		depUsuario != t.DepositoDestinoID.Hex() {
		return ErrNoAutorizadoDeposito
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return err
	}

	ahora := time.Now()

	return s.repo.ActualizarEstado(
		ctx,
		idTransferencia,
		TransferenciaSolicitada,
		TransferenciaRechazada,
		&uOID,
		&ahora,
		usuarioID,
		motivo,
	)
}

// Completar cambia atómicamente el estado a 'completada' y acredita el stock en destino.
func (s *Service) Completar(
	ctx context.Context,
	idTransferencia,
	usuarioID,
	rol,
	depUsuario string,
) error {

	t, err := s.repo.EncontrarPorID(ctx, idTransferencia)
	if err != nil {
		return ErrTransferenciaInexistente
	}

	if t.Estado != TransferenciaAprobada {
		return ErrEstadoInvalido
	}

	// El personal del depósito destino (operario o gerente) puede completar transferencias destinadas a su depósito.
	// El administrador puede completar transferencias de cualquier depósito.
	if rol != "administrador" &&
		depUsuario != t.DepositoDestinoID.Hex() {
		return ErrNoAutorizadoDeposito
	}

	uOID, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return err
	}

	// Obtener el nombre del depósito de origen.
	depositoOrigen, err := s.depositoRepo.EncontrarPorId(
		ctx,
		t.DepositoOrigenID.Hex(),
	)
	if err != nil {
		return err
	}

	// Cambiar el estado de forma atómica.
	if err := s.repo.ActualizarEstado(
		ctx,
		idTransferencia,
		TransferenciaAprobada,
		TransferenciaCompletada,
		nil,
		nil,
		usuarioID,
		"",
	); err != nil {
		return err
	}

	// Asegurar que exista la fila de stock en el depósito destino.
	_, err = s.depositoStockRepo.AsegurarExiste(
		ctx,
		t.DepositoDestinoID,
		t.ProductoID,
		uOID,
	)
	if err != nil {
		s.revertirEstado(
			ctx,
			idTransferencia,
			TransferenciaCompletada,
			TransferenciaAprobada,
			usuarioID,
		)
		return err
	}

	// Acreditar stock en el depósito destino.
	_, err = s.depositoStockRepo.AjustarStock(
		ctx,
		t.DepositoDestinoID,
		t.ProductoID,
		t.Cantidad,
		uOID,
	)
	if err != nil {
		s.revertirEstado(
			ctx,
			idTransferencia,
			TransferenciaCompletada,
			TransferenciaAprobada,
			usuarioID,
		)
		return err
	}

	// Registrar movimiento de entrada por transferencia.
	mov := movimientostock.MovimientoStock{
		ProductoID:           t.ProductoID,
		DepositoID:           t.DepositoDestinoID,
		Cantidad:             t.Cantidad,
		Tipo:                 movimientostock.Transferencia,
		UsuarioResponsableID: uOID,
		FechaHora:            time.Now(),
		Motivo: "Recepción de mercadería transferida desde: " +
			depositoOrigen.Nombre,
		Auditoria: auditoria.Auditoria{
			CreadoPor: uOID,
			CreadoEn:  time.Now(),
		},
	}

	if _, err = s.movimientoRepo.Crear(ctx, mov); err != nil {
		s.revertirStock(
			ctx,
			t.DepositoDestinoID,
			t.ProductoID,
			-t.Cantidad,
			uOID,
		)
		s.revertirEstado(
			ctx,
			idTransferencia,
			TransferenciaCompletada,
			TransferenciaAprobada,
			usuarioID,
		)
		return err
	}

	// Evaluar alertas del depósito de destino.
	if _, err := s.evaluarAlertas(
		ctx,
		t.DepositoDestinoID,
		t.ProductoID,
		uOID,
	); err != nil {
		return err
	}

	return nil
}

func (s *Service) Listar(
	ctx context.Context,
	rol,
	depUsuario,
	usuarioID string,
	filtro FiltroTransferenciasDTO,
) (TransferenciasPaginadasDTO, error) {

	depositoPermitido := ""

	if rol == "operario" || rol == "gerente" {
		depositoPermitido = depUsuario
	}

	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 20
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}
	idSolicitante := ""

	if rol == "operario" {
		idSolicitante = usuarioID
	}

	items, total, err := s.repo.Listar(
		ctx,
		filtro,
		depositoPermitido,
		idSolicitante,
	)

	if err != nil {
		return TransferenciasPaginadasDTO{}, err
	}

	pagina := filtro.Pagina
	if pagina < 1 {
		pagina = 1
	}

	tamanioPagina := filtro.TamanioPagina

	dtos := make([]TransferenciaDTO, 0, len(items))

	for _, it := range items {
		dtos = append(dtos, it.ConvertirADTO())
	}

	return TransferenciasPaginadasDTO{
		Items:         dtos,
		Total:         total,
		Pagina:        pagina,
		TamanioPagina: tamanioPagina,
	}, nil
}

func (s *Service) ObtenerPorID(
	ctx context.Context,
	id,
	rol,
	depUsuario string,
) (Transferencia, error) {

	t, err := s.repo.EncontrarPorID(ctx, id)
	if err != nil {
		return Transferencia{}, ErrTransferenciaInexistente
	}

	if (rol == "operario" || rol == "gerente") &&
		depUsuario != t.DepositoOrigenID.Hex() &&
		depUsuario != t.DepositoDestinoID.Hex() {
		return Transferencia{}, ErrNoAutorizadoDeposito
	}

	return t, nil
}

const diasAvisoVencimiento = 30

// evaluarAlertas evalúa las alertas del producto en un depósito determinado.
func (s *Service) evaluarAlertas(
	ctx context.Context,
	depositoID bson.ObjectID,
	productoID bson.ObjectID,
	usuarioID bson.ObjectID,
) (bool, error) {

	prod, err := s.productoRepo.EncontrarPorId(ctx, productoID.Hex())
	if err != nil {
		return false, err
	}

	stockFila, err := s.depositoStockRepo.EncontrarPorDepositoYProducto(
		ctx,
		depositoID,
		productoID,
	)

	var stockActual int32
	minimo := prod.StockMinimo

	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return false, err
		}
	} else {
		stockActual = stockFila.Stock

		if stockFila.StockMinimo != nil {
			minimo = *stockFila.StockMinimo
		}
	}

	ahora := time.Now()
	stockBajo := stockActual < minimo

	// Evaluar stock mínimo.
	if err := s.evaluadorStock.EvaluarStock(
		ctx,
		depositoID,
		productoID,
		stockActual,
		minimo,
		usuarioID,
	); err != nil {
		return stockBajo, err
	}

	// Evaluar vencimiento independientemente del stock mínimo.
	if prod.FechaVencimiento != nil && stockActual > 0 {
		limite := ahora.AddDate(0, 0, diasAvisoVencimiento)

		if !prod.FechaVencimiento.Before(ahora) &&
			!prod.FechaVencimiento.After(limite) {

			alertaExistente, err := s.alertaStockRepo.BuscarActiva(
				ctx,
				depositoID,
				productoID,
				alertastock.AlertaVencimiento,
			)

			if errors.Is(err, mongo.ErrNoDocuments) {
				fechaVencimiento := *prod.FechaVencimiento

				_, err = s.alertaStockRepo.Crear(ctx, alertastock.AlertaStock{
					Tipo:             alertastock.AlertaVencimiento,
					ProductoID:       productoID,
					DepositoID:       depositoID,
					StockActual:      stockActual,
					FechaVencimiento: &fechaVencimiento,
					Resuelta:         false,
					Auditoria: auditoria.Auditoria{
						CreadoPor: usuarioID,
						CreadoEn:  ahora,
					},
				})

				if err != nil {
					return stockBajo, err
				}
			} else if err != nil {
				return stockBajo, err
			} else {
				_ = s.alertaStockRepo.ActualizarVencimiento(
					ctx,
					alertaExistente.ID.Hex(),
					*prod.FechaVencimiento,
					stockActual,
					usuarioID,
				)
			}
		}
	}

	return stockBajo, nil
}
