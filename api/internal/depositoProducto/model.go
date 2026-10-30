package depositoproducto

import (
	"api/internal/auditoria"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// DepositoProducto es la fila de stock de UN producto en UN depósito
// puntual (ver Deposito.RegistrarRutas y Producto: cada depósito mantiene su
// propio stock, independiente del de los demás). No hay una fila para cada
// combinación posible desde el principio: se crea recién con el primer
// movimiento de stock que la necesite (ver AsegurarExiste en repository.go).
//
// StockMinimo es un *int32 (puntero) a propósito: nil significa "este
// depósito NO tiene una configuración particular, usa el stock mínimo
// general del producto" (ver Producto.StockMinimo). Un valor no-nil,
// incluido *0, es un override explícito que reemplaza al general para este
// depósito puntual — esa es la única forma de distinguir "no configurado"
// de "configurado en cero".
type DepositoProducto struct {
	ID          bson.ObjectID       `bson:"_id,omitempty"`
	DepositoID  bson.ObjectID       `bson:"deposito_id"`
	ProductoID  bson.ObjectID       `bson:"producto_id"`
	Stock       int32               `bson:"stock"`
	StockMinimo *int32              `bson:"stock_minimo,omitempty"`
	Auditoria   auditoria.Auditoria `bson:"auditoria"`
}

// Se duplica el filtro para mantener desacopladas las capas.
type ListarStockFiltroRepositorio struct {
	CategoriaID    string
	SoloBajoMinimo bool
	Pagina         int
	TamanioPagina  int
}

// Type que devuelve la consulta por stock
type StockDepositoFila struct {
	DepositoID          bson.ObjectID `bson:"deposito_id"`
	DepositoNombre      string        `bson:"deposito_nombre"`
	ProductoID          bson.ObjectID `bson:"producto_id"`
	ProductoNombre      string        `bson:"producto_nombre"`
	CategoriaID         bson.ObjectID `bson:"categoria_id"`
	CategoriaNombre     string        `bson:"categoria_nombre"`
	UnidadMedida        string        `bson:"unidad_medida"`
	Stock               int32         `bson:"stock"`
	StockMinimo         int32         `bson:"stock_minimo_efectivo"`
	StockMinimoOverride *int32        `bson:"stock_minimo_override"`
	CostoUnitario       float64       `bson:"costo_unitario"`
	Valorizado          float64       `bson:"valorizado"`
	BajoMinimo          bool          `bson:"bajo_minimo"`
}

// MapearAStockDepositoDTO convierte una fila de stock del repositorio
// al DTO que devuelve la API, ocultando la valorización económica para operarios.
func MapearAStockDepositoDTO(fila StockDepositoFila, rol string) StockDepositoDTO {
	dto := StockDepositoDTO{
		DepositoID:          fila.DepositoID.Hex(),
		DepositoNombre:      fila.DepositoNombre,
		ProductoID:          fila.ProductoID.Hex(),
		ProductoNombre:      fila.ProductoNombre,
		CategoriaID:         fila.CategoriaID.Hex(),
		CategoriaNombre:     fila.CategoriaNombre,
		UnidadMedida:        fila.UnidadMedida,
		Stock:               fila.Stock,
		StockMinimo:         fila.StockMinimo,
		StockMinimoOverride: fila.StockMinimoOverride,
		BajoMinimo:          fila.BajoMinimo,
	}

	if rol != "operario" {
		costoUnitario := fila.CostoUnitario
		valorizado := fila.Valorizado
		dto.CostoUnitario = &costoUnitario
		dto.Valorizado = &valorizado
	}

	return dto
}
