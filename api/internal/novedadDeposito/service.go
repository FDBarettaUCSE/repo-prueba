package novedaddeposito

import (
	"api/internal/deposito"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

const MaxTamanioPagina = 100

var (
	ErrAccesoDenegado      = errors.New("no tenés permiso para operar sobre este depósito")
	ErrSinDestinatarios    = errors.New(`hay que indicar al menos un depósito destino, o "todos"`)
	ErrNovedadNoEncontrada = errors.New("novedad no encontrada")
)

// Service resuelve el envío, la lectura y la respuesta de novedades.
// Depende de deposito.DepositoRepositorio para validar que el depósito
// destino exista y, en la difusión, para resolver "todos" a la lista real de
// depósitos activos.
type Service struct {
	repositorio         NovedadDepositoRepositorio
	depositoRepositorio deposito.DepositoRepositorio
}

func NuevoService(repo NovedadDepositoRepositorio, repoDeposito deposito.DepositoRepositorio) *Service {
	return &Service{
		repositorio:         repo,
		depositoRepositorio: repoDeposito,
	}
}

// verificarAccesoDeposito: Administradores y Auditores pueden operar sobre
// cualquier depósito; Gerente y Operario solo sobre el suyo — que ya viene
// resuelto en el JWT (ver middleware.UsuarioContexto.DepositoID), sin
// necesidad de ir a buscarlo a la base.
func verificarAccesoDeposito(rol, depositoUsuario, depositoID string) error {
	if rol == "auditor" || rol == "administrador" {
		return nil
	}

	if depositoUsuario != depositoID {
		return ErrAccesoDenegado
	}

	return nil
}

// lectorParaRol devuelve el ObjectID del usuario cuando el rol es operario
// (el estado de lectura y el contador de no leídas son individuales por
// operario), o nil para el resto de los roles.
func lectorParaRol(rol, usuarioID string) (*bson.ObjectID, error) {
	if rol != "operario" {
		return nil, nil
	}
	oid, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return nil, err
	}
	return &oid, nil
}

func crearNovedad(
	ctx context.Context,
	repo NovedadDepositoRepositorio,
	texto, depositoID, usuarioID string,
) (NovedadDTO, error) {

	novedad, err := (EnviarNovedadDTO{
		Texto: texto,
	}).ConvertirAModelo(depositoID)

	if err != nil {
		return NovedadDTO{}, err
	}

	oidUsuario, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return NovedadDTO{}, err
	}

	novedad.Auditoria.CreadoPor = oidUsuario
	novedad.Auditoria.CreadoEn = novedad.FechaEnvio

	guardada, err := repo.Crear(ctx, novedad)
	if err != nil {
		return NovedadDTO{}, err
	}

	return guardada.ConvertirADTO(nil), nil
}

// EnviarADeposito registra una novedad para UN depósito puntual.
// El Gerente solo puede usarla sobre SU propio depósito.
// El Administrador y el Auditor pueden apuntar a cualquiera.
func (s *Service) EnviarADeposito(
	ctx context.Context,
	rol,
	depositoUsuario,
	depositoID,
	usuarioID string,
	dto EnviarNovedadDTO,
) (NovedadDTO, error) {

	if rol == "gerente" && depositoUsuario != depositoID {
		return NovedadDTO{}, ErrAccesoDenegado
	}

	if _, err := s.depositoRepositorio.EncontrarPorId(ctx, depositoID); err != nil {
		return NovedadDTO{}, err
	}

	return crearNovedad(
		ctx,
		s.repositorio,
		dto.Texto,
		depositoID,
		usuarioID,
	)
}

// EnviarDifusion manda la misma novedad a varios depósitos elegidos,
// o a todos.
func (s *Service) EnviarDifusion(
	ctx context.Context,
	usuarioID string,
	dto EnviarNovedadDifusionDTO,
) ([]NovedadDTO, error) {

	depositosDestino := dto.DepositosIDs

	if dto.Todos {
		depositos, err := s.depositoRepositorio.EncontrarTodos(ctx)
		if err != nil {
			return nil, err
		}

		depositosDestino = make([]string, 0, len(depositos))

		for _, deposito := range depositos {
			depositosDestino = append(
				depositosDestino,
				deposito.ID.Hex(),
			)
		}
	} else {
		if len(depositosDestino) == 0 {
			return nil, ErrSinDestinatarios
		}

		if _, err := s.depositoRepositorio.EncontrarPorIds(
			ctx,
			depositosDestino,
		); err != nil {
			return nil, err
		}
	}

	if len(depositosDestino) == 0 {
		return nil, ErrSinDestinatarios
	}

	resultado := make([]NovedadDTO, 0, len(depositosDestino))

	for _, depositoID := range depositosDestino {
		enviada, err := crearNovedad(
			ctx,
			s.repositorio,
			dto.Texto,
			depositoID,
			usuarioID,
		)

		if err != nil {
			return nil, err
		}

		resultado = append(resultado, enviada)
	}

	return resultado, nil
}

