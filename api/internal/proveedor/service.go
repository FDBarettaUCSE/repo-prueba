package proveedor

import (
	"api/internal/producto"
	productoproveedor "api/internal/productoProveedor"
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrNombreProveedorVacio = errors.New(
		"el nombre del proveedor es obligatorio",
	)

	ErrEmailProveedorInvalido = errors.New(
		"el email del proveedor es obligatorio",
	)

	ErrCUITInvalido = errors.New(
		"el CUIT ingresado no es válido",
	)

	ErrCUITYaRegistrado = errors.New(
		"ya existe un proveedor registrado con ese CUIT",
	)

	ErrProveedorInexistente = errors.New(
		"el proveedor solicitado no existe",
	)

	ErrProductoInexistente = errors.New(
		"el producto ingresado no existe",
	)

	ErrCostoInvalido = errors.New(
		"el costo debe ser mayor a cero",
	)

	ErrTiempoEntregaInvalido = errors.New(
		"el tiempo de entrega no puede ser negativo",
	)

	ErrProductoSuministradoInexistente = productoproveedor.ErrProductoSuministradoInexistente
	ErrProductoYaSuministrado          = productoproveedor.ErrProductoYaSuministrado
)

type ProveedorService struct {
	proveedorRepositorio         ProveedorRepositorio
	productoProveedorRepositorio productoproveedor.ProductoProveedorRepositorio
	productoRepositorio          producto.ProductoRepositorio
}

func NuevoProveedorService(
	repoProveedor ProveedorRepositorio,
	repoProductoProveedor productoproveedor.ProductoProveedorRepositorio,
	repoProducto producto.ProductoRepositorio,
) *ProveedorService {
	return &ProveedorService{
		proveedorRepositorio:         repoProveedor,
		productoProveedorRepositorio: repoProductoProveedor,
		productoRepositorio:          repoProducto,
	}
}

func (s *ProveedorService) ListarTodosProveedores(
	ctx context.Context,
) ([]ProveedorDTO, error) {

	proveedores, err := s.proveedorRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return nil, err
	}

	proveedoresDTO := make([]ProveedorDTO, 0, len(proveedores))

	for _, p := range proveedores {
		proveedoresDTO = append(
			proveedoresDTO,
			MapeoAProveedorDTO(p),
		)
	}

	return proveedoresDTO, nil
}

func (s *ProveedorService) BuscarProveedorPorID(
	ctx context.Context,
	id string,
) (ProveedorDTO, error) {

	_, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ProveedorDTO{}, ErrProveedorInexistente
	}

	proveedor, err := s.proveedorRepositorio.EncontrarPorId(
		ctx,
		id,
	)
	if err != nil {
		return ProveedorDTO{}, ErrProveedorInexistente
	}

	return MapeoAProveedorDTO(proveedor), nil
}

func (s *ProveedorService) CrearProveedor(
	ctx context.Context,
	idAdmin string,
	d CrearProveedorDTO,
) (ProveedorDTO, error) {

	if strings.TrimSpace(d.Nombre) == "" {
		return ProveedorDTO{}, ErrNombreProveedorVacio
	}

	if strings.TrimSpace(d.Email) == "" {
		return ProveedorDTO{}, ErrEmailProveedorInvalido
	}

	if !EsCUITValido(d.CUIT) {
		return ProveedorDTO{}, ErrCUITInvalido
	}

	if _, err := s.proveedorRepositorio.EncontrarPorCUIT(
		ctx,
		d.CUIT,
	); err == nil {
		return ProveedorDTO{}, ErrCUITYaRegistrado
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProveedorDTO{}, err
	}

	proveedor := MapeoCrearDTOAProveedor(d)

	proveedor.Auditoria.CreadoPor = oidAdmin
	proveedor.Auditoria.CreadoEn = time.Now()

	proveedor, err = s.proveedorRepositorio.CrearProveedor(
		ctx,
		proveedor,
	)
	if err != nil {
		return ProveedorDTO{}, err
	}

	return MapeoAProveedorDTO(proveedor), nil
}

