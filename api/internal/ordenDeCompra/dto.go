package ordendecompra

import (
	ordendecompraitem "api/internal/ordenDeCompraItem"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CrearOrdenCompraDTO struct {
	ProveedorID string `json:"proveedorId"`
	DepositoID  string `json:"depositoId"`
}

type ActualizarOrdenCompraDTO struct {
	ID          string `json:"id"`
	ProveedorID string `json:"proveedorId"`
	DepositoID  string `json:"depositoId"`
}

type OrdenCompraDTO struct {
	ID          string                                 `json:"id"`
	ProveedorID string                                 `json:"proveedorId"`
	DepositoID  string                                 `json:"depositoId"`
	Estado      EstadoOrdenCompra                      `json:"estado"`
	Items       []ordendecompraitem.OrdenCompraItemDTO `json:"items"`
}

func MapeoCrearDTOAOrdenCompra(dto CrearOrdenCompraDTO) (OrdenCompra, error) {

	proveedorID, err := bson.ObjectIDFromHex(dto.ProveedorID)
	if err != nil {
		return OrdenCompra{}, err
	}

	depositoID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return OrdenCompra{}, err
	}

	return OrdenCompra{
		ProveedorID: proveedorID,
		DepositoID:  depositoID,
		Estado:      OrdenBorrador,
	}, nil
}

func MapeoActualizarDTOAOrdenCompra(dto ActualizarOrdenCompraDTO) (OrdenCompra, error) {

	proveedorID, err := bson.ObjectIDFromHex(dto.ProveedorID)
	if err != nil {
		return OrdenCompra{}, err
	}

	depositoID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return OrdenCompra{}, err
	}

	id, err := bson.ObjectIDFromHex(dto.ID)
	if err != nil {
		return OrdenCompra{}, err
	}

	return OrdenCompra{
		ID:          id,
		ProveedorID: proveedorID,
		DepositoID:  depositoID,
	}, nil
}

// Para actualizar los productos dentro de una orden de compra

type RegistrarRecepcionDTO struct {
	Items []RecepcionItemDTO `json:"items"`
}

type RecepcionItemDTO struct {
	ItemID   string `json:"itemId"`
	Cantidad int32  `json:"cantidad"`
}

// Ver si hay alguna manera mas facil de hacerlo.

func MapeoRegistrarRecepcionDTO(
	dto RegistrarRecepcionDTO,
) (Recepcion, error) {

	if len(dto.Items) == 0 {
		return Recepcion{}, errors.New(
			"la recepción debe contener al menos un item",
		)
	}

	recepcion := Recepcion{
		Items: make([]RecepcionItem, 0, len(dto.Items)),
	}

	itemsVistos := make(map[bson.ObjectID]bool)

	for _, dtoItem := range dto.Items {

		if dtoItem.Cantidad <= 0 {
			return Recepcion{}, errors.New(
				"la cantidad recibida debe ser mayor a cero",
			)
		}

		itemID, err := bson.ObjectIDFromHex(dtoItem.ItemID)
		if err != nil {
			return Recepcion{}, err
		}

		if itemsVistos[itemID] {
			return Recepcion{}, errors.New(
				"no se puede repetir un item dentro de una recepción",
			)
		}

		itemsVistos[itemID] = true

		recepcion.Items = append(
			recepcion.Items,
			RecepcionItem{
				ItemID:   itemID,
				Cantidad: dtoItem.Cantidad,
			},
		)
	}

	return recepcion, nil
}
