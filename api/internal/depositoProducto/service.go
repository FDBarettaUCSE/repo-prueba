package depositoproducto

import (
	"api/internal/categoria"
	"api/internal/deposito"
	"api/internal/producto"
	"api/internal/usuario"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrAccesoDenegado      = errors.New("no tenés permiso para operar sobre este depósito")
	ErrProductoInexistente = errors.New("el producto ingresado no existe")
)

// Service resuelve el stock por depósito. Depende directamente de los
// repositorios de deposito/producto/categoria/usuario (no de sus Services)
// para enriquecer la respuesta y validar accesos — mismo criterio que ya usa
// proveedor.Service con producto.ProductoRepositorio.
type Service struct {
	repositorio          DepositoProductoRepositorio
	depositoRepositorio  deposito.DepositoRepositorio
	productoRepositorio  producto.ProductoRepositorio
	categoriaRepositorio categoria.CategoriaRepositorio
	usuarioRepositorio   usuario.UsuarioRepositorio
}

func NuevoService(
	repo DepositoProductoRepositorio,
	repoDeposito deposito.DepositoRepositorio,
	repoProducto producto.ProductoRepositorio,
	repoCategoria categoria.CategoriaRepositorio,
	repoUsuario usuario.UsuarioRepositorio,
) *Service {
	return &Service{
		repositorio:          repo,
		depositoRepositorio:  repoDeposito,
		productoRepositorio:  repoProducto,
		categoriaRepositorio: repoCategoria,
		usuarioRepositorio:   repoUsuario,
	}
}

// verificarAcceso confirma que el usuario autenticado puede operar sobre
// depositoID. Administrador y Auditor tienen visibilidad total (el Auditor,
// de solo lectura — eso se controla en las rutas, no acá: acá solo
// chequeamos "sobre qué depósito", nunca "puede escribir"). Gerente y
// Operario solo pueden operar sobre el depósito al que están asignados.
func (s *Service) verificarAcceso(ctx context.Context, usuarioID, rol, depositoID string) error {
	if rol == string(usuario.Administrador) || rol == string(usuario.Auditor) {
		return nil
	}

	u, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return err
	}

	if u.DepositoAsociado.Hex() != depositoID {
		return ErrAccesoDenegado
	}
	return nil
}

// mapearAStockDTO arma la vista enriquecida. dp puede ser el valor cero de
// DepositoProducto (existe == false) cuando todavía no hubo ningún
// movimiento del producto en el depósito: en ese caso el stock es 0 y no hay
// override, pero el producto de todas formas se muestra (con su stock
// mínimo general).
func mapearAStockDTO(dp DepositoProducto, existe bool, dep deposito.Deposito, p producto.Producto, categoriaNombre string) StockDepositoDTO {
	stockMinimoEfectivo := p.StockMinimo
	var override *int32

	if existe && dp.StockMinimo != nil {
		stockMinimoEfectivo = *dp.StockMinimo
		override = dp.StockMinimo
	}

	var stock int32
	if existe {
		stock = dp.Stock
	}

	return StockDepositoDTO{
		DepositoID:          dep.ID.Hex(),
		DepositoNombre:      dep.Nombre,
		ProductoID:          p.ID.Hex(),
		ProductoNombre:      p.Nombre,
		CategoriaID:         p.Categoria.Hex(),
		CategoriaNombre:     categoriaNombre,
		UnidadMedida:        string(p.UnidadMedida),
		Stock:               stock,
		StockMinimo:         stockMinimoEfectivo,
		StockMinimoOverride: override,
		CostoUnitario:       p.CostoUnitario,
		Valorizado:          float64(stock) * p.CostoUnitario,
		BajoMinimo:          stock < stockMinimoEfectivo,
	}
}

// ObtenerStockProducto devuelve el stock de UN producto en UN depósito.
func (s *Service) ObtenerStockProducto(ctx context.Context, usuarioID, rol, depositoID, productoID string) (StockDepositoDTO, error) {
	if err := s.verificarAcceso(ctx, usuarioID, rol, depositoID); err != nil {
		return StockDepositoDTO{}, err
	}

	dep, err := s.depositoRepositorio.EncontrarPorId(ctx, depositoID)
	if err != nil {
		return StockDepositoDTO{}, err
	}

	p, err := s.productoRepositorio.EncontrarPorId(ctx, productoID)
	if err != nil {
		return StockDepositoDTO{}, ErrProductoInexistente
	}

	cat, err := s.categoriaRepositorio.EncontrarPorId(ctx, p.Categoria.Hex())
	if err != nil {
		return StockDepositoDTO{}, err
	}

	oidDeposito, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return StockDepositoDTO{}, err
	}
	oidProducto, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return StockDepositoDTO{}, err
	}

	dp, err := s.repositorio.EncontrarPorDepositoYProducto(ctx, oidDeposito, oidProducto)
	existe := true
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			existe = false
		} else {
			return StockDepositoDTO{}, err
		}
	}

	return mapearAStockDTO(dp, existe, dep, p, cat.Nombre), nil
}

