package producto

import (
	"api/internal/categoria"
	evaluacionstock "api/internal/evaluacionStock"
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxTamanioPagina = 100

var (
	ErrProductoInexistente = errors.New(
		"el producto solicitado no existe",
	)

	ErrCategoriaInexistente = errors.New(
		"la categoría solicitada no existe",
	)

	ErrUnidadMedidaInvalida = errors.New(
		"la unidad de medida ingresada no es válida",
	)

	ErrNombreProductoVacio = errors.New(
		"el nombre del producto es obligatorio",
	)

	ErrCostoUnitarioInvalido = errors.New(
		"el costo unitario no puede ser negativo",
	)

	ErrStockMinimoInvalido = errors.New(
		"el stock mínimo no puede ser negativo",
	)
)

// RegistroStockProducto se hace para evitar la dependencia circular
// entre depositoProducto y producto
type RegistroStockProducto struct {
	DepositoID  bson.ObjectID
	ProductoID  bson.ObjectID
	Stock       int32
	StockMinimo *int32
}

// validarDatosProducto aplica las reglas de negocio comunes al alta y a la
// modificación: nombre informado, costo >= 0 y stock mínimo >= 0.
func validarDatosProducto(nombre string, costoUnitario float64, stockMinimo int32) error {
	if strings.TrimSpace(nombre) == "" {
		return ErrNombreProductoVacio
	}
	if costoUnitario < 0 {
		return ErrCostoUnitarioInvalido
	}
	if stockMinimo < 0 {
		return ErrStockMinimoInvalido
	}
	return nil
}

type ProductoService struct {
	productoRepositorio      ProductoRepositorio
	categoriaRepositorio     categoria.CategoriaRepositorio
	obtenerStocksPorProducto func(context.Context, bson.ObjectID) ([]RegistroStockProducto, error)
	evaluadorStock           evaluacionstock.EvaluadorStock
}

func NuevoProductoService(
	repoProducto ProductoRepositorio,
	repoCategoria categoria.CategoriaRepositorio,
	obtenerStocksPorProducto func(context.Context, bson.ObjectID) ([]RegistroStockProducto, error),
	evaluadorStock evaluacionstock.EvaluadorStock,
) *ProductoService {
	return &ProductoService{
		productoRepositorio:      repoProducto,
		categoriaRepositorio:     repoCategoria,
		obtenerStocksPorProducto: obtenerStocksPorProducto,
		evaluadorStock:           evaluadorStock,
	}
}

func (s *ProductoService) ListarTodosProductos(
	ctx context.Context,
) ([]ProductoDTO, error) {

	productos, err := s.productoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return []ProductoDTO{}, err
	}

	categoriaIDs := []string{}
	categoriaIDsMap := make(map[string]bool)

	for _, producto := range productos {
		id := producto.Categoria.Hex()

		if !categoriaIDsMap[id] {
			categoriaIDs = append(categoriaIDs, id)
			categoriaIDsMap[id] = true
		}
	}

	categorias, err := s.categoriaRepositorio.EncontrarPorIds(
		ctx,
		categoriaIDs,
	)
	if err != nil {
		return []ProductoDTO{}, err
	}

	categoriasMap := make(map[string]string)

	for _, categoria := range categorias {
		categoriasMap[categoria.ID.Hex()] = categoria.Nombre
	}

	productosDTO := make([]ProductoDTO, 0, len(productos))

	for _, producto := range productos {

		categoriaNombre :=
			categoriasMap[producto.Categoria.Hex()]

		productosDTO = append(
			productosDTO,
			MapeoAProductoDTO(
				producto,
				categoriaNombre,
			),
		)
	}

	return productosDTO, nil
}

func (s *ProductoService) ListarProductos(
	ctx context.Context,
	filtro FiltroProductosDTO,
) (ProductosPaginadosDTO, error) {

	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 10
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	productos, total, err :=
		s.productoRepositorio.Listar(ctx, filtro)

	if err != nil {
		return ProductosPaginadosDTO{}, err
	}

	categoriaIDsMap := make(map[string]bool)
	var categoriaIDs []string

	for _, producto := range productos {

		id := producto.Categoria.Hex()

		if !categoriaIDsMap[id] {
			categoriaIDs = append(categoriaIDs, id)
			categoriaIDsMap[id] = true
		}
	}

	categorias, err :=
		s.categoriaRepositorio.EncontrarPorIds(
			ctx,
			categoriaIDs,
		)

	if err != nil {
		return ProductosPaginadosDTO{}, err
	}

	categoriasMap := make(map[string]string)

	for _, categoria := range categorias {
		categoriasMap[categoria.ID.Hex()] =
			categoria.Nombre
	}

	items := make(
		[]ProductoDTO,
		0,
		len(productos),
	)

	for _, producto := range productos {

		categoriaNombre :=
			categoriasMap[producto.Categoria.Hex()]

		items = append(
			items,
			MapeoAProductoDTO(
				producto,
				categoriaNombre,
			),
		)
	}

	pagina := filtro.Pagina

	if pagina < 1 {
		pagina = 1
	}

	return ProductosPaginadosDTO{
		Items:         items,
		Total:         total,
		Pagina:        pagina,
		TamanioPagina: filtro.TamanioPagina,
	}, nil
}

