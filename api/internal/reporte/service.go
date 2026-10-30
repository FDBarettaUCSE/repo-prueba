package reporte

import (
	movimientostock "api/internal/movimientoStock"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrAccesoDenegado = errors.New(
		"no tenés permiso para operar sobre este depósito",
	)

	ErrDepositoInvalido = errors.New(
		"el identificador del depósito no es válido",
	)

	ErrCategoriaInvalida = errors.New(
		"el identificador de la categoría no es válido",
	)

	ErrTipoMovimientoInvalido = errors.New(
		"el tipo de movimiento ingresado no es válido",
	)

	ErrFechaDesdeInvalida = errors.New(
		"la fecha desde no tiene un formato válido, debe ser AAAA-MM-DD",
	)

	ErrFechaHastaInvalida = errors.New(
		"la fecha hasta no tiene un formato válido, debe ser AAAA-MM-DD",
	)

	ErrRangoFechasInvalido = errors.New(
		"la fecha desde debe ser anterior o igual a la fecha hasta",
	)

	ErrDepositoNoAsignado = errors.New(
		"el usuario no tiene un depósito asociado",
	)
)

type Service struct {
	repositorio ReporteRepositorio
}

func NuevoService(repositorio ReporteRepositorio) *Service {
	return &Service{
		repositorio: repositorio,
	}
}

func resolverAlcanceDeposito(
	rol string,
	depositoUsuario string,
	depositoPedido string,
) (string, error) {

	// Administrador y auditor pueden consultar cualquier depósito
	// o todos los depósitos si no especifican uno.
	if rol == "administrador" || rol == "auditor" {
		return depositoPedido, nil
	}

	// Gerente y operario necesitan tener un depósito asociado.
	if depositoUsuario == "" {
		return "", ErrDepositoNoAsignado
	}

	// Si especifican un depósito, debe ser el propio.
	if depositoPedido != "" && depositoPedido != depositoUsuario {
		return "", ErrAccesoDenegado
	}

	// Si no especificaron depósito, se limita automáticamente
	// al depósito del usuario.
	return depositoUsuario, nil
}

func convertirObjectID(
	id string,
	errInvalido error,
) (*bson.ObjectID, error) {

	if id == "" {
		return nil, nil
	}

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, errInvalido
	}

	return &oid, nil
}

// ==================== INVENTARIO ====================