// ListarStockDeposito lista el stock de TODOS los productos activos en un
// depósito (incluyendo los que todavía tienen stock 0 / sin movimientos),
// con filtro opcional por categoría y por "solo los que están bajo mínimo",
// y paginación (ver Clase 8).
func (s *Service) ListarStockDeposito(ctx context.Context, usuarioID, rol, depositoID string, filtro ListarStockFiltro) (StockDepositoPaginadoDTO, error) {
	if err := s.verificarAcceso(ctx, usuarioID, rol, depositoID); err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	dep, err := s.depositoRepositorio.EncontrarPorId(ctx, depositoID)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	oidDeposito, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	productos, err := s.productoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	filas, err := s.repositorio.EncontrarPorDeposito(ctx, oidDeposito)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	filasPorProducto := make(map[string]DepositoProducto, len(filas))
	for _, f := range filas {
		filasPorProducto[f.ProductoID.Hex()] = f
	}

	categoriasCache := make(map[string]string)

	items := []StockDepositoDTO{}
	for _, p := range productos {
		if filtro.CategoriaID != "" && p.Categoria.Hex() != filtro.CategoriaID {
			continue
		}

		categoriaNombre, ok := categoriasCache[p.Categoria.Hex()]
		if !ok {
			cat, err := s.categoriaRepositorio.EncontrarPorId(ctx, p.Categoria.Hex())
			if err == nil {
				categoriaNombre = cat.Nombre
			}
			categoriasCache[p.Categoria.Hex()] = categoriaNombre
		}

		dp, existe := filasPorProducto[p.ID.Hex()]
		dto := mapearAStockDTO(dp, existe, dep, p, categoriaNombre)

		if filtro.SoloBajoMinimo && !dto.BajoMinimo {
			continue
		}

		items = append(items, dto)
	}

	total := int64(len(items))

	pagina := filtro.Pagina
	if pagina < 1 {
		pagina = 1
	}
	tamañoPagina := filtro.TamañoPagina
	if tamañoPagina < 1 {
		tamañoPagina = 20
	}

	inicio := (pagina - 1) * tamañoPagina
	if inicio > len(items) {
		inicio = len(items)
	}
	fin := inicio + tamañoPagina
	if fin > len(items) {
		fin = len(items)
	}

	return StockDepositoPaginadoDTO{
		Items:        items[inicio:fin],
		Total:        total,
		Pagina:       pagina,
		TamañoPagina: tamañoPagina,
	}, nil
}

// ConfigurarStockMinimo fija el override particular del depósito para un
// producto (reemplaza al stock mínimo general del producto — ver enunciado,
// sección "Productos y categorías").
func (s *Service) ConfigurarStockMinimo(ctx context.Context, usuarioID, rol, depositoID, productoID string, dto ConfigurarStockMinimoDTO) (StockDepositoDTO, error) {
	return s.configurarOQuitar(ctx, usuarioID, rol, depositoID, productoID, &dto.StockMinimo)
}

// QuitarStockMinimo elimina el override: el depósito vuelve a regirse por el
// stock mínimo general del producto.
func (s *Service) QuitarStockMinimo(ctx context.Context, usuarioID, rol, depositoID, productoID string) (StockDepositoDTO, error) {
	return s.configurarOQuitar(ctx, usuarioID, rol, depositoID, productoID, nil)
}

func (s *Service) configurarOQuitar(ctx context.Context, usuarioID, rol, depositoID, productoID string, stockMinimo *int32) (StockDepositoDTO, error) {
	if err := s.verificarAcceso(ctx, usuarioID, rol, depositoID); err != nil {
		return StockDepositoDTO{}, err
	}

	dep, err := s.depositoRepositorio.EncontrarPorId(ctx, depositoID)
	if err != nil {
		return StockDepositoDTO{}, err
	}

	p, err := s.productoRepositorio.EncontrarPorId(ctx, productoID)
	if err != nil {
		return StockDepositoDTO{}, ErrProductoInexistente
	}

	cat, err := s.categoriaRepositorio.EncontrarPorId(ctx, p.Categoria.Hex())
	if err != nil {
		return StockDepositoDTO{}, err
	}

	oidDeposito, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return StockDepositoDTO{}, err
	}
	oidProducto, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return StockDepositoDTO{}, err
	}
	oidUsuario, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return StockDepositoDTO{}, err
	}

	dp, err := s.repositorio.ConfigurarStockMinimo(ctx, oidDeposito, oidProducto, stockMinimo, oidUsuario)
	if err != nil {
		return StockDepositoDTO{}, err
	}

	return mapearAStockDTO(dp, true, dep, p, cat.Nombre), nil
}
