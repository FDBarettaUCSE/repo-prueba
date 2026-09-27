package producto

import (
	"api/internal/categoria"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductoService struct {
	productoRepositorio  ProductoRepositorio
	categoriaRepositorio categoria.CategoriaRepositorio
}

func NuevoProductoService(repoProducto ProductoRepositorio, repoCategoria categoria.CategoriaRepositorio) *ProductoService {
	return &ProductoService{
		productoRepositorio:  repoProducto,
		categoriaRepositorio: repoCategoria,
	}
}

// Falta agregarle sus respectivos inventarios.
func (s *ProductoService) ListarTodosProductos(ctx context.Context) ([]ProductoDTO, error) {

	productos, err := s.productoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return []ProductoDTO{}, err
	}

	// Obtenemos los IDs de categorías sin repetir.
	categoriaIDs := []string{}
	categoriaIDsMap := make(map[string]bool)

	for _, producto := range productos {
		id := producto.Categoria.Hex()

		if !categoriaIDsMap[id] {
			categoriaIDs = append(categoriaIDs, id)
			categoriaIDsMap[id] = true
		}
	}

	// Buscamos todas las categorías necesarias en una sola consulta.
	categorias, err := s.categoriaRepositorio.EncontrarPorIds(ctx, categoriaIDs)
	if err != nil {
		return []ProductoDTO{}, err
	}

	// Armamos un mapa ID → nombre.
	categoriasMap := make(map[string]string)

	for _, categoria := range categorias {
		categoriasMap[categoria.ID.Hex()] = categoria.Nombre
	}

	// Armamos los DTO.
	productosDTO := []ProductoDTO{}

	for _, producto := range productos {
		categoriaNombre := categoriasMap[producto.Categoria.Hex()]

		productosDTO = append(
			productosDTO,
			MapeoAProductoDTO(producto, categoriaNombre),
		)
	}

	return productosDTO, nil
}
func (s *ProductoService) BuscarProductoPorID(ctx context.Context, id string) (ProductoDTO, error) {

	producto, err := s.productoRepositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return ProductoDTO{}, err
	}

	categoria, err := s.categoriaRepositorio.EncontrarPorId(
		ctx,
		producto.Categoria.Hex(),
	)
	if err != nil {
		return ProductoDTO{}, err
	}

	return MapeoAProductoDTO(producto, categoria.Nombre), nil
}
func (s *ProductoService) CrearProducto(
	ctx context.Context,
	idAdmin string,
	d CrearProductoDTO,
) (ProductoDTO, error) {

	// Primero corroboramos que exista la categoría.
	categoria, err := s.categoriaRepositorio.EncontrarPorId(ctx, d.CategoriaID)
	if err != nil {
		return ProductoDTO{}, err
	}

	producto, err := MapeoCrearDTOAProducto(d)
	if err != nil {
		return ProductoDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProductoDTO{}, err
	}

	producto.Auditoria.CreadoPor = oidAdmin
	producto.Auditoria.CreadoEn = time.Now()

	producto, err = s.productoRepositorio.CrearProducto(ctx, producto)
	if err != nil {
		return ProductoDTO{}, err
	}

	return MapeoAProductoDTO(producto, categoria.Nombre), nil
}

func (s *ProductoService) ActualizarProducto(
	ctx context.Context,
	id, idAdmin string,
	d ActualizarProductoDTO,
) (ProductoDTO, error) {

	// Verificamos que la categoría exista.
	categoria, err := s.categoriaRepositorio.EncontrarPorId(ctx, d.CategoriaID)
	if err != nil {
		return ProductoDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProductoDTO{}, err
	}

	producto, err := MapeoActualizarDTOAProducto(d)
	if err != nil {
		return ProductoDTO{}, err
	}

	producto.Auditoria.ModificadoPor = &oidAdmin

	ahora := time.Now()
	producto.Auditoria.ActualizadoEn = &ahora

	producto, err = s.productoRepositorio.ActualizarProducto(ctx, id, producto)
	if err != nil {
		return ProductoDTO{}, err
	}

	return MapeoAProductoDTO(producto, categoria.Nombre), nil
}

func (s *ProductoService) ConfigurarStockMinimo(
	ctx context.Context,
	id, idAdmin string,
	d StockMinimoDTO,
) error {

	return s.productoRepositorio.ActualizarStockMinimo(
		ctx,
		id,
		d.StockMinimo,
		idAdmin,
	)
}

func (s *ProductoService) EliminarProducto(ctx context.Context, id, idAdmin string) error {
	return s.productoRepositorio.EliminarProducto(ctx, id, idAdmin)
}
