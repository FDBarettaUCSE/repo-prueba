package ordendecompra

import (
	ordendecompraitem "api/internal/ordenDeCompraItem"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CrearOrdenCompraDTO struct {
	ProveedorID string `json:"proveedorId" binding:"required"`
	DepositoID  string `json:"depositoId" binding:"required"`
}

type ActualizarOrdenCompraDTO struct {
	ID          string `json:"id"`
	ProveedorID string `json:"proveedorId" binding:"required"`
	DepositoID  string `json:"depositoId" binding:"required"`
}

type OrdenCompraDTO struct {
	ID          string                                 `json:"id"`
	ProveedorID string                                 `json:"proveedorId"`
	DepositoID  string                                 `json:"depositoId"`
	Estado      EstadoOrdenCompra                      `json:"estado"`
	Items       []ordendecompraitem.OrdenCompraItemDTO `json:"items"`
}

// FiltroOrdenesCompraDTO agrupa los filtros y paginación para órdenes de compra
type FiltroOrdenesCompraDTO struct {
	DepositoID    string `form:"deposito_id"`
	Estado        string `form:"estado"`
	ProveedorID   string `form:"proveedor_id"`
	Pagina        int    `form:"pagina"`
	TamanioPagina int    `form:"tamanioPagina"`
}

// OrdenesCompraPaginadasDTO envuelve el listado con metadata de paginación
type OrdenesCompraPaginadasDTO struct {
	Items         []OrdenCompraDTO `json:"items"`
	Total         int64            `json:"total"`
	Pagina        int              `json:"pagina"`
	TamanioPagina int              `json:"tamanioPagina"`
}

func MapeoCrearDTOAOrdenCompra(dto CrearOrdenCompraDTO) (OrdenCompra, error) {

	proveedorID, err := bson.ObjectIDFromHex(dto.ProveedorID)
	if err != nil {
		return OrdenCompra{}, fmt.Errorf("%w: id de proveedor inválido", ErrDatosOrdenInvalidos)
	}

	depositoID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return OrdenCompra{}, fmt.Errorf("%w: id de depósito inválido", ErrDatosOrdenInvalidos)
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
		return OrdenCompra{}, fmt.Errorf("%w: id de proveedor inválido", ErrDatosOrdenInvalidos)
	}

	depositoID, err := bson.ObjectIDFromHex(dto.DepositoID)
	if err != nil {
		return OrdenCompra{}, fmt.Errorf("%w: id de depósito inválido", ErrDatosOrdenInvalidos)
	}

	id, err := bson.ObjectIDFromHex(dto.ID)
	if err != nil {
		return OrdenCompra{}, ErrIDOrdenCompraInvalido
	}

	return OrdenCompra{
		ID:          id,
		ProveedorID: proveedorID,
		DepositoID:  depositoID,
	}, nil
}

// Para actualizar los productos dentro de una orden de compra

type RegistrarRecepcionDTO struct {
	Items []RecepcionItemDTO `json:"items" binding:"required,min=1,dive"`
}

type RecepcionItemDTO struct {
	ItemID   string `json:"itemId" binding:"required"`
	Cantidad int32  `json:"cantidad" binding:"required,gt=0"`
}

// Ver si hay alguna manera mas facil de hacerlo.

func MapeoRegistrarRecepcionDTO(
	dto RegistrarRecepcionDTO,
) (Recepcion, error) {

	if len(dto.Items) == 0 {
		return Recepcion{}, fmt.Errorf(
			"%w: la recepción debe contener al menos un item",
			ErrRecepcionInvalida,
		)
	}

	recepcion := Recepcion{
		Items: make([]RecepcionItem, 0, len(dto.Items)),
	}

	itemsVistos := make(map[bson.ObjectID]bool)

	for _, dtoItem := range dto.Items {

		if dtoItem.Cantidad <= 0 {
			return Recepcion{}, fmt.Errorf(
				"%w: la cantidad recibida debe ser mayor a cero",
				ErrRecepcionInvalida,
			)
		}

		itemID, err := bson.ObjectIDFromHex(dtoItem.ItemID)
		if err != nil {
			return Recepcion{}, fmt.Errorf(
				"%w: id de item inválido",
				ErrRecepcionInvalida,
			)
		}

		if itemsVistos[itemID] {
			return Recepcion{}, fmt.Errorf(
				"%w: no se puede repetir un item dentro de una recepción",
				ErrRecepcionInvalida,
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
