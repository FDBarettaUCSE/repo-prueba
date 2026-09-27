package proveedor

import (
	"api/internal/producto"
	"api/internal/productoProveedor"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrCUITInvalido        = errors.New("el CUIT ingresado no es válido")
	ErrCUITYaRegistrado    = errors.New("ya existe un proveedor registrado con ese CUIT")
	ErrProductoInexistente = errors.New("el producto ingresado no existe")
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

func (s *ProveedorService) ListarTodosProveedores(ctx context.Context) ([]ProveedorDTO, error) {
	proveedores, err := s.proveedorRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return nil, err
	}

	proveedoresDTO := []ProveedorDTO{}
	for _, p := range proveedores {
		proveedoresDTO = append(proveedoresDTO, MapeoAProveedorDTO(p))
	}
	return proveedoresDTO, nil
}

func (s *ProveedorService) BuscarProveedorPorID(ctx context.Context, id string) (ProveedorDTO, error) {
	proveedor, err := s.proveedorRepositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return ProveedorDTO{}, err
	}
	return MapeoAProveedorDTO(proveedor), nil
}

func (s *ProveedorService) CrearProveedor(ctx context.Context, idAdmin string, d CrearProveedorDTO) (ProveedorDTO, error) {
	if !EsCUITValido(d.CUIT) {
		return ProveedorDTO{}, ErrCUITInvalido
	}

	if _, err := s.proveedorRepositorio.EncontrarPorCUIT(ctx, d.CUIT); err == nil {
		return ProveedorDTO{}, ErrCUITYaRegistrado
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return ProveedorDTO{}, err
	}

	proveedor := MapeoCrearDTOAProveedor(d)
	proveedor.Auditoria.CreadoPor = oidAdmin
	proveedor.Auditoria.CreadoEn = time.Now()

	proveedor, err = s.proveedorRepositorio.CrearProveedor(ctx, proveedor)
	if err != nil {
		return ProveedorDTO{}, err
	}
	return MapeoAProveedorDTO(proveedor), nil
}

func (s *ProveedorService) ActualizarProveedor(ctx context.Context, id, idAdmin string, d ActualizarProveedorDTO) (ProveedorDTO, error) {
	if !EsCUITValido(d.CUIT) {
		return ProveedorDTO{}, ErrCUITInvalido
	}

	// Si el CUIT cambió, hay que confirmar que no choque con el de otro proveedor.
	if existente, err := s.proveedorRepositorio.EncontrarPorCUIT(ctx, d.CUIT); err == nil && existente.ID.Hex() != id {
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

	proveedor, err = s.proveedorRepositorio.ActualizarProveedor(ctx, id, proveedor)
	if err != nil {
		return ProveedorDTO{}, err
	}
	return MapeoAProveedorDTO(proveedor), nil
}

func (s *ProveedorService) EliminarProveedor(ctx context.Context, id, idAdmin string) error {
	if err := s.proveedorRepositorio.EliminarProveedor(ctx, id, idAdmin); err != nil {
		return err
	}

	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	// No debe quedar un producto "suministrado" por un proveedor dado de baja.
	return s.productoProveedorRepositorio.EliminarTodosDeProveedor(ctx, oid)
}

// --- Productos que suministra ---

func (s *ProveedorService) ListarProductosSuministrados(ctx context.Context, proveedorID string) ([]ProductoSuministradoDTO, error) {
	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return nil, err
	}

	vinculos, err := s.productoProveedorRepositorio.EncontrarPorProveedor(ctx, oidProveedor)
	if err != nil {
		return nil, err
	}

	resultado := []ProductoSuministradoDTO{}
	for _, v := range vinculos {
		prod, err := s.productoRepositorio.EncontrarPorId(ctx, v.ProductoID.Hex())
		if err != nil {
			// Si el producto fue eliminado mientras tanto, no rompemos el
			// listado entero: lo salteamos.
			continue
		}
		resultado = append(resultado, ProductoSuministradoDTO{
			ProductoID:        v.ProductoID.Hex(),
			ProductoNombre:    prod.Nombre,
			Costo:             v.CostoUnidad,
			TiempoEntregaDias: v.TiempoEstimadoDias,
		})
	}
	return resultado, nil
}

// ListarProveedoresDeProducto es la vista inversa: dado un producto, quiénes
// lo suministran y en qué condiciones. La va a usar el módulo de Órdenes de
// Compra para elegir un proveedor (por ejemplo, el de menor costo) al generar
// la sugerencia automática por stock bajo mínimo.
func (s *ProveedorService) ListarProveedoresDeProducto(ctx context.Context, productoID string) ([]ProveedorDeProductoDTO, error) {
	oidProducto, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return nil, err
	}

	vinculos, err := s.productoProveedorRepositorio.EncontrarPorProducto(ctx, oidProducto)
	if err != nil {
		return nil, err
	}

	resultado := []ProveedorDeProductoDTO{}
	for _, v := range vinculos {
		prov, err := s.proveedorRepositorio.EncontrarPorId(ctx, v.ProveedorID.Hex())
		if err != nil {
			continue
		}
		resultado = append(resultado, ProveedorDeProductoDTO{
			ProveedorID:       prov.ID.Hex(),
			ProveedorNombre:   prov.Nombre,
			Costo:             v.CostoUnidad,
			TiempoEntregaDias: v.TiempoEstimadoDias,
		})
	}
	return resultado, nil
}

// AgregarProductoSuministrado también sirve para actualizar costo/tiempo si
// el vínculo ya existía (ver ProductoProveedorRepositorio.Guardar: es un upsert).
func (s *ProveedorService) AgregarProductoSuministrado(ctx context.Context, proveedorID string, d AgregarProductoSuministradoDTO) error {
	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return err
	}

	// Confirmamos que el proveedor y el producto existan antes de vincularlos.
	if _, err := s.proveedorRepositorio.EncontrarPorId(ctx, proveedorID); err != nil {
		return err
	}
	prod, err := s.productoRepositorio.EncontrarPorId(ctx, d.ProductoID)
	if err != nil {
		return ErrProductoInexistente
	}

	if d.Costo <= 0 {
		return errors.New("el costo debe ser mayor a cero")
	}
	if d.TiempoEntregaDias < 0 {
		return errors.New("el tiempo de entrega no puede ser negativo")
	}

	return s.productoProveedorRepositorio.Guardar(ctx, productoproveedor.ProductoProveedor{
		ProductoID:         prod.ID,
		ProveedorID:        oidProveedor,
		CostoUnidad:        d.Costo,
		TiempoEstimadoDias: d.TiempoEntregaDias,
	})
}

func (s *ProveedorService) ActualizarProductoSuministrado(ctx context.Context, proveedorID, productoID string, d ActualizarProductoSuministradoDTO) error {
	return s.AgregarProductoSuministrado(ctx, proveedorID, AgregarProductoSuministradoDTO{
		ProductoID:        productoID,
		Costo:             d.Costo,
		TiempoEntregaDias: d.TiempoEntregaDias,
	})
}

func (s *ProveedorService) EliminarProductoSuministrado(ctx context.Context, proveedorID, productoID string) error {
	oidProveedor, err := bson.ObjectIDFromHex(proveedorID)
	if err != nil {
		return err
	}
	oidProducto, err := bson.ObjectIDFromHex(productoID)
	if err != nil {
		return err
	}
	return s.productoProveedorRepositorio.Eliminar(ctx, oidProducto, oidProveedor)
}
