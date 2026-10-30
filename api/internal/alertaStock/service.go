package alertastock

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxTamanioPagina = 100

var (
	ErrAccesoDenegado = errors.New("no tenés permiso para operar sobre este depósito")
)

// Service resuelve la consulta y resolución de alertas de stock.
type Service struct {
	repositorio AlertaStockRepositorio
}

func NuevoService(
	repo AlertaStockRepositorio,
) *Service {
	return &Service{
		repositorio: repo,
	}
}

// resolverAlcanceDeposito decide el depósito efectivo sobre el que va a operar
// cada consulta: Administrador y Auditor pueden pedir cualquiera (o "" = toda
// la empresa, sin filtrar); Gerente y Operario quedan siempre forzados al
// suyo. Si piden explícitamente otro depósito, se deniega el acceso.
func resolverAlcanceDeposito(rol, depositoUsuario, depositoIDPedido string) (string, error) {
	if rol == "administrador" || rol == "auditor" {
		return depositoIDPedido, nil
	}
	if depositoIDPedido != "" && depositoIDPedido != depositoUsuario {
		return "", ErrAccesoDenegado
	}
	return depositoUsuario, nil
}

// ListarAlertas devuelve las alertas de stock paginadas, con filtro opcional
// por depósito y por estado de resolución. El campo Pendientes del resultado
// es independiente de la página pedida.
func (s *Service) ListarAlertas(ctx context.Context, rol, depUsuario string, filtro FiltroAlertasDTO) (AlertasPaginadasDTO, error) {
	depositoIDStr, err := resolverAlcanceDeposito(rol, depUsuario, filtro.DepositoID)
	if err != nil {
		return AlertasPaginadasDTO{}, err
	}

	var depOID *bson.ObjectID
	if depositoIDStr != "" {
		oid, err := bson.ObjectIDFromHex(depositoIDStr)
		if err != nil {
			return AlertasPaginadasDTO{}, err
		}
		depOID = &oid
	}

	if filtro.Pagina < 1 {
		filtro.Pagina = 1
	}
	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 20
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	soloPendientes := filtro.SoloPendientes == "true"

	alertas, total, err := s.repositorio.Listar(
		ctx,
		depOID,
		soloPendientes,
		filtro.Pagina,
		filtro.TamanioPagina,
	)
	if err != nil {
		return AlertasPaginadasDTO{}, err
	}

	pendientes, err := s.repositorio.ContarPendientes(ctx, depOID)
	if err != nil {
		return AlertasPaginadasDTO{}, err
	}

	items := make([]AlertaStockDTO, 0, len(alertas))

	for _, alerta := range alertas {
		items = append(items, ConvertirADTO(alerta))
	}

	return AlertasPaginadasDTO{
		Items:         items,
		Total:         total,
		Pendientes:    pendientes,
		Pagina:        filtro.Pagina,
		TamanioPagina: filtro.TamanioPagina,
	}, nil
}
