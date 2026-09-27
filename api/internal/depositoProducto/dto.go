package depositoproducto

// StockDepositoDTO es la vista enriquecida de una fila de stock: además de
// depósito/producto/cantidad, trae todo lo que el frontend necesita mostrar
// sin tener que salir a pedirlo aparte a /productos o /depositos (nombre,
// categoría, unidad de medida, costo, valorización, si está bajo mínimo).
type StockDepositoDTO struct {
	DepositoID          string  `json:"depositoId"`
	DepositoNombre      string  `json:"depositoNombre"`
	ProductoID          string  `json:"productoId"`
	ProductoNombre      string  `json:"productoNombre"`
	CategoriaID         string  `json:"categoriaId"`
	CategoriaNombre     string  `json:"categoriaNombre"`
	UnidadMedida        string  `json:"unidadMedida"`
	Stock               int32   `json:"stock"`
	StockMinimo         int32   `json:"stockMinimo"`                   // efectivo: override si existe, sino el general del producto
	StockMinimoOverride *int32  `json:"stockMinimoOverride,omitempty"` // no-nil solo si HAY una configuración particular para este depósito
	CostoUnitario       float64 `json:"costoUnitario"`
	Valorizado          float64 `json:"valorizado"` // stock * costoUnitario
	BajoMinimo          bool    `json:"bajoMinimo"`
}

// ConfigurarStockMinimoDTO es el body de PUT .../stock-minimo. gte=0 (y no
// "required") a propósito: 0 es un valor válido de stock mínimo, así que no
// puede tratarse como "campo vacío".
type ConfigurarStockMinimoDTO struct {
	StockMinimo int32 `json:"stockMinimo" binding:"gte=0"`
}

// ListarStockFiltro agrupa los filtros y la paginación soportados por
// ListarStockDeposito (ver Clase 8 — "paginación y filtros en listados").
type ListarStockFiltro struct {
	CategoriaID    string
	SoloBajoMinimo bool
	Pagina         int
	TamañoPagina   int
}

// StockDepositoPaginadoDTO envuelve el listado con metadata de paginación,
// mismo criterio que va a repetirse en el resto de los listados del sistema.
type StockDepositoPaginadoDTO struct {
	Items        []StockDepositoDTO `json:"items"`
	Total        int64              `json:"total"`
	Pagina       int                `json:"pagina"`
	TamañoPagina int                `json:"tamañoPagina"`
}
