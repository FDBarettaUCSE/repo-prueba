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

type ActualizarProductoDTO struct {
	ID               string       `json:"id"`
	Nombre           string       `json:"nombre"`
	Descripcion      string       `json:"descripcion"`
	CategoriaID      string       `json:"categoriaId"`
	UnidadMedida     UnidadMedida `json:"unidadMedida"`
	CostoUnitario    float64      `json:"costoUnitario"`
	FechaVencimiento *time.Time   `json:"fechaVencimiento,omitempty"`
}

type StockMinimoDTO struct {
	StockMinimo int32 `json:"stockMinimo"`
}

type CrearProductoDTO struct {
	Nombre           string       `json:"nombre"`
	Descripcion      string       `json:"descripcion"`
	CategoriaID      string       `json:"categoriaId"`
	UnidadMedida     UnidadMedida `json:"unidadMedida"`
	CostoUnitario    float64      `json:"costoUnitario"`
	StockMinimo      int32        `json:"stockMinimo"`
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

func MapeoActualizarDTOAProducto(d ActualizarProductoDTO) (Producto, error) {
	id, err := bson.ObjectIDFromHex(d.ID)
	if err != nil {
		return Producto{}, err
	}

	categoriaID, err := bson.ObjectIDFromHex(d.CategoriaID)
	if err != nil {
		return Producto{}, err
	}

	return Producto{
		ID:               id,
		Nombre:           d.Nombre,
		Descripcion:      d.Descripcion,
		Categoria:        categoriaID,
		UnidadMedida:     d.UnidadMedida,
		CostoUnitario:    d.CostoUnitario,
		FechaVencimiento: d.FechaVencimiento,
	}, nil
}
