package categoria

import "go.mongodb.org/mongo-driver/v2/bson"

type CategoriaDTO struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

type CrearCategoriaDTO struct {
	Nombre string `json:"nombre" binding:"required"`
}

type ActualizarCategoriaDTO struct {
	Nombre string `json:"nombre" binding:"required"`
}

func MapeoACategoria(c CategoriaDTO) (Categoria, error) {
	oid, err := bson.ObjectIDFromHex(c.ID)
	if err != nil {
		return Categoria{}, err
	}

	return Categoria{
		ID:     oid,
		Nombre: c.Nombre,
	}, nil
}

func MapeoCrearDTOACategoria(c CrearCategoriaDTO) (Categoria, error) {
	return Categoria{
		Nombre: c.Nombre,
	}, nil

}

func MapeoActualizarDTOACategoria(
	c ActualizarCategoriaDTO,
) Categoria {
	return Categoria{
		Nombre: c.Nombre,
	}
}