// ListarDeDeposito lista las novedades de un depósito,
// paginadas y con filtro opcional "solo no leídas".
func (s *Service) ListarDeDeposito(
	ctx context.Context,
	rol,
	depositoUsuario,
	depositoID,
	usuarioID string,
	filtro FiltroNovedadesDTO,
) (NovedadesPaginadoDTO, error) {

	if err := verificarAccesoDeposito(
		rol,
		depositoUsuario,
		depositoID,
	); err != nil {
		return NovedadesPaginadoDTO{}, err
	}

	oidDeposito, err := bson.ObjectIDFromHex(depositoID)
	if err != nil {
		return NovedadesPaginadoDTO{}, err
	}

	lector, err := lectorParaRol(rol, usuarioID)
	if err != nil {
		return NovedadesPaginadoDTO{}, err
	}

	noLeidas, err := s.repositorio.ContarNoLeidas(
		ctx,
		oidDeposito,
		lector,
	)
	if err != nil {
		return NovedadesPaginadoDTO{}, err
	}

	pagina := filtro.Pagina
	if pagina < 1 {
		pagina = 1
	}

	tamanioPagina := filtro.TamanioPagina
	if tamanioPagina < 1 {
		tamanioPagina = 20
	} else if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	novedades, total, err := s.repositorio.ListarPorDeposito(
		ctx,
		oidDeposito,
		lector,
		filtro.SoloNoLeidas == "true",
		pagina,
		tamanioPagina,
	)
	if err != nil {
		return NovedadesPaginadoDTO{}, err
	}

	items := make([]NovedadDTO, 0, len(novedades))

	for _, n := range novedades {
		items = append(items, n.ConvertirADTO(lector))
	}

	return NovedadesPaginadoDTO{
		Items:         items,
		Total:         total,
		NoLeidas:      noLeidas,
		Pagina:        pagina,
		TamanioPagina: tamanioPagina,
	}, nil
}

// Obtener devuelve una novedad puntual,
// validando acceso al depósito al que pertenece.
func (s *Service) Obtener(
	ctx context.Context,
	rol,
	depositoUsuario,
	id,
	usuarioID string,
) (NovedadDTO, error) {

	n, err := s.repositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return NovedadDTO{}, err
	}

	if err := verificarAccesoDeposito(
		rol,
		depositoUsuario,
		n.DepositoID.Hex(),
	); err != nil {
		return NovedadDTO{}, err
	}

	lector, err := lectorParaRol(rol, usuarioID)
	if err != nil {
		return NovedadDTO{}, err
	}

	return n.ConvertirADTO(lector), nil
}

// MarcarLeida marca la novedad como leída para el operario autenticado,
// sin responderla.
func (s *Service) MarcarLeida(
	ctx context.Context,
	rol,
	depositoUsuario,
	id,
	usuarioID string,
) (NovedadDTO, error) {

	n, err := s.repositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return NovedadDTO{}, err
	}

	if err := verificarAccesoDeposito(
		rol,
		depositoUsuario,
		n.DepositoID.Hex(),
	); err != nil {
		return NovedadDTO{}, err
	}

	oidUsuario, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return NovedadDTO{}, err
	}

	actualizada, err := s.repositorio.MarcarLeida(ctx, id, oidUsuario)
	if err != nil {
		return NovedadDTO{}, err
	}

	return actualizada.ConvertirADTO(&oidUsuario), nil
}

// Responder registra la respuesta del operario y,
// en el mismo paso, marca la novedad como leída para ese operario.
func (s *Service) Responder(
	ctx context.Context,
	rol,
	depositoUsuario,
	id,
	usuarioID string,
	dto ResponderNovedadDTO,
) (NovedadDTO, error) {

	n, err := s.repositorio.EncontrarPorId(ctx, id)
	if err != nil {
		return NovedadDTO{}, err
	}

	if err := verificarAccesoDeposito(
		rol,
		depositoUsuario,
		n.DepositoID.Hex(),
	); err != nil {
		return NovedadDTO{}, err
	}

	oidUsuario, err := bson.ObjectIDFromHex(usuarioID)
	if err != nil {
		return NovedadDTO{}, err
	}

	respuesta := RespuestaNovedad{
		Texto:          dto.Texto,
		UsuarioID:      oidUsuario,
		FechaRespuesta: time.Now(),
	}

	actualizada, err := s.repositorio.Responder(
		ctx,
		id,
		respuesta,
	)
	if err != nil {
		return NovedadDTO{}, err
	}

	return actualizada.ConvertirADTO(&oidUsuario), nil
}
