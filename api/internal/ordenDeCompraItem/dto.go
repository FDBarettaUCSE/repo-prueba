package ordendecompraitem

import "go.mongodb.org/mongo-driver/v2/bson"

type OrdenCompraItemDTO struct {
	ID                 string  `json:"id"`
	OrdenCompraID      string  `json:"ordenCompraId"`
	ProductoID         string  `json:"productoId"`
	CantidadSolicitada int32   `json:"cantidadSolicitada"`
	CantidadRecibida   int32   `json:"cantidadRecibida"`
	CostoUnitario      float64 `json:"costoUnitario"`
}

type CrearOrdenCompraItemDTO struct {
	ProductoID         string `json:"productoId"`
	CantidadSolicitada int32  `json:"cantidadSolicitada"`
}

type ActualizarOrdenCompraItemDTO struct {
	ID                 string `json:"id"`
	ProductoID         string `json:"productoId"`
	CantidadSolicitada int32  `json:"cantidadSolicitada"`
}

func MapeoCrearDTOAOrdenCompraItem(dto CrearOrdenCompraItemDTO) (OrdenCompraItem, error) {

	productoID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	return OrdenCompraItem{
		ProductoID:         productoID,
		CantidadSolicitada: dto.CantidadSolicitada,
		CantidadRecibida:   0,
	}, nil
}

func MapeoActualizarDTOAOrdenCompraItem(
	dto ActualizarOrdenCompraItemDTO,
) (OrdenCompraItem, error) {

	productoID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	id, err := bson.ObjectIDFromHex(dto.ID)
	if err != nil {
		return OrdenCompraItem{}, err
	}

	return OrdenCompraItem{
		ID:                 id,
		ProductoID:         productoID,
		CantidadSolicitada: dto.CantidadSolicitada,
	}, nil
}
