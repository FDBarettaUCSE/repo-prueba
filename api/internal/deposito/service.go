package deposito

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrResponsableInvalido  = errors.New("el responsable del depósito debe ser un gerente")
	ErrDepositoNoEncontrado = errors.New("depósito no encontrado")
)

type DepositoService struct {
	depositoRepositorio    DepositoRepositorio
	responsableRepositorio ResponsableRepositorio
}

func NuevoDepositoService(
	repoDeposito DepositoRepositorio,
	repoResponsable ResponsableRepositorio,
) *DepositoService {
	return &DepositoService{
		depositoRepositorio:    repoDeposito,
		responsableRepositorio: repoResponsable,
	}
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

func (s *DepositoService) CrearDeposito(
	ctx context.Context,
	idAdmin string,
	d CrearDepositoDTO,
) (DepositoDTO, error) {

	err := s.validarResponsable(ctx, d.ResponsableID)
	if err != nil {
		return DepositoDTO{}, err
	}

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

func (s *DepositoService) ActualizarDeposito(
	ctx context.Context,
	id string,
	idAdmin string,
	d ActualizarDepositoDTO,
) (DepositoDTO, error) {

	err := s.validarResponsable(ctx, d.ResponsableID)
	if err != nil {
		return DepositoDTO{}, err
	}

	deposito, err := MapeoActualizarDTOADeposito(d)
	if err != nil {
		return DepositoDTO{}, err
	}

	oidAdmin, err := bson.ObjectIDFromHex(idAdmin)
	if err != nil {
		return DepositoDTO{}, err
	}

	deposito.Auditoria.ModificadoPor = &oidAdmin

	ahora := time.Now()
	deposito.Auditoria.ActualizadoEn = &ahora

	deposito, err = s.depositoRepositorio.ActualizarDeposito(
		ctx,
		id,
		deposito,
	)
	if err != nil {
		return DepositoDTO{}, err
	}

	return MapeoADepositoDTO(deposito), nil
}

func (s *DepositoService) EliminarDeposito(ctx context.Context, id, idAdmin string) error {
	return s.depositoRepositorio.EliminarDeposito(ctx, id, idAdmin)
}

func (s *DepositoService) validarResponsable(
	ctx context.Context,
	responsableID string,
) error {
	if responsableID == "" {
		return nil
	}

	rol, err := s.responsableRepositorio.EncontrarRolPorId(ctx, responsableID)
	if err != nil {
		return err
	}

	if rol != "gerente" {
		return ErrResponsableInvalido
	}

	return nil
}
