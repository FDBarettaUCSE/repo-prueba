package reporte

// ===== Reporte de inventario (stock y valorización) =====

// FiltroInventarioDTO: depósito y categoría opcionales — si no se manda
// depósito (o el rol no puede elegir), se agrega TODA la empresa.
type FiltroInventarioDTO struct {
	DepositoID  string `form:"deposito_id"`
	CategoriaID string `form:"categoria_id"`
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
	PorDeposito     []ValorizacionDepositoDTO  `json:"por_deposito"`
	PorCategoria    []ValorizacionCategoriaDTO `json:"por_categoria"`
	StockTotal      int32                      `json:"stock_total"`
	ValorizadoTotal float64                    `json:"valorizado_total"`
}

// ===== Reporte de movimientos de stock =====

// FiltroReporteMovimientosDTO: mismos filtros que el listado crudo de
// /movimientos (ver movimientoStock), reusados acá para agregar en vez de
// paginar.
type FiltroReporteMovimientosDTO struct {
	DepositoID string `form:"deposito_id"`
	Tipo       string `form:"tipo"`
	FechaDesde string `form:"fecha_desde"`
	FechaHasta string `form:"fecha_hasta"`
}

type ResumenMovimientoDTO struct {
	DepositoID      string `json:"deposito_id"`
	DepositoNombre  string `json:"deposito_nombre"`
	Fecha           string `json:"fecha"` // YYYY-MM-DD
	Tipo            string `json:"tipo"`
	CantidadEventos int64  `json:"cantidad_eventos"` // cuántos movimientos
	UnidadesNetas   int32  `json:"unidades_netas"`   // suma con signo de "cantidad" (ver movimientoStock.MovimientoStock.Cantidad)
}

type ReporteMovimientosDTO struct {
	Resumen          []ResumenMovimientoDTO `json:"resumen"`
	TotalMovimientos int64                  `json:"total_movimientos"`
}

// ===== Reporte de alertas (stock mínimo y próximo vencimiento) =====

// FiltroAlertasDTO: DiasVencimiento es la ventana configurable del
// enunciado ("productos próximos a vencer, dentro de una ventana de días
// configurable"); si no se manda, el service usa un default (30).
type FiltroAlertasDTO struct {
	DepositoID      string `form:"deposito_id"`
	DiasVencimiento int    `form:"dias_vencimiento"`
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
	BajoMinimo []ProductoBajoMinimoDTO `json:"bajo_minimo"`
	PorVencer  []ProductoPorVencerDTO  `json:"por_vencer"`
}

// ===== Reporte de órdenes de compra y transferencias pendientes =====

type FiltroPendientesDTO struct {
	DepositoID string `form:"deposito_id"`
}

type EstadoOrdenesDepositoDTO struct {
	DepositoID     string `json:"deposito_id"`
	DepositoNombre string `json:"deposito_nombre"`
	Borrador       int64  `json:"borrador"`
	Confirmada     int64  `json:"confirmada"`
	Parcial        int64  `json:"parcial"`
	Completada     int64  `json:"completada"`
}

// EstadoTransferenciasDepositoDTO mira las transferencias desde el punto de
// vista de "qué le falta hacer a este depósito", no un conteo crudo por
// estado: SolicitadasComoOrigen son las que este depósito pidió y todavía
// esperan que el destino resuelva; PendientesDeAprobar son las que ESTE
// depósito (como destino) tiene que aprobar o rechazar; AprobadasPorRecibir
// son las que ya aprobó como destino y todavía no se completaron (falta que
// llegue la mercadería). Ver enunciado: el dashboard del Gerente necesita
// justamente "las transferencias que requieren su aprobación".
type EstadoTransferenciasDepositoDTO struct {
	DepositoID            string `json:"deposito_id"`
	DepositoNombre        string `json:"deposito_nombre"`
	SolicitadasComoOrigen int64  `json:"solicitadas_como_origen"`
	PendientesDeAprobar   int64  `json:"pendientes_de_aprobar"`
	AprobadasPorRecibir   int64  `json:"aprobadas_por_recibir"`
}

type ReportePendientesDTO struct {
	Ordenes        []EstadoOrdenesDepositoDTO        `json:"ordenes"`
	Transferencias []EstadoTransferenciasDepositoDTO `json:"transferencias"`
}
