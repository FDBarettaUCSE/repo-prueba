package categoria

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CategoriaService struct {
	categoriaRepositorio CategoriaRepositorio
}

func NuevaCategoriaService(repoCategoria CategoriaRepositorio) *CategoriaService {
	return &CategoriaService{categoriaRepositorio: repoCategoria}
}

func (s *CategoriaService) ListarTodasCategorias(ctx context.Context) ([]CategoriaDTO, error) {

	categorias, err := s.categoriaRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return []CategoriaDTO{}, err
	}

	categoriasDTO := []CategoriaDTO{}

	for _, categoria := range categorias {
		categoriasDTO = append(categoriasDTO, MapeoACategoriaDTO(categoria))
	}

	return categoriasDTO, nil
}

func (s *CategoriaService) BuscarCategoriaPorID(ctx context.Context, id string) (CategoriaDTO, error) {
	categoria, err := s.categoriaRepositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return CategoriaDTO{}, err
	}

	return MapeoACategoriaDTO(categoria), nil
}

func (s *CategoriaService) CrearCategoria(ctx context.Context, idAdmin string, c CrearCategoriaDTO) (CategoriaDTO, error) {

	categoria, err := MapeoCrearDTOACategoria(c)
	if err != nil {
		return CategoriaDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return CategoriaDTO{}, err
	}

	categoria.Auditoria.CreadoPor = oidAdmin
	categoria.Auditoria.CreadoEn = time.Now()

	categoria, err = s.categoriaRepositorio.CrearCategoria(ctx, categoria)
	if err != nil {
		return CategoriaDTO{}, err
	}

	return MapeoACategoriaDTO(categoria), nil
}

func (s *CategoriaService) ActualizarCategoria(ctx context.Context, id, idAdmin string, c CategoriaDTO) (CategoriaDTO, error) {
	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return CategoriaDTO{}, err
	}

	categoria, err := MapeoACategoria(c)
	if err != nil {
		return CategoriaDTO{}, err
	}

	categoria.Auditoria.ModificadoPor = &oidAdmin

	ahora := time.Now()
	categoria.Auditoria.ActualizadoEn = &ahora

	categoria, err = s.categoriaRepositorio.ActualizarCategoria(ctx, id, categoria)
	if err != nil {
		return CategoriaDTO{}, err
	}

	return MapeoACategoriaDTO(categoria), nil
}

func (s *CategoriaService) EliminarCategoria(ctx context.Context, id, idAdmin string) error {
	return s.categoriaRepositorio.EliminarCategoria(ctx, id, idAdmin)
}
