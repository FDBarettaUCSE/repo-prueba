package ordendecompra

import (
	"api/internal/auditoria"
	"api/internal/deposito"
	depositoproducto "api/internal/depositoProducto"
	evaluacionstock "api/internal/evaluacionStock"
	movimientostock "api/internal/movimientoStock"
	ordendecompraitem "api/internal/ordenDeCompraItem"
	"api/internal/producto"
	productoproveedor "api/internal/productoProveedor"
	"api/internal/proveedor"
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxTamanioPagina = 100

var (
	ErrDepositoNoPermitido = errors.New(
		"no tenés autorización para operar sobre este depósito",
	)

	ErrEstadoOrdenInvalido = errors.New(
		"la orden de compra no se encuentra en el estado requerido para esta acción",
	)

	ErrOrdenSinProductos = errors.New(
		"no se puede confirmar una orden sin productos",
	)

	ErrCantidadSolicitadaInvalida = errors.New(
		"la cantidad solicitada debe ser mayor a cero",
	)

	ErrItemNoPerteneceOrden = errors.New(
		"el item no pertenece a la orden de compra",
	)

	ErrCantidadRecibidaExcedeSolicitada = errors.New(
		"la cantidad recibida no puede superar la cantidad solicitada",
	)

	ErrDepositoNoPermitidoActualizar = errors.New(
		"deposito no permitido",
	)

	ErrItemRecepcionNoPerteneceOrden = errors.New(
		"uno de los items no pertenece a la orden de compra",
	)

	ErrOrdenCompraInexistente = errors.New(
		"la orden de compra solicitada no existe",
	)

	ErrIDOrdenCompraInvalido = errors.New(
		"el id de la orden de compra no es válido",
	)

	ErrProveedorInexistente = errors.New(
		"el proveedor indicado no existe",
	)

	ErrDepositoInexistente = errors.New(
		"el depósito indicado no existe",
	)

	ErrDatosOrdenInvalidos = errors.New(
		"los datos de la orden de compra no son válidos",
	)

	ErrRecepcionInvalida = errors.New(
		"los datos de la recepción no son válidos",
	)

	ErrProductoNoSuministrado = productoproveedor.ErrProductoSuministradoInexistente
)

type OrdenDeCompraService struct {
	ordenCompraRepositorio       OrdenCompraRepositorio
	ordenCompraItemRepositorio   ordendecompraitem.OrdenCompraItemRepositorio
	proveedorRepositorio         proveedor.ProveedorRepositorio
	depositoRepositorio          deposito.DepositoRepositorio
	productoProveedorRepositorio productoproveedor.ProductoProveedorRepositorio
	movimientoStockRepositorio   movimientostock.MovimientoStockRepositorio
	depositoProductoRepositorio  depositoproducto.DepositoProductoRepositorio
	productoRepositorio          producto.ProductoRepositorio
	evaluadorStock               evaluacionstock.EvaluadorStock
}

func NuevoOrdenDeCompraService(
	ordenCompraRepositorio OrdenCompraRepositorio,
	ordenCompraItemRepositorio ordendecompraitem.OrdenCompraItemRepositorio,
	proveedorRepositorio proveedor.ProveedorRepositorio,
	depositoRepositorio deposito.DepositoRepositorio,
	productoProveedorRepositorio productoproveedor.ProductoProveedorRepositorio,
	movimientoStockRepositorio movimientostock.MovimientoStockRepositorio,
	depositoProductoRepositorio depositoproducto.DepositoProductoRepositorio,
	productoRepositorio producto.ProductoRepositorio,
	evaluadorStock evaluacionstock.EvaluadorStock,
) *OrdenDeCompraService {

	return &OrdenDeCompraService{
		ordenCompraRepositorio:       ordenCompraRepositorio,
		ordenCompraItemRepositorio:   ordenCompraItemRepositorio,
		proveedorRepositorio:         proveedorRepositorio,
		depositoRepositorio:          depositoRepositorio,
		productoProveedorRepositorio: productoProveedorRepositorio,
		movimientoStockRepositorio:   movimientoStockRepositorio,
		depositoProductoRepositorio:  depositoProductoRepositorio,
		productoRepositorio:          productoRepositorio,
		evaluadorStock:               evaluadorStock,
	}
}

func (s *OrdenDeCompraService) CrearOrdenCompra(
	ctx context.Context,
	dto CrearOrdenCompraDTO,
	idUsuario, rol, depUsuario string,
) (OrdenCompraDTO, error) {

	if rol == "gerente" && dto.DepositoID != depUsuario {
		return OrdenCompraDTO{}, ErrDepositoNoPermitido
	}

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
		return OrdenCompraDTO{}, ErrProveedorInexistente
	}

	// Verificar que el depósito exista.
	_, err = s.depositoRepositorio.EncontrarPorId(
		ctx,
		dto.DepositoID,
	)
	if err != nil {
		return OrdenCompraDTO{}, ErrDepositoInexistente
	}

	oidUsuario, err := bson.ObjectIDFromHex(idUsuario)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	ahora := time.Now()

	orden.Auditoria.CreadoPor = oidUsuario
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
	id, rol, depUsuario string,
) (OrdenCompraDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		id,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if (rol == "gerente" || rol == "operario") && depUsuario != orden.DepositoID.Hex() {
		return OrdenCompraDTO{}, ErrDepositoNoPermitido
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
	rol, depUsuario string,
) ([]OrdenCompraDTO, error) {

	var ordenes []OrdenCompra
	var err error

	if rol == "gerente" || rol == "operario" {
		depOID, errOID := bson.ObjectIDFromHex(depUsuario)
		if errOID != nil {
			return nil, errOID
		}
		ordenes, err = s.ordenCompraRepositorio.EncontrarPorDeposito(ctx, depOID)
	} else {
		ordenes, err = s.ordenCompraRepositorio.EncontrarTodos(ctx)
	}
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

func (s *OrdenDeCompraService) Listar(
	ctx context.Context,
	rol, depUsuario string,
	filtro FiltroOrdenesCompraDTO,
) (OrdenesCompraPaginadasDTO, error) {

	depositoPermitido := ""
	if rol == "gerente" || rol == "operario" {
		depositoPermitido = depUsuario
	}

	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 20
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	ordenes, total, err := s.ordenCompraRepositorio.Listar(ctx, filtro, depositoPermitido)
	if err != nil {
		return OrdenesCompraPaginadasDTO{}, err
	}

	pagina := filtro.Pagina
	if pagina < 1 {
		pagina = 1
	}
	tamanioPagina := filtro.TamanioPagina

	resultado := make([]OrdenCompraDTO, 0, len(ordenes))

	for i := range ordenes {
		ordenDTO := MapeoAOrdenCompraDTO(ordenes[i])

		items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
			ctx,
			ordenes[i].ID.Hex(),
		)
		if err != nil {
			return OrdenesCompraPaginadasDTO{}, err
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

	return OrdenesCompraPaginadasDTO{
		Items:         resultado,
		Total:         total,
		Pagina:        pagina,
		TamanioPagina: tamanioPagina,
	}, nil
}

// GenerarSugerenciasPorStockBajoMinimo escanea los productos cuyo stock se encuentra
// por debajo del mínimo para un depósito y genera automáticamente órdenes de compra
// sugeridas en estado borrador con el proveedor de menor costo.
func (s *OrdenDeCompraService) GenerarSugerenciasPorStockBajoMinimo(
	ctx context.Context,
	rol, depUsuario, depositoID, idUsuario string,
) ([]OrdenCompraDTO, error) {

	depositoObjetivo := depositoID
	if rol == "gerente" || rol == "operario" {
		if depositoID != "" && depositoID != depUsuario {
			return nil, ErrDepositoNoPermitido
		}
		depositoObjetivo = depUsuario
	}

	var depositosAProcesar []deposito.Deposito
	if depositoObjetivo != "" {
		dep, err := s.depositoRepositorio.EncontrarPorId(ctx, depositoObjetivo)
		if err != nil {
			return nil, ErrDepositoInexistente
		}
		depositosAProcesar = []deposito.Deposito{dep}
	} else {
		todos, err := s.depositoRepositorio.EncontrarTodos(ctx)
		if err != nil {
			return nil, err
		}
		depositosAProcesar = todos
	}

	oidUsuario, err := bson.ObjectIDFromHex(idUsuario)
	if err != nil {
		return nil, err
	}

	ordenesModificadas := make(map[string]bool)

	for _, dep := range depositosAProcesar {
		filasStock, err := s.depositoProductoRepositorio.EncontrarPorDeposito(ctx, dep.ID)
		if err != nil {
			continue
		}

		// Buscar órdenes activas en el depósito para no duplicar pedidos
		ordenesDep, err := s.ordenCompraRepositorio.EncontrarPorDeposito(ctx, dep.ID)
		if err != nil {
			continue
		}

		productosEnOrdenActiva := make(map[bson.ObjectID]bool)
		ordenesBorradorPorProveedor := make(map[bson.ObjectID]OrdenCompra)

		for _, o := range ordenesDep {
			if o.Estado == OrdenBorrador {
				ordenesBorradorPorProveedor[o.ProveedorID] = o
			}
			if o.Estado == OrdenBorrador || o.Estado == OrdenConfirmada || o.Estado == OrdenParcial {
				items, errI := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(ctx, o.ID.Hex())
				if errI == nil {
					for _, item := range items {
						productosEnOrdenActiva[item.ProductoID] = true
					}
				}
			}
		}

		for _, sf := range filasStock {
			if productosEnOrdenActiva[sf.ProductoID] {
				continue // Ya tiene una orden activa
			}

			prod, err := s.productoRepositorio.EncontrarPorId(ctx, sf.ProductoID.Hex())
			if err != nil {
				continue
			}

			stockMinimo := prod.StockMinimo
			if sf.StockMinimo != nil {
				stockMinimo = *sf.StockMinimo
			}

			if sf.Stock >= stockMinimo {
				continue // Stock adecuado
			}

			// Stock bajo mínimo -> Buscar proveedores asociados a este producto
			proveedoresProd, err := s.productoProveedorRepositorio.EncontrarPorProducto(ctx, prod.ID)
			if err != nil || len(proveedoresProd) == 0 {
				continue // No tiene proveedor asociado
			}

			// Elegir el de menor costo
			var mejorVinculo *productoproveedor.ProductoProveedor
			for i := range proveedoresProd {
				if mejorVinculo == nil || proveedoresProd[i].CostoUnidad < mejorVinculo.CostoUnidad {
					mejorVinculo = &proveedoresProd[i]
				}
			}

			cantidadReponer := stockMinimo - sf.Stock
			if cantidadReponer <= 0 {
				cantidadReponer = 1
			}

			// Obtener o crear orden en borrador para este proveedor en este depósito
			ordenBorrador, existe := ordenesBorradorPorProveedor[mejorVinculo.ProveedorID]
			if !existe {
				nuevaOrden := OrdenCompra{
					ProveedorID: mejorVinculo.ProveedorID,
					DepositoID:  dep.ID,
					Estado:      OrdenBorrador,
					Auditoria: auditoria.Auditoria{
						CreadoPor: oidUsuario,
						CreadoEn:  time.Now(),
					},
				}
				creada, errCrear := s.ordenCompraRepositorio.CrearOrdenCompra(ctx, nuevaOrden)
				if errCrear != nil {
					continue
				}
				ordenBorrador = creada
				ordenesBorradorPorProveedor[mejorVinculo.ProveedorID] = creada
			}

			// Crear item en la orden borrador
			ahora := time.Now()
			nuevoItem := ordendecompraitem.OrdenCompraItem{
				OrdenCompraID:      ordenBorrador.ID,
				ProductoID:         prod.ID,
				CantidadSolicitada: cantidadReponer,
				CantidadRecibida:   0,
				CostoUnitario:      mejorVinculo.CostoUnidad,
				Auditoria: auditoria.Auditoria{
					CreadoPor: oidUsuario,
					CreadoEn:  ahora,
				},
			}

			_, errItem := s.ordenCompraItemRepositorio.CrearOrdenCompraItem(ctx, nuevoItem)
			if errItem == nil {
				productosEnOrdenActiva[prod.ID] = true
				ordenesModificadas[ordenBorrador.ID.Hex()] = true
			}
		}
	}

	resultado := make([]OrdenCompraDTO, 0, len(ordenesModificadas))
	for idOrden := range ordenesModificadas {
		dto, err := s.EncontrarPorId(ctx, idOrden, "administrador", "")
		if err == nil {
			resultado = append(resultado, dto)
		}
	}

	return resultado, nil
}

func (s *OrdenDeCompraService) ActualizarOrdenCompra(
	ctx context.Context,
	id string,
	dto ActualizarOrdenCompraDTO,
	idUsuario string,
	rol string,
	depUsuario string,
) (OrdenCompraDTO, error) {

	ordenActual, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		dto.ID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if err := s.validarAccesoDepositoOrden(ordenActual, rol, depUsuario); err != nil {
		return OrdenCompraDTO{}, err
	}

	if ordenActual.Estado != OrdenBorrador {
		return OrdenCompraDTO{}, ErrEstadoOrdenInvalido
	}

	// Un gerente tampoco puede cambiar la orden
	// para que pase a pertenecer a otro depósito.
	if rol == "gerente" && dto.DepositoID != depUsuario {
		return OrdenCompraDTO{}, ErrDepositoNoPermitidoActualizar
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
		return OrdenCompraDTO{}, ErrProveedorInexistente
	}

	// Verificar que el depósito exista.
	_, err = s.depositoRepositorio.EncontrarPorId(
		ctx,
		dto.DepositoID,
	)
	if err != nil {
		return OrdenCompraDTO{}, ErrDepositoInexistente
	}

	oidAdmin, err := bson.ObjectIDFromHex(idUsuario)
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
		id,
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
	rol string,
	depUsuario string,
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

	if err := s.validarAccesoDepositoOrden(orden, rol, depUsuario); err != nil {
		return OrdenCompraDTO{}, err
	}

	if orden.Estado != OrdenBorrador {
		return OrdenCompraDTO{}, ErrEstadoOrdenInvalido
	}

	items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
		ctx,
		id,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if len(items) == 0 {
		return OrdenCompraDTO{}, ErrOrdenSinProductos
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
	rol string,
	depUsuario string,
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

	if err := s.validarAccesoDepositoOrden(orden, rol, depUsuario); err != nil {
		return err
	}

	if orden.Estado != OrdenBorrador {
		return ErrEstadoOrdenInvalido
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
	idUsuario string,
	rol string,
	depUsuario string,
) (ordendecompraitem.OrdenCompraItemDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	if err := s.validarAccesoDepositoOrden(orden, rol, depUsuario); err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	if orden.Estado != OrdenBorrador {
		return ordendecompraitem.OrdenCompraItemDTO{}, ErrEstadoOrdenInvalido
	}

	if dto.CantidadSolicitada <= 0 {
		return ordendecompraitem.OrdenCompraItemDTO{}, ErrCantidadSolicitadaInvalida
	}

	productoID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, fmt.Errorf("%w: id de producto inválido", ErrDatosOrdenInvalidos)
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
		return ordendecompraitem.OrdenCompraItemDTO{}, fmt.Errorf("%w: %v", ErrDatosOrdenInvalidos, err)
	}

	oidUsuario, err := bson.ObjectIDFromHex(idUsuario)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	ahora := time.Now()

	item.OrdenCompraID = orden.ID
	item.CostoUnitario = productoProveedor.CostoUnidad
	item.Auditoria.CreadoPor = oidUsuario
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
	idUsuario string,
	rol string,
	depUsuario string,
) (ordendecompraitem.OrdenCompraItemDTO, error) {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	if err := s.validarAccesoDepositoOrden(orden, rol, depUsuario); err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	if orden.Estado != OrdenBorrador {
		return ordendecompraitem.OrdenCompraItemDTO{}, ErrEstadoOrdenInvalido
	}

	if dto.CantidadSolicitada <= 0 {
		return ordendecompraitem.OrdenCompraItemDTO{}, ErrCantidadSolicitadaInvalida
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
		return ordendecompraitem.OrdenCompraItemDTO{}, ErrItemNoPerteneceOrden
	}

	productoID, err := bson.ObjectIDFromHex(dto.ProductoID)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, fmt.Errorf("%w: id de producto inválido", ErrDatosOrdenInvalidos)
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
		return ordendecompraitem.OrdenCompraItemDTO{}, fmt.Errorf("%w: %v", ErrDatosOrdenInvalidos, err)
	}

	oidUsuario, err := bson.ObjectIDFromHex(idUsuario)
	if err != nil {
		return ordendecompraitem.OrdenCompraItemDTO{}, err
	}

	ahora := time.Now()

	item.ID = itemActual.ID
	item.OrdenCompraID = itemActual.OrdenCompraID
	item.CantidadRecibida = itemActual.CantidadRecibida
	item.CostoUnitario = productoProveedor.CostoUnidad

	item.Auditoria = itemActual.Auditoria
	item.Auditoria.ModificadoPor = &oidUsuario
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
	rol string,
	depUsuario string,
) error {

	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return err
	}

	if err := s.validarAccesoDepositoOrden(orden, rol, depUsuario); err != nil {
		return err
	}

	if orden.Estado != OrdenBorrador {
		return ErrEstadoOrdenInvalido
	}

	item, err := s.ordenCompraItemRepositorio.EncontrarPorId(
		ctx,
		itemID,
	)
	if err != nil {
		return err
	}

	if item.OrdenCompraID != orden.ID {
		return ErrItemNoPerteneceOrden
	}

	return s.ordenCompraItemRepositorio.EliminarOrdenCompraItem(
		ctx,
		itemID,
	)
}

// Revisar si hacer eliminación fisica o logica. Deberia ser logica pero aca es fisica.

func (s *OrdenDeCompraService) RegistrarRecepcion(
	ctx context.Context,
	ordenCompraID string,
	dto RegistrarRecepcionDTO,
	idAdmin string,
	rol string,
	depUsuario string,
) (OrdenCompraDTO, error) {

	// 1. Buscar y validar la orden.
	orden, err := s.ordenCompraRepositorio.EncontrarPorId(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	if err := s.validarAccesoDepositoOrden(orden, rol, depUsuario); err != nil {
		return OrdenCompraDTO{}, err
	}

	if orden.Estado != OrdenConfirmada && orden.Estado != OrdenParcial {
		return OrdenCompraDTO{}, ErrEstadoOrdenInvalido
	}

	// 2. Mapear la recepción.
	recepcion, err := MapeoRegistrarRecepcionDTO(dto)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// 3. Obtener y validar los items antes de modificar datos.
	items, err := s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
		ctx,
		ordenCompraID,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	itemsPorID, err := s.validarRecepcion(items, recepcion)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// 4. Convertir el usuario responsable.
	usuarioOID, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// 5. Aplicar la recepción.
	aplicados, stocksActualizados, err := s.aplicarRecepcion(
		ctx,
		orden,
		recepcion,
		itemsPorID,
		usuarioOID,
		idAdmin,
	)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	// 6. Actualizar el estado de la orden.
	if err := s.actualizarEstadoRecepcion(
		ctx,
		orden,
		ordenCompraID,
		aplicados,
		usuarioOID,
		idAdmin,
	); err != nil {
		return OrdenCompraDTO{}, err
	}

	// 7. Evaluar alertas una vez finalizada correctamente la recepción.
	// Es una operación secundaria: sus errores no revierten la recepción.
	s.evaluarAlertasRecepcion(
		ctx,
		orden.DepositoID,
		stocksActualizados,
		usuarioOID,
	)

	// 8. Construir la respuesta.
	return s.construirOrdenCompraDTO(ctx, ordenCompraID)
}

// Validación de la recepción
// Devuelve un mapa para no tener que volver a recorrer todos los ítems cada vez que se procesa uno.
func (s *OrdenDeCompraService) validarRecepcion(
	items []ordendecompraitem.OrdenCompraItem,
	recepcion Recepcion,
) (map[bson.ObjectID]ordendecompraitem.OrdenCompraItem, error) {

	itemsPorID := make(
		map[bson.ObjectID]ordendecompraitem.OrdenCompraItem,
		len(items),
	)

	for i := range items {
		itemsPorID[items[i].ID] = items[i]
	}

	for _, recepcionItem := range recepcion.Items {
		item, existe := itemsPorID[recepcionItem.ItemID]
		if !existe {
			return nil, ErrItemRecepcionNoPerteneceOrden
		}

		if item.CantidadRecibida+recepcionItem.Cantidad >
			item.CantidadSolicitada {
			return nil, ErrCantidadRecibidaExcedeSolicitada
		}
	}

	return itemsPorID, nil
}

type stockRecepcionAplicado struct {
	productoID bson.ObjectID
	stock      depositoproducto.DepositoProducto
}

// Aplicacion de recepcion
// Esta función concentra las modificaciones de stock, los ítems y los movimientos.
// Si falla alguno de esos pasos, realiza la compensación utilizando tu revertirRecepcion actual.
func (s *OrdenDeCompraService) aplicarRecepcion(
	ctx context.Context,
	orden OrdenCompra,
	recepcion Recepcion,
	itemsPorID map[bson.ObjectID]ordendecompraitem.OrdenCompraItem,
	usuarioOID bson.ObjectID,
	idAdmin string,
) ([]pasoRecepcionAplicado, []stockRecepcionAplicado, error) {

	var aplicados []pasoRecepcionAplicado
	var stocksActualizados []stockRecepcionAplicado

	ahora := time.Now()

	for _, recepcionItem := range recepcion.Items {
		item := itemsPorID[recepcionItem.ItemID]

		// Asegurar que exista el registro producto-depósito.
		_, err := s.depositoProductoRepositorio.AsegurarExiste(
			ctx,
			orden.DepositoID,
			item.ProductoID,
			usuarioOID,
		)
		if err != nil {
			s.revertirRecepcion(
				ctx,
				orden.DepositoID,
				aplicados,
				usuarioOID,
				idAdmin,
			)
			return nil, nil, err
		}

		// Aumentar el stock.
		stockActualizado, err := s.depositoProductoRepositorio.AjustarStock(
			ctx,
			orden.DepositoID,
			item.ProductoID,
			recepcionItem.Cantidad,
			usuarioOID,
		)
		if err != nil {
			s.revertirRecepcion(
				ctx,
				orden.DepositoID,
				aplicados,
				usuarioOID,
				idAdmin,
			)
			return nil, nil, err
		}

		paso := pasoRecepcionAplicado{
			itemID:     recepcionItem.ItemID.Hex(),
			productoID: item.ProductoID,
			cantidad:   recepcionItem.Cantidad,
		}

		// Actualizar cantidad recibida del item.
		err = s.ordenCompraItemRepositorio.RegistrarRecepcion(
			ctx,
			recepcionItem.ItemID.Hex(),
			recepcionItem.Cantidad,
			idAdmin,
		)
		if err != nil {
			s.revertirRecepcion(
				ctx,
				orden.DepositoID,
				append(aplicados, paso),
				usuarioOID,
				idAdmin,
			)
			return nil, nil, err
		}
		paso.itemActualizado = true

		// Registrar movimiento de ingreso.
		movimiento := movimientostock.MovimientoStock{
			ProductoID:           item.ProductoID,
			DepositoID:           orden.DepositoID,
			Cantidad:             recepcionItem.Cantidad,
			Tipo:                 movimientostock.Ingreso,
			UsuarioResponsableID: usuarioOID,
			FechaHora:            ahora,
			Motivo:               "Recepción de orden de compra",
			Auditoria: auditoria.Auditoria{
				CreadoPor: usuarioOID,
				CreadoEn:  ahora,
			},
		}

		_, err = s.movimientoStockRepositorio.Crear(ctx, movimiento)
		if err != nil {
			s.revertirRecepcion(
				ctx,
				orden.DepositoID,
				append(aplicados, paso),
				usuarioOID,
				idAdmin,
			)
			return nil, nil, err
		}

		paso.movimientoRegistrado = true
		aplicados = append(aplicados, paso)

		stocksActualizados = append(
			stocksActualizados,
			stockRecepcionAplicado{
				productoID: item.ProductoID,
				stock:      stockActualizado,
			},
		)
	}

	return aplicados, stocksActualizados, nil
}

// Esta función consulta los ítems ya actualizados, determina el estado correspondiente y lo persiste.
// Si falla, conserva el comportamiento original: solicita revertir los pasos aplicados.
func (s *OrdenDeCompraService) actualizarEstadoRecepcion(
	ctx context.Context,
	orden OrdenCompra,
	ordenCompraID string,
	aplicados []pasoRecepcionAplicado,
	usuarioOID bson.ObjectID,
	idAdmin string,
) error {

	itemsActualizados, err :=
		s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
			ctx,
			ordenCompraID,
		)
	if err != nil {
		s.revertirRecepcion(
			ctx,
			orden.DepositoID,
			aplicados,
			usuarioOID,
			idAdmin,
		)
		return err
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
		err = s.ordenCompraRepositorio.ActualizarEstadoOrdenCompra(
			ctx,
			ordenCompraID,
			nuevoEstado,
			idAdmin,
		)
		if err != nil {
			s.revertirRecepcion(
				ctx,
				orden.DepositoID,
				aplicados,
				usuarioOID,
				idAdmin,
			)
			return err
		}
	}

	return nil
}

// Se ejecuta después de completar la recepción.
// Se conserva la prioridad del mínimo específico del depósito
// sobre el mínimo general del producto.
func (s *OrdenDeCompraService) evaluarAlertasRecepcion(
	ctx context.Context,
	depositoID bson.ObjectID,
	stocksActualizados []stockRecepcionAplicado,
	usuarioOID bson.ObjectID,
) {

	for _, registro := range stocksActualizados {
		productoActual, err := s.productoRepositorio.EncontrarPorId(
			ctx,
			registro.productoID.Hex(),
		)
		if err != nil {
			continue
		}

		minimoEfectivo := productoActual.StockMinimo

		if registro.stock.StockMinimo != nil {
			minimoEfectivo = *registro.stock.StockMinimo
		}

		_ = s.evaluadorStock.EvaluarStock(
			ctx,
			depositoID,
			registro.productoID,
			registro.stock.Stock,
			minimoEfectivo,
			usuarioOID,
		)
	}
}

func (s *OrdenDeCompraService) construirOrdenCompraDTO(
	ctx context.Context,
	ordenCompraID string,
) (OrdenCompraDTO, error) {

	ordenActualizada, err :=
		s.ordenCompraRepositorio.EncontrarPorId(
			ctx,
			ordenCompraID,
		)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	itemsFinales, err :=
		s.ordenCompraItemRepositorio.EncontrarPorOrdenCompra(
			ctx,
			ordenCompraID,
		)
	if err != nil {
		return OrdenCompraDTO{}, err
	}

	dtoRespuesta := MapeoAOrdenCompraDTO(ordenActualizada)

	dtoRespuesta.Items = make(
		[]ordendecompraitem.OrdenCompraItemDTO,
		0,
		len(itemsFinales),
	)

	for i := range itemsFinales {
		dtoRespuesta.Items = append(
			dtoRespuesta.Items,
			ordendecompraitem.MapeoAOrdenCompraItemDTO(
				&itemsFinales[i],
			),
		)
	}

	return dtoRespuesta, nil
}

func (s *OrdenDeCompraService) ActualizarEstadoOrdenCompra(
	ctx context.Context,
	orden OrdenCompra,
	idUsuario string,
) (OrdenCompra, error) {

	err := s.ordenCompraRepositorio.ActualizarEstadoOrdenCompra(
		ctx,
		orden.ID.Hex(),
		orden.Estado,
		idUsuario,
	)
	if err != nil {
		return OrdenCompra{}, err
	}

	return orden, nil
}

// pasoRecepcionAplicado registra qué se modificó para un item durante una
// recepción, para poder deshacerlo si un paso posterior falla.
type pasoRecepcionAplicado struct {
	itemID               string
	productoID           bson.ObjectID
	cantidad             int32
	itemActualizado      bool
	movimientoRegistrado bool
}

// revertirRecepcion compensa, en orden inverso, los pasos ya aplicados de
// una recepción que falló a mitad de camino: descuenta el stock sumado,
// resta la cantidad recibida del item y, si el movimiento de ingreso ya
// había quedado registrado, asienta un movimiento inverso (el historial es
// inmutable, no se borra). Si alguna compensación falla se loguea como
// crítico, porque stock, items y movimientos quedan inconsistentes.
func (s *OrdenDeCompraService) revertirRecepcion(
	ctx context.Context,
	depositoID bson.ObjectID,
	aplicados []pasoRecepcionAplicado,
	usuarioOID bson.ObjectID,
	idUsuario string,
) {
	for i := len(aplicados) - 1; i >= 0; i-- {
		p := aplicados[i]

		if p.movimientoRegistrado {
			ahora := time.Now()
			compensatorio := movimientostock.MovimientoStock{
				ProductoID:           p.productoID,
				DepositoID:           depositoID,
				Cantidad:             -p.cantidad,
				Tipo:                 movimientostock.Ajuste,
				UsuarioResponsableID: usuarioOID,
				FechaHora:            ahora,
				Motivo:               "Reversión de recepción de orden de compra fallida",
				Auditoria: auditoria.Auditoria{
					CreadoPor: usuarioOID,
					CreadoEn:  ahora,
				},
			}
			if _, err := s.movimientoStockRepositorio.Crear(ctx, compensatorio); err != nil {
				log.Printf("CRÍTICO: no se pudo registrar el movimiento compensatorio del item %s: %v", p.itemID, err)
			}
		}

		if p.itemActualizado {
			if err := s.ordenCompraItemRepositorio.RegistrarRecepcion(ctx, p.itemID, -p.cantidad, idUsuario); err != nil {
				log.Printf("CRÍTICO: no se pudo revertir la cantidad recibida del item %s: %v", p.itemID, err)
			}
		}

		if _, err := s.depositoProductoRepositorio.AjustarStock(ctx, depositoID, p.productoID, -p.cantidad, usuarioOID); err != nil {
			log.Printf(
				"CRÍTICO: no se pudo revertir el stock (depósito %s, producto %s, cantidad %d): %v",
				depositoID.Hex(), p.productoID.Hex(), p.cantidad, err,
			)
		}
	}
}

func (s *OrdenDeCompraService) validarAccesoDepositoOrden(
	orden OrdenCompra,
	rol string,
	depUsuario string,
) error {
	if rol != "gerente" {
		return nil
	}

	if orden.DepositoID.Hex() != depUsuario {
		return ErrDepositoNoPermitido
	}

	return nil
}
