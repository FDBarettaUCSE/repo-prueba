package producto

import (
	"api/internal/auditoria"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UnidadMedida string

const (
	Unidad    UnidadMedida = "unidad"
	Kilogramo UnidadMedida = "kilogramo"
	Litro     UnidadMedida = "litro"
	Metro     UnidadMedida = "metro"
)

type Producto struct {
	ID               bson.ObjectID       `bson:"_id,omitempty"`
	Nombre           string              `bson:"nombre"`
	Descripcion      string              `bson:"descripcion"`
	Categoria        bson.ObjectID       `bson:"categoria"`
	UnidadMedida     UnidadMedida        `bson:"unidad_medida"`
	CostoUnitario    float64             `bson:"costo_unitario"`
	StockMinimo      int32               `bson:"stock_minimo"`
	FechaVencimiento *time.Time          `bson:"fecha_vencimiento, omitempty"`
	FechaEliminacion *time.Time          `bson:"fecha_eliminacion, omitempty"`
	Auditoria        auditoria.Auditoria `bson:"auditoria"`
}

func MapeoAProductoDTO(p Producto, categoriaNombre string) ProductoDTO {
	return ProductoDTO{
		ID:               p.ID.Hex(),
		Nombre:           p.Nombre,
		Descripcion:      p.Descripcion,
		CategoriaID:      p.Categoria.Hex(),
		CategoriaNombre:  categoriaNombre,
		UnidadMedida:     p.UnidadMedida,
		CostoUnitario:    p.CostoUnitario,
		StockMinimo:      p.StockMinimo,
		FechaVencimiento: p.FechaVencimiento,
	}
}
