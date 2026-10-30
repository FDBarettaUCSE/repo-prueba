package producto

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductoDTO struct {
	ID               string       `json:"id"`
	Nombre           string       `json:"nombre"`
	Descripcion      string       `json:"descripcion"`
	CategoriaID      string       `json:"categoriaId"`
	CategoriaNombre  string       `json:"categoriaNombre"`
	UnidadMedida     UnidadMedida `json:"unidadMedida"`
	CostoUnitario    float64      `json:"costoUnitario"`
	StockMinimo      int32        `json:"stockMinimo"`
	FechaVencimiento *time.Time   `json:"fechaVencimiento,omitempty"`
}

// Costo y stock mínimo usan gte=0 (y no "required") a propósito: 0 es un
// valor válido, así que no puede tratarse como "campo vacío". El service
// vuelve a validar los rangos (ver validarDatosProducto) para no depender
// solo del binding.
type ActualizarProductoDTO struct {
	Nombre           string       `json:"nombre" binding:"required"`
	Descripcion      string       `json:"descripcion"`
	CategoriaID      string       `json:"categoriaId" binding:"required"`
	UnidadMedida     UnidadMedida `json:"unidadMedida" binding:"required"`
	CostoUnitario    float64      `json:"costoUnitario" binding:"gte=0"`
	FechaVencimiento *time.Time   `json:"fechaVencimiento,omitempty"`
}

type StockMinimoDTO struct {
	StockMinimo int32 `json:"stockMinimo" binding:"gte=0"`
}

type CrearProductoDTO struct {
	Nombre           string       `json:"nombre" binding:"required"`
	Descripcion      string       `json:"descripcion"`
	CategoriaID      string       `json:"categoriaId" binding:"required"`
	UnidadMedida     UnidadMedida `json:"unidadMedida" binding:"required"`
	CostoUnitario    float64      `json:"costoUnitario" binding:"gte=0"`
	StockMinimo      int32        `json:"stockMinimo" binding:"gte=0"`
	FechaVencimiento *time.Time   `json:"fechaVencimiento,omitempty"`
}

func MapeoCrearDTOAProducto(p CrearProductoDTO) (Producto, error) {
	categoriaID, err := bson.ObjectIDFromHex(p.CategoriaID)
	if err != nil {
		return Producto{}, err
	}

	return Producto{
		Nombre:           p.Nombre,
		Descripcion:      p.Descripcion,
		Categoria:        categoriaID,
		UnidadMedida:     p.UnidadMedida,
		CostoUnitario:    p.CostoUnitario,
		StockMinimo:      p.StockMinimo,
		FechaVencimiento: p.FechaVencimiento,
	}, nil
}

func MapeoActualizarDTOAProducto(
	d ActualizarProductoDTO,
) (Producto, error) {

	categoriaID, err := bson.ObjectIDFromHex(d.CategoriaID)
	if err != nil {
		return Producto{}, err
	}

	return Producto{
		Nombre:           d.Nombre,
		Descripcion:      d.Descripcion,
		Categoria:        categoriaID,
		UnidadMedida:     d.UnidadMedida,
		CostoUnitario:    d.CostoUnitario,
		FechaVencimiento: d.FechaVencimiento,
	}, nil
}

// FiltroProductosDTO agrupa los filtros y paginación para el catálogo de productos
type FiltroProductosDTO struct {
	CategoriaID   string `form:"categoria_id"`
	Pagina        int    `form:"pagina"`
	TamanioPagina int    `form:"tamanioPagina"`
	Todos         string `form:"todos"`
}

// ProductosPaginadosDTO envuelve el listado de productos con metadata de paginación
type ProductosPaginadosDTO struct {
	Items         []ProductoDTO `json:"items"`
	Total         int64         `json:"total"`
	Pagina        int           `json:"pagina"`
	TamanioPagina int           `json:"tamanioPagina"`
}

