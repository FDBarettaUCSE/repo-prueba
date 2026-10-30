package depositoproducto

// MaxTamanioPagina acota la cantidad máxima de filas por página para evitar
// que un cliente pida toda la colección en una sola consulta.
const MaxTamanioPagina = 100

// StockDepositoDTO es la vista enriquecida de una fila de stock: además de
// depósito/producto/cantidad, trae todo lo que el frontend necesita mostrar
// sin tener que salir a pedirlo aparte a /productos o /depositos (nombre,
// categoría, unidad de medida, si está bajo mínimo). CostoUnitario y
// Valorizado se omiten cuando consulta un Operario, que no tiene permiso de
// acceso a la valorización del inventario.
type StockDepositoDTO struct {
	DepositoID          string   `json:"depositoId"`
	DepositoNombre      string   `json:"depositoNombre"`
	ProductoID          string   `json:"productoId"`
	ProductoNombre      string   `json:"productoNombre"`
	CategoriaID         string   `json:"categoriaId"`
	CategoriaNombre     string   `json:"categoriaNombre"`
	UnidadMedida        string   `json:"unidadMedida"`
	Stock               int32    `json:"stock"`
	StockMinimo         int32    `json:"stockMinimo"`                   // efectivo: override si existe, sino el general del producto
	StockMinimoOverride *int32   `json:"stockMinimoOverride,omitempty"` // no-nil solo si HAY una configuración particular para este depósito
	CostoUnitario       *float64 `json:"costoUnitario,omitempty"`
	Valorizado          *float64 `json:"valorizado,omitempty"` // stock * costoUnitario
	BajoMinimo          bool     `json:"bajoMinimo"`
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
	TamanioPagina  int
}

// StockDepositoPaginadoDTO envuelve el listado con metadata de paginación,
// mismo criterio que va a repetirse en el resto de los listados del sistema.
type StockDepositoPaginadoDTO struct {
	Items         []StockDepositoDTO `json:"items"`
	Total         int64              `json:"total"`
	Pagina        int                `json:"pagina"`
	TamanioPagina int                `json:"tamanioPagina"`
}

func MapeoAFiltroRepositorio(f ListarStockFiltro) ListarStockFiltroRepositorio {
	return ListarStockFiltroRepositorio{
		CategoriaID:    f.CategoriaID,
		SoloBajoMinimo: f.SoloBajoMinimo,
		Pagina:         f.Pagina,
		TamanioPagina:  f.TamanioPagina,
	}
}
