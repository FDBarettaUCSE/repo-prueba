package ordendecompra

import (
	"api/internal/deposito"
	ordendecompraitem "api/internal/ordenDeCompraItem"
	productoproveedor "api/internal/productoProveedor"
	"api/internal/proveedor"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrdenDeCompraService struct {
	ordenCompraRepositorio       OrdenCompraRepositorio
	ordenCompraItemRepositorio   ordendecompraitem.OrdenCompraItemRepositorio
	proveedorRepositorio         proveedor.ProveedorRepositorio
	depositoRepositorio          deposito.DepositoRepositorio
	productoProveedorRepositorio productoproveedor.ProductoProveedorRepositorio
}

func NuevoOrdenDeCompraervice(
	ordenCompraRepositorio OrdenCompraRepositorio,
	ordenCompraItemRepositorio ordendecompraitem.OrdenCompraItemRepositorio,
	proveedorRepositorio proveedor.ProveedorRepositorio,
	depositoRepositorio deposito.DepositoRepositorio,
) *OrdenDeCompraService {
	return &OrdenDeCompraService{
		ordenCompraRepositorio:     ordenCompraRepositorio,
		ordenCompraItemRepositorio: ordenCompraItemRepositorio,
		proveedorRepositorio:       proveedorRepositorio,
		depositoRepositorio:        depositoRepositorio,
	}
}

func (s *OrdenDeCompraService) CrearOrdenCompra(
	ctx context.Context,
	dto CrearOrdenCompraDTO,
	idAdmin string,
) (OrdenCompraDTO, error) {

	orden, err := MapeoCrearDTOAOrdenCompra(dto)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// Verificar que el proveedor exista.
	_, err = s.proveedorRepositorio.EncontrarPorId(
		ctx,
		dto.ProveedorID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// Verificar que el depósito exista.
	_, err = s.depositoRepositorio.EncontrarPorId(
		ctx,
		dto.DepositoID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	ahora := time.Now()

	orden.Auditoria.CreadoPor = oidAdmin
	orden.Auditoria.CreadoEn = ahora

	ordenCreada, err := s.ordenCompraRepositorio.CrearOrdenCompra(
		ctx,
		orden,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	return MapeoAOrdenCompraDTO(ordenCreada), nil
}

func (s *OrdenDeCompraService) EncontrarPorId(
	ctx context.Context,
	id string,
) (OrdenCompraDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		id,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
		ctx,
		id,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	resultado := MapeoAOrdenCompraDTO(orden)

	resultado.Items = make(
		[]ordendecompraitem.OrdenCompraItemDTO,
		0,
		len(items),
	)

	for i := range items {
		itemDTO := ordendecompraitem.MapeoAOrdenCompraItemDTO(
			&items[i],
		)

		resultado.Items = append(
			resultado.Items,
			itemDTO,
		)
	}

	return resultado, nil
}

func (s *OrdenDeCompraService) EncontrarTodos(
	ctx context.Context,
) ([]OrdenCompraDTO, error) {

	ordenes, err := s.ordenCompraRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return nil, err
	}

	resultado := make([]OrdenCompraDTO, 0, len(ordenes))

	for i := range ordenes {

		ordenDTO := MapeoAOrdenCompraDTO(ordenes[i])

		items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
			ctx,
			ordenes[i].ID.Hex(),
		)
		if err != nil {
			return nil, err
		}

		ordenDTO.Items = make(
			[]ordendecompraitem.OrdenCompraItemDTO,
			0,
			len(items),
		)

		for j := range items {

			itemDTO := ordendecompraitem.MapeoAOrdenCompraItemDTO(
				&items[j],
			)

			ordenDTO.Items = append(
				ordenDTO.Items,
				itemDTO,
			)
		}

		resultado = append(resultado, ordenDTO)
	}

	return resultado, nil
}

func (s *OrdenDeCompraService) ActualizarOrdenCompra(
	ctx context.Context,
	dto ActualizarOrdenCompraDTO,
	idAdmin string,
) (OrdenCompraDTO, error) {

	ordenActual, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		dto.ID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if ordenActual.Estado != OrdenBorrador {
		return OrdenCompraDTO{}, errors.New(
			"solo se puede modificar una orden de compra en borrador",
		)
	}

	nuevaOrden, err := MapeoActualizarDTOAOrdenCompra(dto)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// Verificar que el proveedor exista.
	_, err = s.proveedorRepositorio.EncontrarPorId(
		ctx,
		dto.ProveedorID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// Verificar que el depósito exista.
	_, err = s.depositoRepositorio.EncontrarPorId(
		ctx,
		dto.DepositoID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	ahora := time.Now()

	nuevaOrden.Auditoria.ModificadoPor = &oidAdmin
	nuevaOrden.Auditoria.ActualizadoEn = &ahora
	nuevaOrden.Estado = ordenActual.Estado

	// Si cambió el proveedor, se eliminan todos los ítems
	// porque los productos y sus costos dependen del proveedor.
	if ordenActual.ProveedorID != nuevaOrden.ProveedorID {

		err = s.ordenCompraItemRepositorio.EliminarPorOrdenCompra(
			ctx,
			dto.ID,
		)
		if err != nil {
			return OrdenCompraDTO{}, err
		}
	}

	ordenActualizada, err := s.ordenCompraRepositorio.ActualizarOrdenCompra(
		ctx,
		dto.ID,
		nuevaOrden,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	return MapeoAOrdenCompraDTO(ordenActualizada), nil
}

func (s *OrdenDeCompraService) ConfirmarOrdenCompra(
	ctx context.Context,
	idUsuario string,
	id string,
) (OrdenCompraDTO, error) {

	oidUsuario, err := bson.ObjectIDFromHex(idUsuario)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		id,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if orden.Estado != OrdenBorrador {
		return OrdenCompraDTO{}, errors.New(
			"solo se pueden confirmar órdenes en borrador",
		)
	}

	items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
		ctx,
		id,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if len(items) == 0 {
		return OrdenCompraDTO{}, errors.New(
			"no se puede confirmar una orden sin productos",
		)
	}

	ahora := time.Now()

	orden.Estado = OrdenConfirmada
	orden.Auditoria.ModificadoPor = &oidUsuario
	orden.Auditoria.ActualizadoEn = &ahora

	ordenActualizada, err := s.ordenCompraRepositorio.ActualizarOrdenCompra(
		ctx,
		id,
		orden,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	resultado := MapeoAOrdenCompraDTO(ordenActualizada)

	resultado.Items = make(
		[]ordendecompraitem.OrdenCompraItemDTO,
		0,
		len(items),
	)

	for i := range items {

		itemDTO := ordendecompraitem.MapeoAOrdenCompraItemDTO(
			&items[i],
		)

		resultado.Items = append(
			resultado.Items,
			itemDTO,
		)
	}

	return resultado, nil
}

func (s *OrdenDeCompraService) DescartarOrdenCompra(
	ctx context.Context,
	idUsuario string,
	id string,
) error {

	oidUsuario, err := bson.ObjectIDFromHex(idUsuario)
	if err != nil {
		return err
	}

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		id,
	)
	if err != nil {
		return err
	}

	if orden.Estado != OrdenBorrador {
		return errors.New(
			"solo se pueden descartar órdenes en borrador",
		)
	}

	ahora := time.Now()

	orden.Estado = OrdenDescartada
	orden.Auditoria.ModificadoPor = &oidUsuario
	orden.Auditoria.ActualizadoEn = &ahora

	_, err = s.ordenCompraRepositorio.ActualizarOrdenCompra(
		ctx,
		id,
		orden,
	)

	return err
}

func (s *OrdenDeCompraService) AgregarItem(
	ctx context.Context,
	ordenCompraID string,
	dto ordendecompraitem.CrearOrdenCompraItemDTO,
	idAdmin string,
) (ordendecompraitem.OrdenCompraItemDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	if orden.Estado != OrdenBorrador {
		return ordendecompraitem.OrdenCompraItemDTO{}, errors.New(
			"solo se pueden agregar productos a una orden en borrador",
		)
	}

	if dto.CantidadSolicitada <= 0 {
		return ordendecompraitem.OrdenCompraItemDTO{}, errors.New(
			"la cantidad solicitada debe ser mayor a cero",
		)
	}

	productoID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	// Verificar que el proveedor de la orden suministre el producto
	// y obtener el costo correspondiente.
	productoProveedor, err := s.productoProveedorRepositorio.EncontrarUno(
		ctx,
		productoID,
		orden.ProveedorID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	item, err := ordendecompraitem.MapeoCrearDTOAOrdenCompraItem(dto)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	ahora := time.Now()

	item.OrdenCompraID = orden.ID
	item.CostoUnitario = productoProveedor.CostoUnidad
	item.Auditoria.CreadoPor = oidAdmin
	item.Auditoria.CreadoEn = ahora

	itemCreado, err := s.ordenCompraItemRepositorio.CrearOrdenCompraItem(
		ctx,
		item,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	return ordendecompraitem.MapeoAOrdenCompraItemDTO(&itemCreado), nil
}

func (s *OrdenDeCompraService) ActualizarItem(
	ctx context.Context,
	ordenCompraID string,
	dto ordendecompraitem.ActualizarOrdenCompraItemDTO,
	idAdmin string,
) (ordendecompraitem.OrdenCompraItemDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	if orden.Estado != OrdenBorrador {
		return ordendecompraitem.OrdenCompraItemDTO{}, errors.New(
			"solo se pueden modificar productos de una orden en borrador",
		)
	}

	if dto.CantidadSolicitada <= 0 {
		return ordendecompraitem.OrdenCompraItemDTO{}, errors.New(
			"la cantidad solicitada debe ser mayor a cero",
		)
	}

	itemActual, err := s.ordenCompraItemRepositorio.EncontrarPorId(
		ctx,
		dto.ID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	// Evita modificar un item que pertenece a otra orden.
	if itemActual.OrdenCompraID != orden.ID {
		return ordendecompraitem.OrdenCompraItemDTO{}, errors.New(
			"el item no pertenece a la orden de compra",
		)
	}

	productoID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	// Verificar que el nuevo producto sea suministrado
	// por el proveedor actual de la orden.
	productoProveedor, err := s.productoProveedorRepositorio.EncontrarUno(
		ctx,
		productoID,
		orden.ProveedorID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	item, err := ordendecompraitem.MapeoActualizarDTOAOrdenCompraItem(dto)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	ahora := time.Now()

	item.ID = itemActual.ID
	item.OrdenCompraID = itemActual.OrdenCompraID
	item.CantidadRecibida = itemActual.CantidadRecibida
	item.CostoUnitario = productoProveedor.CostoUnidad

	item.Auditoria = itemActual.Auditoria
	item.Auditoria.ModificadoPor = &oidAdmin
	item.Auditoria.ActualizadoEn = &ahora

	itemActualizado, err := s.ordenCompraItemRepositorio.ActualizarOrdenCompraItem(
		ctx,
		dto.ID,
		item,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	return ordendecompraitem.MapeoAOrdenCompraItemDTO(&itemActualizado), nil
}

// Revisar si hacer eliminación fisica o logica. Deberia ser logica pero aca es fisica.
func (s *OrdenDeCompraService) EliminarItem(
	ctx context.Context,
	ordenCompraID string,
	itemID string,
) error {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return err
	}

	if orden.Estado != OrdenBorrador {
		return errors.New(
			"solo se pueden eliminar productos de una orden en borrador",
		)
	}

	item, err := s.ordenCompraItemRepositorio.EncontrarPorId(
		ctx,
		itemID,
	)
	if err != nil {
		return err
	}

	if item.OrdenCompraID != orden.ID {
		return errors.New(
			"el item no pertenece a la orden de compra",
		)
	}

	return s.ordenCompraItemRepositorio.EliminarOrdenCompraItem(
		ctx,
		itemID,
	)
}

func (s *OrdenDeCompraService) RegistrarRecepcion(
	ctx context.Context,
	ordenCompraID string,
	dto RegistrarRecepcionDTO,
	idAdmin string,
) (OrdenCompraDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if orden.Estado != OrdenConfirmada &&
		orden.Estado != OrdenParcial {

		return OrdenCompraDTO{}, errors.New(
			"solo se puede recibir mercadería de una orden confirmada o parcial",
		)
	}

	recepcion, err := MapeoRegistrarRecepcionDTO(dto)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// Validamos TODA la recepción antes de modificar cualquier item.
	for _, recepcionItem := range recepcion.Items {

		var item *ordendecompraitem.OrdenCompraItem

		for i := range items {
			if items[i].ID == recepcionItem.ItemID {
				item = &items[i]
				break
			}
		}

		if item == nil {
			return OrdenCompraDTO{}, errors.New(
				"uno de los items no pertenece a la orden de compra",
			)
		}

		if item.CantidadRecibida+recepcionItem.Cantidad >
			item.CantidadSolicitada {

			return OrdenCompraDTO{}, errors.New(
				"la cantidad recibida no puede superar la cantidad solicitada",
			)
		}
	}

	// Si todas las cantidades son válidas,
	// registramos las recepciones.
	for _, recepcionItem := range recepcion.Items {

		err := s.ordenCompraItemRepositorio.RegistrarRecepcion(
			ctx,
			recepcionItem.ItemID.Hex(),
			recepcionItem.Cantidad,
			idAdmin,
		)
		if err != nil {
			return OrdenCompraDTO{}, err
		}
	}

	// Volvemos a consultar los items para determinar
	// el nuevo estado de la orden.
	itemsActualizados, err :=
		s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
			ctx,
			ordenCompraID,
		)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	todosCompletos := true

	for _, item := range itemsActualizados {

		if item.CantidadRecibida < item.CantidadSolicitada {
			todosCompletos = false
			break
		}
	}

	nuevoEstado := OrdenParcial

	if todosCompletos {
		nuevoEstado = OrdenCompletada
	}

	if orden.Estado != nuevoEstado {

		orden.Estado = nuevoEstado

		ordenActualizada, err :=
			s.ActualizarEstadoOrdenCompra(
				ctx,
				orden,
			)
		if err != nil {
			return OrdenCompraDTO{}, err
		}

		orden = ordenActualizada
	}

	return MapeoAOrdenCompraDTO(orden), nil
}

func (s *OrdenDeCompraService) ActualizarEstadoOrdenCompra(
	ctx context.Context,
	orden OrdenCompra,
) (OrdenCompra, error) {

	err := s.ordenCompraRepositorio.ActualizarEstadoOrdenCompra(
		ctx,
		orden.ID.Hex(),
		orden.Estado,
	)
	if err != nil {
		return OrdenCompra{}, err
	}

	return orden, nil
}