func (s *ProveedorService) ActualizarProveedor(
	ctx context.Context,
	id string,
	idAdmin string,
	d ActualizarProveedorDTO,
) (ProveedorDTO, error) {

	if strings.TrimSpace(d.Nombre) == "" {
		return ProveedorDTO{}, ErrNombreProveedorVacio
	}

	if strings.TrimSpace(d.Email) == "" {
		return ProveedorDTO{}, ErrEmailProveedorInvalido
	}

	if !EsCUITValido(d.CUIT) {
		return ProveedorDTO{}, ErrCUITInvalido
	}

	// Primero verificamos que el proveedor exista.
	if _, err := s.proveedorRepositorio.EncontrarPorId(
		ctx,
		id,
	); err != nil {
		return ProveedorDTO{}, ErrProveedorInexistente
	}

	// Si el CUIT cambió, no puede coincidir con otro proveedor.
	if existente, err := s.proveedorRepositorio.EncontrarPorCUIT(
		ctx,
		d.CUIT,
	); err == nil && existente.ID.Hex() != id {
		return ProveedorDTO{}, ErrCUITYaRegistrado
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProveedorDTO{}, err
	}

	proveedor := MapeoActualizarDTOAProveedor(d)

	proveedor.Auditoria.ModificadoPor = &oidAdmin

	ahora := time.Now()
	proveedor.Auditoria.ActualizadoEn = &ahora

	proveedor, err = s.proveedorRepositorio.ActualizarProveedor(
		ctx,
		id,
		proveedor,
	)
	if err != nil {
		return ProveedorDTO{}, err
	}

	return MapeoAProveedorDTO(proveedor), nil
}

func (s *ProveedorService) EliminarProveedor(
	ctx context.Context,
	id string,
	idAdmin string,
) error {

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return ErrProveedorInexistente
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	// Verificamos que el proveedor exista.
	if _, err := s.proveedorRepositorio.EncontrarPorId(
		ctx,
		id,
	); err != nil {
		return ErrProveedorInexistente
	}

	// Eliminamos lógicamente el proveedor y auditamos la operación.
	if err := s.proveedorRepositorio.EliminarProveedor(
		ctx,
		id,
		idAdmin,
	); err != nil {
		return err
	}

	// Eliminamos lógicamente sus relaciones y auditamos la operación.
	return s.productoProveedorRepositorio.EliminarTodosDeProveedor(
		ctx,
		oid,
		oidAdmin,
	)
}

// --- Productos que suministra ---

func (s *ProveedorService) ListarProductosSuministrados(
	ctx context.Context,
	proveedorID string,
) ([]ProductoSuministradoDTO, error) {

	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return nil, ErrProveedorInexistente
	}

	if _, err := s.proveedorRepositorio.EncontrarPorId(
		ctx,
		proveedorID,
	); err != nil {
		return nil, ErrProveedorInexistente
	}

	vinculos, err := s.productoProveedorRepositorio.EncontrarPorProveedor(
		ctx,
		oidProveedor,
	)

	if err != nil {
		return nil, err
	}

	if len(vinculos) == 0 {
		return []ProductoSuministradoDTO{}, nil
	}

	productoIDs := make(
		[]bson.ObjectID,
		0,
		len(vinculos),
	)

	for _, vinculo := range vinculos {
		productoIDs = append(
			productoIDs,
			vinculo.ProductoID,
		)
	}

	productos, err := s.productoRepositorio.EncontrarPorIds(
		ctx,
		productoIDs,
	)
	if err != nil {
		return nil, err
	}

	productosPorID := make(
		map[bson.ObjectID]producto.Producto,
		len(productos),
	)

	for _, p := range productos {
		productosPorID[p.ID] = p
	}

	resultado := make(
		[]ProductoSuministradoDTO,
		0,
		len(vinculos),
	)

	for _, vinculo := range vinculos {

		p, existe := productosPorID[vinculo.ProductoID]
		if !existe {
			continue
		}

		resultado = append(
			resultado,
			ProductoSuministradoDTO{
				ProductoID:        p.ID.Hex(),
				ProductoNombre:    p.Nombre,
				Costo:             vinculo.CostoUnidad,
				TiempoEntregaDias: vinculo.TiempoEstimadoDias,
			},
		)
	}

	return resultado, nil
}

