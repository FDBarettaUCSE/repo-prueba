package depositoproducto

import (
	"api/internal/categoria"
	"api/internal/deposito"
	"api/internal/producto"
	"api/internal/usuario"
	"context"
	"errors"

	evaluacionstock "api/internal/evaluacionStock"

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
	evaluadorStock       evaluacionstock.EvaluadorStock
}

func NuevoService(
	repo DepositoProductoRepositorio,
	repoDeposito deposito.DepositoRepositorio,
	repoProducto producto.ProductoRepositorio,
	repoCategoria categoria.CategoriaRepositorio,
	repoUsuario usuario.UsuarioRepositorio,
	evaluadorStock evaluacionstock.EvaluadorStock,
) *Service {
	return &Service{
		repositorio:          repo,
		depositoRepositorio:  repoDeposito,
		productoRepositorio:  repoProducto,
		categoriaRepositorio: repoCategoria,
		usuarioRepositorio:   repoUsuario,
		evaluadorStock:       evaluadorStock,
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

	if u.DepositoAsociado.IsZero() || u.DepositoAsociado.Hex() != depositoID {
		return ErrAccesoDenegado
	}

	return nil
}

// ObtenerStockProducto devuelve el stock de UN producto en UN depósito.
func (s *Service) ObtenerStockProducto(
	ctx context.Context,
	usuarioID, rol, depositoID, productoID string,
) (StockDepositoDTO, error) {
	if err := s.verificarAcceso(ctx, usuarioID, rol, depositoID); err != nil {
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

	dp, err := s.repositorio.EncontrarPorDepositoYProducto(
		ctx,
		oidDeposito,
		oidProducto,
	)

	existe := true
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			existe = false
		} else {
			return StockDepositoDTO{}, err
		}
	}

	// Determinar el stock y el mínimo efectivo.
	var stock int32
	var override *int32

	if existe {
		stock = dp.Stock
		override = dp.StockMinimo
	}

	stockMinimo := p.StockMinimo
	if override != nil {
		stockMinimo = *override
	}

	// Preparar la fila con los mismos datos que devuelve el listado.
	fila := StockDepositoFila{
		DepositoID:          dep.ID,
		DepositoNombre:      dep.Nombre,
		ProductoID:          p.ID,
		ProductoNombre:      p.Nombre,
		CategoriaID:         p.Categoria,
		CategoriaNombre:     cat.Nombre,
		UnidadMedida:        string(p.UnidadMedida),
		Stock:               stock,
		StockMinimo:         stockMinimo,
		StockMinimoOverride: override,
		CostoUnitario:       p.CostoUnitario,
		Valorizado:          float64(stock) * p.CostoUnitario,
		BajoMinimo:          stock < stockMinimo,
	}

	return MapearAStockDepositoDTO(fila, rol), nil
}

// ListarStockDeposito lista el stock de TODOS los productos activos en un
// depósito (incluyendo los que todavía tienen stock 0 / sin movimientos),
// con filtro opcional por categoría y por "solo los que están bajo mínimo",
// y paginación.
func (s *Service) ListarStockDeposito(
	ctx context.Context,
	usuarioID, rol, depositoID string,
	filtro ListarStockFiltro,
) (StockDepositoPaginadoDTO, error) {
	// Verificar que el usuario tenga acceso al depósito.
	if err := s.verificarAcceso(ctx, usuarioID, rol, depositoID); err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	// Obtener los datos del depósito.
	dep, err := s.depositoRepositorio.EncontrarPorId(ctx, depositoID)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	// Convertir el ID del depósito a ObjectID.
	oidDeposito, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	// Normalizar la paginación para que coincida con la consulta.
	pagina := filtro.Pagina
	if pagina < 1 {
		pagina = 1
	}

	tamanioPagina := filtro.TamanioPagina
	if tamanioPagina < 1 {
		tamanioPagina = 20
	}
	if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	// Convertir el filtro al tipo utilizado por el repositorio.
	filtro.Pagina = pagina
	filtro.TamanioPagina = tamanioPagina
	filtroRepo := MapeoAFiltroRepositorio(filtro)

	// Obtener los resultados y el total desde MongoDB.
	filas, total, err := s.repositorio.ListarStockDeposito(
		ctx,
		oidDeposito,
		filtroRepo,
	)
	if err != nil {
		return StockDepositoPaginadoDTO{}, err
	}

	// Convertir las filas obtenidas al DTO del frontend.
	items := make([]StockDepositoDTO, 0, len(filas))

	for _, fila := range filas {
		fila.DepositoID = dep.ID
		fila.DepositoNombre = dep.Nombre
		items = append(items, MapearAStockDepositoDTO(fila, rol))
	}

	return StockDepositoPaginadoDTO{
		Items:         items,
		Total:         total,
		Pagina:        pagina,
		TamanioPagina: tamanioPagina,
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

func (s *Service) configurarOQuitar(
	ctx context.Context,
	usuarioID, rol, depositoID, productoID string,
	stockMinimo *int32,
) (StockDepositoDTO, error) {
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

	dp, err := s.repositorio.ConfigurarStockMinimo(
		ctx,
		oidDeposito,
		oidProducto,
		stockMinimo,
		oidUsuario,
	)
	if err != nil {
		return StockDepositoDTO{}, err
	}

	stockMinimoEfectivo := p.StockMinimo
	if dp.StockMinimo != nil {
		stockMinimoEfectivo = *dp.StockMinimo
	}

	if err := s.evaluadorStock.EvaluarStock(
		ctx,
		oidDeposito,
		oidProducto,
		dp.Stock,
		stockMinimoEfectivo,
		oidUsuario,
	); err != nil {
		return StockDepositoDTO{}, err
	}

	fila := StockDepositoFila{
		DepositoID:          dep.ID,
		DepositoNombre:      dep.Nombre,
		ProductoID:          p.ID,
		ProductoNombre:      p.Nombre,
		CategoriaID:         p.Categoria,
		CategoriaNombre:     cat.Nombre,
		UnidadMedida:        string(p.UnidadMedida),
		Stock:               dp.Stock,
		StockMinimo:         stockMinimoEfectivo,
		StockMinimoOverride: dp.StockMinimo,
		CostoUnitario:       p.CostoUnitario,
		Valorizado:          float64(dp.Stock) * p.CostoUnitario,
		BajoMinimo:          dp.Stock < stockMinimoEfectivo,
	}

	return MapearAStockDepositoDTO(fila, rol), nil
}
