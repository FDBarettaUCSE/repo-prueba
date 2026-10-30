package reporte

// ===== Paginación =====

type PaginacionDTO struct {
	Pagina        int   `json:"pagina"`
	TamanioPagina int   `json:"tamanioPagina"`
	Total         int64 `json:"total"`
}

// ===== Reporte de inventario =====

type FiltroInventarioDTO struct {
	DepositoID  string `form:"deposito_id"`
	CategoriaID string `form:"categoria_id"`
	Pagina      int    `form:"pagina"`
}

type ValorizacionDepositoDTO struct {
	DepositoID     string  `json:"deposito_id"`
	DepositoNombre string  `json:"deposito_nombre"`
	Stock          int32   `json:"stock"`
	Valorizado     float64 `json:"valorizado"`
}

type ValorizacionCategoriaDTO struct {
	CategoriaID     string  `json:"categoria_id"`
	CategoriaNombre string  `json:"categoria_nombre"`
	Stock           int32   `json:"stock"`
	Valorizado      float64 `json:"valorizado"`
}

type ReporteInventarioDTO struct {
	PorDeposito        []ValorizacionDepositoDTO `json:"por_deposito"`
	PaginacionDeposito PaginacionDTO             `json:"paginacion_deposito"`

	PorCategoria        []ValorizacionCategoriaDTO `json:"por_categoria"`
	PaginacionCategoria PaginacionDTO              `json:"paginacion_categoria"`

	StockTotal      int32   `json:"stock_total"`
	ValorizadoTotal float64 `json:"valorizado_total"`
}

// ===== Reporte de movimientos =====

type FiltroReporteMovimientosDTO struct {
	DepositoID string `form:"deposito_id"`
	Tipo       string `form:"tipo"`
	FechaDesde string `form:"fecha_desde"`
	FechaHasta string `form:"fecha_hasta"`
	Pagina     int    `form:"pagina"`
}

type ResumenMovimientoDTO struct {
	DepositoID      string `json:"deposito_id"`
	DepositoNombre  string `json:"deposito_nombre"`
	Fecha           string `json:"fecha"`
	Tipo            string `json:"tipo"`
	CantidadEventos int64  `json:"cantidad_eventos"`
	UnidadesNetas   int32  `json:"unidades_netas"`
}

type ReporteMovimientosDTO struct {
	Resumen          []ResumenMovimientoDTO `json:"resumen"`
	TotalMovimientos int64                  `json:"total_movimientos"`
	Paginacion       PaginacionDTO          `json:"paginacion"`
}

// ===== Reporte de alertas =====

type FiltroAlertasDTO struct {
	DepositoID      string `form:"deposito_id"`
	DiasVencimiento int    `form:"dias_vencimiento"`
	Pagina          int    `form:"pagina"`
}

type ProductoBajoMinimoDTO struct {
	DepositoID     string `json:"deposito_id"`
	DepositoNombre string `json:"deposito_nombre"`
	ProductoID     string `json:"producto_id"`
	ProductoNombre string `json:"producto_nombre"`
	Stock          int32  `json:"stock"`
	StockMinimo    int32  `json:"stock_minimo"`
}

type ProductoPorVencerDTO struct {
	DepositoID       string `json:"deposito_id"`
	DepositoNombre   string `json:"deposito_nombre"`
	ProductoID       string `json:"producto_id"`
	ProductoNombre   string `json:"producto_nombre"`
	Stock            int32  `json:"stock"`
	FechaVencimiento string `json:"fecha_vencimiento"`
	DiasParaVencer   int    `json:"dias_para_vencer"`
}

type ReporteAlertasDTO struct {
	BajoMinimo       []ProductoBajoMinimoDTO `json:"bajo_minimo"`
	PaginacionMinimo PaginacionDTO           `json:"paginacion_minimo"`

	PorVencer             []ProductoPorVencerDTO `json:"por_vencer"`
	PaginacionVencimiento PaginacionDTO          `json:"paginacion_vencimiento"`
}

// ===== Reporte de pendientes =====

type FiltroPendientesDTO struct {
	DepositoID string `form:"deposito_id"`
	Pagina     int    `form:"pagina"`
}

type EstadoOrdenesDepositoDTO struct {
	DepositoID     string `json:"deposito_id"`
	DepositoNombre string `json:"deposito_nombre"`
	Borrador       int64  `json:"borrador"`
	Confirmada     int64  `json:"confirmada"`
	Parcial        int64  `json:"parcial"`
	Completada     int64  `json:"completada"`
}

type EstadoTransferenciasDepositoDTO struct {
	DepositoID            string `json:"deposito_id"`
	DepositoNombre        string `json:"deposito_nombre"`
	SolicitadasComoOrigen int64  `json:"solicitadas_como_origen"`
	PendientesDeAprobar   int64  `json:"pendientes_de_aprobar"`
	AprobadasPorRecibir   int64  `json:"aprobadas_por_recibir"`
}

type ReportePendientesDTO struct {
	Ordenes           []EstadoOrdenesDepositoDTO `json:"ordenes"`
	PaginacionOrdenes PaginacionDTO              `json:"paginacion_ordenes"`

	Transferencias           []EstadoTransferenciasDepositoDTO `json:"transferencias"`
	PaginacionTransferencias PaginacionDTO                     `json:"paginacion_transferencias"`
}
