package alertastock

// AlertaStockDTO es la vista de salida de una alerta, enriquecida con los
// nombres del producto y del depósito para que el frontend no tenga que
// resolver IDs a nombres por su cuenta — mismo criterio que usa
// depositoProducto.StockDepositoDTO y reporte.ProductoBajoMinimoDTO.
type AlertaStockDTO struct {
	ID               string     `json:"id"`
	Tipo             TipoAlerta `json:"tipo"`
	ProductoID       string     `json:"producto_id"`
	ProductoNombre   string     `json:"producto_nombre"`
	DepositoID       string     `json:"deposito_id"`
	DepositoNombre   string     `json:"deposito_nombre"`
	StockMinimo      int32      `json:"stock_minimo"`
	StockActual      int32      `json:"stock_actual"`
	FechaVencimiento *string    `json:"fecha_vencimiento,omitempty"`
	Resuelta         bool       `json:"resuelta"`
	FechaCreacion    string     `json:"fecha_creacion"`
	FechaResolucion  *string    `json:"fecha_resolucion,omitempty"`
	ModificadoPor    *string    `json:"modificado_por,omitempty"`
}

// AlertasPaginadasDTO envuelve el listado con metadata de paginación y el
// contador de alertas pendientes (no resueltas). Pendientes es independiente
// del filtro y de la página pedidos — mismo criterio que
// novedadDeposito.NovedadesPaginadoDTO con NoLeidas.
type AlertasPaginadasDTO struct {
	Items         []AlertaStockDTO `json:"items"`
	Total         int64            `json:"total"`
	Pendientes    int64            `json:"pendientes"`
	Pagina        int              `json:"pagina"`
	TamanioPagina int              `json:"tamanioPagina"`
}

// FiltroAlertasDTO agrupa los filtros y la paginación del listado de alertas.
type FiltroAlertasDTO struct {
	DepositoID     string `form:"deposito_id"`
	SoloPendientes string `form:"solo_pendientes"` // "true" para excluir las resueltas
	Pagina         int    `form:"pagina"`
	TamanioPagina  int    `form:"tamanioPagina"`
}