func (s *ProveedorService) ListarProveedoresDeProducto(
	ctx context.Context,
	productoID string,
) ([]ProveedorDeProductoDTO, error) {

	oidProducto, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return nil, ErrProductoInexistente
	}

	// El producto debe existir para poder consultar sus proveedores.
	if _, err := s.productoRepositorio.EncontrarPorId(
		ctx,
		productoID,
	); err != nil {
		return nil, ErrProductoInexistente
	}

	vinculos, err := s.productoProveedorRepositorio.EncontrarPorProducto(
		ctx,
		oidProducto,
	)
	if err != nil {
		return nil, err
	}

	if len(vinculos) == 0 {
		return []ProveedorDeProductoDTO{}, nil
	}

	proveedorIDs := make(
		[]bson.ObjectID,
		0,
		len(vinculos),
	)

	for _, v := range vinculos {
		proveedorIDs = append(
			proveedorIDs,
			v.ProveedorID,
		)
	}

	proveedores, err := s.proveedorRepositorio.EncontrarPorIds(
		ctx,
		proveedorIDs,
	)
	if err != nil {
		return nil, err
	}

	proveedoresPorID := make(
		map[bson.ObjectID]Proveedor,
		len(proveedores),
	)

	for _, proveedor := range proveedores {
		proveedoresPorID[proveedor.ID] = proveedor
	}

	resultado := make(
		[]ProveedorDeProductoDTO,
		0,
		len(vinculos),
	)

	for _, v := range vinculos {

		prov, existe := proveedoresPorID[v.ProveedorID]
		if !existe {
			continue
		}

		resultado = append(
			resultado,
			ProveedorDeProductoDTO{
				ProveedorID:       prov.ID.Hex(),
				ProveedorNombre:   prov.Nombre,
				Costo:             v.CostoUnidad,
				TiempoEntregaDias: v.TiempoEstimadoDias,
			},
		)
	}

	return resultado, nil
}

func (s *ProveedorService) AgregarProductoSuministrado(
	ctx context.Context,
	idAdmin string,
	proveedorID string,
	d AgregarProductoSuministradoDTO,
) error {
	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return ErrProveedorInexistente
	}

	_, err = s.proveedorRepositorio.EncontrarPorId(ctx, proveedorID)
	if err != nil {
		return ErrProveedorInexistente
	}

	prod, err := s.productoRepositorio.EncontrarPorId(ctx, d.ProductoID)
	if err != nil {
		return ErrProductoInexistente
	}

	if d.Costo <= 0 {
		return ErrCostoInvalido
	}

	if d.TiempoEntregaDias < 0 {
		return ErrTiempoEntregaInvalido
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	productoProveedor := MapeoAgregarProductoSuministradoDTOAProductoProveedor(
		d,
		prod.ID,
		oidProveedor,
	)

	productoProveedor.Auditoria.CreadoPor = oidAdmin
	productoProveedor.Auditoria.CreadoEn = time.Now()

	return s.productoProveedorRepositorio.Crear(ctx, productoProveedor)
}

func (s *ProveedorService) ActualizarProductoSuministrado(
	ctx context.Context,
	idAdmin string,
	proveedorID string,
	productoID string,
	d ActualizarProductoSuministradoDTO,
) error {
	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return ErrProveedorInexistente
	}

	_, err = s.proveedorRepositorio.EncontrarPorId(ctx, proveedorID)
	if err != nil {
		return ErrProveedorInexistente
	}

	prod, err := s.productoRepositorio.EncontrarPorId(ctx, productoID)
	if err != nil {
		return ErrProductoInexistente
	}

	if d.Costo <= 0 {
		return ErrCostoInvalido
	}

	if d.TiempoEntregaDias < 0 {
		return ErrTiempoEntregaInvalido
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return err
	}

	productoProveedor := MapeoActualizarProductoSuministradoDTOAProductoProveedor(
		d,
		prod.ID,
		oidProveedor,
	)

	ahora := time.Now()
	productoProveedor.Auditoria.ModificadoPor = &oidAdmin
	productoProveedor.Auditoria.ActualizadoEn = &ahora

	return s.productoProveedorRepositorio.Actualizar(
		ctx,
		productoProveedor,
	)
}

func (s *ProveedorService) EliminarProductoSuministrado(
	ctx context.Context,
	adminID string,
	proveedorID string,
	productoID string,
) error {

	oidAdmin, err := bson.ObjectIDFromHex(adminID)
	if err != nil {
		return err
	}

	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return ErrProveedorInexistente
	}

	if _, err := s.proveedorRepositorio.EncontrarPorId(
		ctx,
		proveedorID,
	); err != nil {
		return ErrProveedorInexistente
	}

	oidProducto, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return ErrProductoInexistente
	}

	if _, err := s.productoRepositorio.EncontrarPorId(
		ctx,
		productoID,
	); err != nil {
		return ErrProductoInexistente
	}

	return s.productoProveedorRepositorio.Eliminar(
		ctx,
		oidProducto,
		oidProveedor,
		oidAdmin,
	)
}