func (s *Service) ReporteInventario(
	ctx context.Context,
	rol string,
	depositoUsuario string,
	filtro FiltroInventarioDTO,
) (ReporteInventarioDTO, error) {

	depositoID, err := resolverAlcanceDeposito(
		rol,
		depositoUsuario,
		filtro.DepositoID,
	)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	oidDeposito, err := convertirObjectID(
		depositoID,
		ErrDepositoInvalido,
	)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	oidCategoria, err := convertirObjectID(
		filtro.CategoriaID,
		ErrCategoriaInvalida,
	)
	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	filtroRepo := FiltroInventario{
		DepositoID:  oidDeposito,
		CategoriaID: oidCategoria,
	}

	porDeposito, totalDepositos, err :=
		s.repositorio.InventarioPorDeposito(
			ctx,
			filtroRepo,
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	porCategoria, totalCategorias, err :=
		s.repositorio.InventarioPorCategoria(
			ctx,
			filtroRepo,
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	stockTotal, valorizadoTotal, err :=
		s.repositorio.TotalesInventario(
			ctx,
			filtroRepo,
		)

	if err != nil {
		return ReporteInventarioDTO{}, err
	}

	respuesta := ReporteInventarioDTO{
		PorDeposito: make(
			[]ValorizacionDepositoDTO,
			0,
			len(porDeposito),
		),

		PaginacionDeposito: PaginacionAPaginaDTO(
			filtro.Pagina,
			totalDepositos,
		),

		PorCategoria: make(
			[]ValorizacionCategoriaDTO,
			0,
			len(porCategoria),
		),

		PaginacionCategoria: PaginacionAPaginaDTO(
			filtro.Pagina,
			totalCategorias,
		),

		StockTotal:      stockTotal,
		ValorizadoTotal: valorizadoTotal,
	}

	for _, r := range porDeposito {
		respuesta.PorDeposito = append(
			respuesta.PorDeposito,
			ValorizacionDepositoDTO{
				DepositoID:     r.DepositoID.Hex(),
				DepositoNombre: r.DepositoNombre,
				Stock:          r.Stock,
				Valorizado:     r.Valorizado,
			},
		)
	}

	for _, r := range porCategoria {
		respuesta.PorCategoria = append(
			respuesta.PorCategoria,
			ValorizacionCategoriaDTO{
				CategoriaID:     r.CategoriaID.Hex(),
				CategoriaNombre: r.CategoriaNombre,
				Stock:           r.Stock,
				Valorizado:      r.Valorizado,
			},
		)
	}

	return respuesta, nil
}

// ==================== MOVIMIENTOS ====================

func (s *Service) ReporteMovimientos(
	ctx context.Context,
	rol string,
	depositoUsuario string,
	filtro FiltroReporteMovimientosDTO,
) (ReporteMovimientosDTO, error) {

	ubicacion, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	depositoID, err := resolverAlcanceDeposito(
		rol,
		depositoUsuario,
		filtro.DepositoID,
	)
	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	oidDeposito, err := convertirObjectID(
		depositoID,
		ErrDepositoInvalido,
	)
	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	var tipoMovimiento *movimientostock.TipoMovimiento

	if filtro.Tipo != "" {
		tipo := movimientostock.TipoMovimiento(filtro.Tipo)

		switch tipo {
		case movimientostock.Ingreso,
			movimientostock.Egreso,
			movimientostock.Ajuste,
			movimientostock.Transferencia:

			tipoMovimiento = &tipo

		default:
			return ReporteMovimientosDTO{}, ErrTipoMovimientoInvalido
		}
	}

	var fechaDesde *time.Time
	var fechaHasta *time.Time

	if filtro.FechaDesde != "" {
		t, err := time.ParseInLocation(
			"2006-01-02",
			filtro.FechaDesde,
			ubicacion,
		)
		if err != nil {
			return ReporteMovimientosDTO{}, ErrFechaDesdeInvalida
		}

		fechaDesde = &t
	}

	if filtro.FechaHasta != "" {
		t, err := time.ParseInLocation(
			"2006-01-02",
			filtro.FechaHasta,
			ubicacion,
		)
		if err != nil {
			return ReporteMovimientosDTO{}, ErrFechaHastaInvalida
		}

		// Límite superior exclusivo para incluir todo el día.
		t = t.AddDate(0, 0, 1)

		fechaHasta = &t
	}

	if fechaDesde != nil &&
		fechaHasta != nil &&
		!fechaDesde.Before(*fechaHasta) {

		return ReporteMovimientosDTO{}, ErrRangoFechasInvalido
	}

	resultados, total, err :=
		s.repositorio.Movimientos(
			ctx,
			FiltroReporteMovimientos{
				DepositoID: oidDeposito,
				Tipo:       tipoMovimiento,
				FechaDesde: fechaDesde,
				FechaHasta: fechaHasta,
			},
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReporteMovimientosDTO{}, err
	}

	respuesta := ReporteMovimientosDTO{
		Resumen: make(
			[]ResumenMovimientoDTO,
			0,
			len(resultados),
		),

		TotalMovimientos: total,

		Paginacion: PaginacionAPaginaDTO(
			filtro.Pagina,
			total,
		),
	}

	for _, r := range resultados {
		respuesta.Resumen = append(
			respuesta.Resumen,
			ResumenMovimientoDTO{
				DepositoID:      r.DepositoID.Hex(),
				DepositoNombre:  r.DepositoNombre,
				Fecha:           r.Fecha,
				Tipo:            r.Tipo,
				CantidadEventos: r.CantidadEventos,
				UnidadesNetas:   r.UnidadesNetas,
			},
		)
	}

	return respuesta, nil
}

// ==================== ALERTAS ====================

func (s *Service) ReporteAlertas(
	ctx context.Context,
	rol string,
	depositoUsuario string,
	filtro FiltroAlertasDTO,
) (ReporteAlertasDTO, error) {

	depositoID, err := resolverAlcanceDeposito(
		rol,
		depositoUsuario,
		filtro.DepositoID,
	)
	if err != nil {
		return ReporteAlertasDTO{}, err
	}

	oidDeposito, err := convertirObjectID(
		depositoID,
		ErrDepositoInvalido,
	)
	if err != nil {
		return ReporteAlertasDTO{}, err
	}

	dias := filtro.DiasVencimiento

	if dias <= 0 {
		dias = 30
	}

	filtroRepo := FiltroAlertas{
		DepositoID:      oidDeposito,
		DiasVencimiento: dias,
	}

	minimos, totalMinimos, err :=
		s.repositorio.AlertasBajoMinimo(
			ctx,
			filtroRepo,
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReporteAlertasDTO{}, err
	}

	vencimientos, totalVencimientos, err :=
		s.repositorio.AlertasPorVencer(
			ctx,
			filtroRepo,
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReporteAlertasDTO{}, err
	}

	respuesta := ReporteAlertasDTO{
		BajoMinimo: make(
			[]ProductoBajoMinimoDTO,
			0,
			len(minimos),
		),

		PaginacionMinimo: PaginacionAPaginaDTO(
			filtro.Pagina,
			totalMinimos,
		),

		PorVencer: make(
			[]ProductoPorVencerDTO,
			0,
			len(vencimientos),
		),

		PaginacionVencimiento: PaginacionAPaginaDTO(
			filtro.Pagina,
			totalVencimientos,
		),
	}

	for _, r := range minimos {
		respuesta.BajoMinimo = append(
			respuesta.BajoMinimo,
			ProductoBajoMinimoDTO{
				DepositoID:     r.DepositoID.Hex(),
				DepositoNombre: r.DepositoNombre,
				ProductoID:     r.ProductoID.Hex(),
				ProductoNombre: r.ProductoNombre,
				Stock:          r.Stock,
				StockMinimo:    r.StockMinimo,
			},
		)
	}

	for _, r := range vencimientos {
		respuesta.PorVencer = append(
			respuesta.PorVencer,
			ProductoPorVencerDTO{
				DepositoID:       r.DepositoID.Hex(),
				DepositoNombre:   r.DepositoNombre,
				ProductoID:       r.ProductoID.Hex(),
				ProductoNombre:   r.ProductoNombre,
				Stock:            r.Stock,
				FechaVencimiento: r.FechaVencimiento.Format("2006-01-02"),
				DiasParaVencer:   r.DiasParaVencer,
			},
		)
	}

	return respuesta, nil
}

// ==================== PENDIENTES ====================

func (s *Service) ReportePendientes(
	ctx context.Context,
	rol string,
	depositoUsuario string,
	filtro FiltroPendientesDTO,
) (ReportePendientesDTO, error) {

	depositoID, err := resolverAlcanceDeposito(
		rol,
		depositoUsuario,
		filtro.DepositoID,
	)
	if err != nil {
		return ReportePendientesDTO{}, err
	}

	oidDeposito, err := convertirObjectID(
		depositoID,
		ErrDepositoInvalido,
	)
	if err != nil {
		return ReportePendientesDTO{}, err
	}

	filtroRepo := FiltroPendientes{
		DepositoID: oidDeposito,
	}

	ordenes, totalOrdenes, err :=
		s.repositorio.OrdenesPendientes(
			ctx,
			filtroRepo,
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReportePendientesDTO{}, err
	}

	transferencias, totalTransferencias, err :=
		s.repositorio.TransferenciasPendientes(
			ctx,
			filtroRepo,
			filtro.Pagina,
			TamanioPagina,
		)

	if err != nil {
		return ReportePendientesDTO{}, err
	}

	respuesta := ReportePendientesDTO{
		Ordenes: make(
			[]EstadoOrdenesDepositoDTO,
			0,
			len(ordenes),
		),

		PaginacionOrdenes: PaginacionAPaginaDTO(
			filtro.Pagina,
			totalOrdenes,
		),

		Transferencias: make(
			[]EstadoTransferenciasDepositoDTO,
			0,
			len(transferencias),
		),

		PaginacionTransferencias: PaginacionAPaginaDTO(
			filtro.Pagina,
			totalTransferencias,
		),
	}

	for _, r := range ordenes {
		respuesta.Ordenes = append(
			respuesta.Ordenes,
			EstadoOrdenesDepositoDTO{
				DepositoID:     r.DepositoID.Hex(),
				DepositoNombre: r.DepositoNombre,
				Borrador:       r.Borrador,
				Confirmada:     r.Confirmada,
				Parcial:        r.Parcial,
				Completada:     r.Completada,
			},
		)
	}

	for _, r := range transferencias {
		respuesta.Transferencias = append(
			respuesta.Transferencias,
			EstadoTransferenciasDepositoDTO{
				DepositoID:            r.DepositoID.Hex(),
				DepositoNombre:        r.DepositoNombre,
				SolicitadasComoOrigen: r.SolicitadasComoOrigen,
				PendientesDeAprobar:   r.PendientesDeAprobar,
				AprobadasPorRecibir:   r.AprobadasPorRecibir,
			},
		)
	}

	return respuesta, nil
}