func (s *ProductoService) BuscarProductoPorID(
	ctx context.Context,
	id string,
) (ProductoDTO, error) {

	if _, err := bson.ObjectIDFromHex(id); err != nil {
		return ProductoDTO{}, ErrProductoInexistente
	}

	producto, err :=
		s.productoRepositorio.EncontrarPorId(
			ctx,
			id,
		)

	if err != nil {
		return ProductoDTO{}, ErrProductoInexistente
	}

	categoria, err :=
		s.categoriaRepositorio.EncontrarPorId(
			ctx,
			producto.Categoria.Hex(),
		)

	if err != nil {
		return ProductoDTO{}, ErrCategoriaInexistente
	}

	return MapeoAProductoDTO(
		producto,
		categoria.Nombre,
	), nil
}

func (s *ProductoService) CrearProducto(
	ctx context.Context,
	idAdmin string,
	d CrearProductoDTO,
) (ProductoDTO, error) {

	if err := validarDatosProducto(d.Nombre, d.CostoUnitario, d.StockMinimo); err != nil {
		return ProductoDTO{}, err
	}

	if !esUnidadMedidaValida(d.UnidadMedida) {
		return ProductoDTO{}, ErrUnidadMedidaInvalida
	}

	if _, err := bson.ObjectIDFromHex(d.CategoriaID); err != nil {
		return ProductoDTO{}, ErrCategoriaInexistente
	}

	categoria, err :=
		s.categoriaRepositorio.EncontrarPorId(
			ctx,
			d.CategoriaID,
		)

	if err != nil {
		return ProductoDTO{}, ErrCategoriaInexistente
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProductoDTO{}, err
	}

	producto, err :=
		MapeoCrearDTOAProducto(d)

	if err != nil {
		return ProductoDTO{}, err
	}

	producto.Auditoria.CreadoPor = oidAdmin
	producto.Auditoria.CreadoEn = time.Now()

	producto, err =
		s.productoRepositorio.CrearProducto(
			ctx,
			producto,
		)

	if err != nil {
		return ProductoDTO{}, err
	}

	return MapeoAProductoDTO(
		producto,
		categoria.Nombre,
	), nil
}

func (s *ProductoService) ActualizarProducto(
	ctx context.Context,
	id string,
	idAdmin string,
	d ActualizarProductoDTO,
) (ProductoDTO, error) {

	// La modificación no toca el stock mínimo (tiene su propio endpoint),
	// por eso se valida con 0.
	if err := validarDatosProducto(d.Nombre, d.CostoUnitario, 0); err != nil {
		return ProductoDTO{}, err
	}

	if !esUnidadMedidaValida(d.UnidadMedida) {
		return ProductoDTO{}, ErrUnidadMedidaInvalida
	}

	if _, err := bson.ObjectIDFromHex(id); err != nil {
		return ProductoDTO{}, ErrProductoInexistente
	}

	if _, err := bson.ObjectIDFromHex(d.CategoriaID); err != nil {
		return ProductoDTO{}, ErrCategoriaInexistente
	}

	// Verificamos primero que el producto exista.
	_, err := s.productoRepositorio.EncontrarPorId(
		ctx,
		id,
	)

	if err != nil {
		return ProductoDTO{}, ErrProductoInexistente
	}

	categoria, err :=
		s.categoriaRepositorio.EncontrarPorId(
			ctx,
			d.CategoriaID,
		)

	if err != nil {
		return ProductoDTO{}, ErrCategoriaInexistente
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProductoDTO{}, err
	}

	producto, err :=
		MapeoActualizarDTOAProducto(d)

	if err != nil {
		return ProductoDTO{}, err
	}

	producto.Auditoria.ModificadoPor = &oidAdmin

	ahora := time.Now()
	producto.Auditoria.ActualizadoEn = &ahora

	producto, err =
		s.productoRepositorio.ActualizarProducto(
			ctx,
			id,
			producto,
		)

	if err != nil {
		return ProductoDTO{}, err
	}

	return MapeoAProductoDTO(
		producto,
		categoria.Nombre,
	), nil
}

func (s *ProductoService) ConfigurarStockMinimo(
	ctx context.Context,
	id string,
	idAdmin string,
	d StockMinimoDTO,
) error {
	if d.StockMinimo < 0 {
		return ErrStockMinimoInvalido
	}

	productoID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrProductoInexistente
	}

	usuarioID, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	_, err = s.productoRepositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return ErrProductoInexistente
	}

	err = s.productoRepositorio.ActualizarStockMinimo(ctx, id, d.StockMinimo, idAdmin)
	if err != nil {
		return err
	}

	stocks, err := s.obtenerStocksPorProducto(ctx, productoID)
	if err != nil {
		return err
	}

	for _, stock := range stocks {
		minimoEfectivo := d.StockMinimo

		if stock.StockMinimo != nil {
			minimoEfectivo = *stock.StockMinimo
		}

		err = s.evaluadorStock.EvaluarStock(
			ctx,
			stock.DepositoID,
			stock.ProductoID,
			stock.Stock,
			minimoEfectivo,
			usuarioID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ProductoService) EliminarProducto(
	ctx context.Context,
	id string,
	idAdmin string,
) error {

	if _, err := bson.ObjectIDFromHex(id); err != nil {
		return ErrProductoInexistente
	}

	_, err := s.productoRepositorio.EncontrarPorId(
		ctx,
		id,
	)

	if err != nil {
		return ErrProductoInexistente
	}

	return s.productoRepositorio.EliminarProducto(
		ctx,
		id,
		idAdmin,
	)
}

func esUnidadMedidaValida(
	unidad UnidadMedida,
) bool {

	switch unidad {
	case Unidad, Kilogramo, Litro, Metro:
		return true

	default:
		return false
	}
}
