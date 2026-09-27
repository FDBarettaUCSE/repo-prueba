package deposito

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type DepositoService struct {
	depositoRepositorio DepositoRepositorio
}

func NuevoDepositoService(repoDeposito DepositoRepositorio) *DepositoService {
	return &DepositoService{depositoRepositorio: repoDeposito}
}

// Falta agregarle sus respectivos inventarios.
func (s *DepositoService) ListarTodosDepositos(ctx context.Context) ([]DepositoDTO, error) {

	depositos, err := s.depositoRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return []DepositoDTO{}, err
	}

	depositosDTO := []DepositoDTO{}

	for _, deposito := range depositos {
		depositosDTO = append(depositosDTO, MapeoADepositoDTO(deposito))
	}

	return depositosDTO, nil
}

func (s *DepositoService) BuscarDepositoPorID(ctx context.Context, id string) (DepositoDTO, error) {
	deposito, err := s.depositoRepositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return DepositoDTO{}, err
	}

	return MapeoADepositoDTO(deposito), nil
}

func (s *DepositoService) CrearDeposito(ctx context.Context, idAdmin string, d CrearDepositoDTO) (DepositoDTO, error) {

	deposito, err := MapeoCrearDTOADeposito(d)
	if err != nil {
		return DepositoDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return DepositoDTO{}, err
	}

	deposito.Auditoria.CreadoPor = oidAdmin
	deposito.Auditoria.CreadoEn = time.Now()

	deposito, err = s.depositoRepositorio.CrearDeposito(ctx, deposito)
	if err != nil {
		return DepositoDTO{}, err
	}

	return MapeoADepositoDTO(deposito), nil
}

func (s *DepositoService) ActualizarDeposito(ctx context.Context, id, idAdmin string, d DepositoDTO) (DepositoDTO, error) {
	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return DepositoDTO{}, err
	}

	deposito, err := MapeoADeposito(d)
	if err != nil {
		return DepositoDTO{}, err
	}

	deposito.Auditoria.ModificadoPor = &oidAdmin

	ahora := time.Now()
	deposito.Auditoria.ActualizadoEn = &ahora

	deposito, err = s.depositoRepositorio.ActualizarDeposito(ctx, id, deposito)
	if err != nil {
		return DepositoDTO{}, err
	}

	return MapeoADepositoDTO(deposito), nil
}

func (s *DepositoService) EliminarDeposito(ctx context.Context, id, idAdmin string) error {
	return s.depositoRepositorio.EliminarDeposito(ctx, id, idAdmin)
}
