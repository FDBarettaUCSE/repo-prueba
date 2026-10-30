package usuario

import (
	"api/internal/auditoria"
	"api/internal/auth"
	"api/internal/deposito"
	"api/internal/middleware"
	"context"
	"errors"
	"os"
	"time"
	"unicode/utf8"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrEmailYaRegistrado      = errors.New("ya existe un usuario registrado con ese email")
	ErrCredencialesInvalidas  = errors.New("credenciales inválidas")
	ErrPasswordActualInvalida = errors.New("la contraseña actual no es correcta")
	ErrDepositoInexistente    = errors.New("el depósito ingresado no existe")
	ErrRolInexistente         = errors.New("el rol ingresado no existe")
	ErrUsuarioInexistente     = errors.New("el usuario solicitado no existe")
	ErrRolRequiereDeposito    = errors.New("el rol seleccionado requiere un depósito asociado")
	ErrRolNoPermiteDeposito   = errors.New("el rol seleccionado no puede tener un depósito asociado")
	ErrUsuarioDesactivado     = errors.New("el usuario se encuentra desactivado")
	ErrAdminInicialSinConfig  = errors.New("faltan las variables de entorno INITIAL_ADMIN_EMAIL y/o INITIAL_ADMIN_PASSWORD: son obligatorias para crear el administrador inicial con la base vacía (ver .env.example)")
	ErrAdminInicialPassword   = errors.New("INITIAL_ADMIN_PASSWORD debe tener entre 8 y 64 caracteres")
	ErrUltimoAdministrador    = errors.New("no se puede cambiar el rol o desactivar al último administrador activo")
)

type UsuarioService struct {
	usuarioRepositorio  UsuarioRepositorio
	depositoRepositorio deposito.DepositoRepositorio
}

func NuevoUsuarioService(repoUsuario UsuarioRepositorio, repoDeposito deposito.DepositoRepositorio) *UsuarioService {
	return &UsuarioService{
		usuarioRepositorio:  repoUsuario,
		depositoRepositorio: repoDeposito,
	}
}

func (s *UsuarioService) Registrar(
	ctx context.Context,
	dto RegistrarDTO,
	adminId string,
) (Usuario, error) {

	if !esRolValido(dto.Rol) {
		return Usuario{}, ErrRolInexistente
	}

	requiereDeposito := dto.Rol == Gerente || dto.Rol == Operario
	tieneDeposito := dto.DepositoAsociado != ""

	if requiereDeposito && !tieneDeposito {
		return Usuario{}, ErrRolRequiereDeposito
	}

	if !requiereDeposito && tieneDeposito {
		return Usuario{}, ErrRolNoPermiteDeposito
	}

	existe, err := s.usuarioRepositorio.ExisteCorreo(ctx, dto.Correo)
	if err != nil {
		return Usuario{}, err
	}
	if existe {
		return Usuario{}, ErrEmailYaRegistrado
	}

	adminOID, err := bson.ObjectIDFromHex(adminId)
	if err != nil {
		return Usuario{}, err
	}

	var depositoOID bson.ObjectID

	if tieneDeposito {
		depositoOID, err = bson.ObjectIDFromHex(dto.DepositoAsociado)
		if err != nil {
			return Usuario{}, err
		}

		_, err = s.depositoRepositorio.EncontrarPorId(
			ctx,
			dto.DepositoAsociado,
		)
		if err != nil {
			return Usuario{}, ErrDepositoInexistente
		}
	}

	hash, err := auth.HashPassword(dto.Contrasenia)
	if err != nil {
		return Usuario{}, err
	}

	nuevoUsuario := dto.NuevoUsuarioRegistrar(
		dto,
		hash,
		adminOID,
		depositoOID,
	)

	return s.usuarioRepositorio.CrearUsuario(ctx, *nuevoUsuario)
}

func (s *UsuarioService) Ingresar(
	ctx context.Context,
	dto IngresarDTO,
) (string, error) {

	u, err := s.usuarioRepositorio.EncontarPorCorreo(ctx, dto.Correo)
	if err != nil {
		return "", ErrCredencialesInvalidas
	}

	if err := auth.CheckPassword(
		u.ContraseniaHasheada,
		dto.Contrasenia,
	); err != nil {
		return "", ErrCredencialesInvalidas
	}

	depoID := ""

	if !u.DepositoAsociado.IsZero() {
		depoID = u.DepositoAsociado.Hex()
	}

	return auth.GenerarToken(
		u.ID.Hex(),
		string(u.Rol),
		depoID,
	)
}

func (s *UsuarioService) ObtenerPerfil(
	ctx context.Context,
	usuarioID string,
) (Usuario, error) {

	usuario, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return Usuario{}, ErrUsuarioInexistente
	}

	return usuario, nil
}

var _ middleware.ValidadorUsuario = (*UsuarioService)(nil)

// ObtenerUsuarioContexto implementa middleware.ValidadorUsuario para que
// AuthMiddleware pueda verificar en cada request si el usuario sigue activo
// en la base de datos y obtener su rol y depósito actuales en tiempo real.
func (s *UsuarioService) ObtenerUsuarioContexto(
	ctx context.Context,
	usuarioID string,
) (middleware.UsuarioContexto, error) {

	u, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		if errors.Is(err, ErrUsuarioInexistente) {
			return middleware.UsuarioContexto{}, middleware.ErrUsuarioNoHabilitado
		}
		return middleware.UsuarioContexto{}, err
	}

	if !u.Activado {
		return middleware.UsuarioContexto{}, middleware.ErrUsuarioNoHabilitado
	}

	depositoID := ""
	if !u.DepositoAsociado.IsZero() {
		depositoID = u.DepositoAsociado.Hex()
	}

	return middleware.UsuarioContexto{
		UsuarioID:  u.ID.Hex(),
		UsuarioRol: string(u.Rol),
		DepositoID: depositoID,
	}, nil
}

// ObtenerUsuarioActivo devuelve el estado actual del usuario y falla si ya no
// existe o fue desactivado.
func (s *UsuarioService) ObtenerUsuarioActivo(
	ctx context.Context,
	usuarioID string,
) (Usuario, error) {

	u, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return Usuario{}, err
	}

	if !u.Activado {
		return Usuario{}, ErrUsuarioDesactivado
	}

	return u, nil
}

func (s *UsuarioService) CambiarContrasenia(
	ctx context.Context,
	usuarioID string,
	adminID string,
	dto CambiarContraseniaDTO,
) error {

	u, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return ErrUsuarioInexistente
	}

	if err := auth.CheckPassword(
		u.ContraseniaHasheada,
		dto.Contrasenia,
	); err != nil {
		return ErrPasswordActualInvalida
	}

	nuevoHash, err := auth.HashPassword(dto.ContraseniaNueva)
	if err != nil {
		return err
	}

	return s.usuarioRepositorio.ActualizarContrasenaHash(
		ctx,
		usuarioID,
		adminID,
		nuevoHash,
	)
}

func (s *UsuarioService) CambiarRolUsuario(
	ctx context.Context,
	usuarioID string,
	adminID string,
	dto CambiarRolDTO,
) error {

	if !esRolValido(dto.Rol) {
		return ErrRolInexistente
	}

	requiereDeposito := dto.Rol == Gerente || dto.Rol == Operario
	tieneDeposito := dto.DepositoAsociado != ""

	if requiereDeposito && !tieneDeposito {
		return ErrRolRequiereDeposito
	}

	if !requiereDeposito && tieneDeposito {
		return ErrRolNoPermiteDeposito
	}

	usuario, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return ErrUsuarioInexistente
	}

	// Evitar que se quite el rol al último administrador activo.
	if usuario.Rol == Administrador && dto.Rol != Administrador {
		cantAdmins, err := s.usuarioRepositorio.ContarUsuarios(ctx, Administrador)
		if err != nil {
			return err
		}

		if cantAdmins <= 1 {
			return ErrUltimoAdministrador
		}
	}

	if tieneDeposito {
		_, err = s.depositoRepositorio.EncontrarPorId(
			ctx,
			dto.DepositoAsociado,
		)
		if err != nil {
			return ErrDepositoInexistente
		}
	}

	return s.usuarioRepositorio.ActualizarRolUsuario(
		ctx,
		usuarioID,
		adminID,
		dto.DepositoAsociado,
		dto.Rol,
	)
}

func (s *UsuarioService) DesactivarUsuario(
	ctx context.Context,
	usuarioID string,
	adminID string,
) error {

	usuario, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return ErrUsuarioInexistente
	}

	// Evitar desactivar al último administrador activo.
	if usuario.Rol == Administrador {
		cantAdmins, err := s.usuarioRepositorio.ContarUsuarios(ctx, Administrador)
		if err != nil {
			return err
		}

		if cantAdmins <= 1 {
			return ErrUltimoAdministrador
		}
	}

	return s.usuarioRepositorio.DesactivarUsuario(
		ctx,
		usuarioID,
		adminID,
	)
}

func (s *UsuarioService) ListarTodos(
	ctx context.Context,
) ([]UsuarioDTO, error) {

	usuarios, err := s.usuarioRepositorio.EncontrarTodos(ctx)
	if err != nil {
		return nil, err
	}

	dtos := make([]UsuarioDTO, 0, len(usuarios))

	for _, u := range usuarios {
		dtos = append(
			dtos,
			(&u).ConvertirAUsuarioDTO(u),
		)
	}

	return dtos, nil
}

func (s *UsuarioService) ListarPaginado(
	ctx context.Context,
	pagina, tamanioPagina int,
) (UsuariosPaginadosDTO, error) {

	if pagina < 1 {
		pagina = 1
	}
	if tamanioPagina < 1 {
		tamanioPagina = 20
	}
	if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	usuarios, total, err := s.usuarioRepositorio.Listar(ctx, pagina, tamanioPagina)
	if err != nil {
		return UsuariosPaginadosDTO{}, err
	}

	dtos := make([]UsuarioDTO, 0, len(usuarios))
	for _, u := range usuarios {
		dtos = append(dtos, (&u).ConvertirAUsuarioDTO(u))
	}

	return UsuariosPaginadosDTO{
		Items:         dtos,
		Total:         total,
		Pagina:        pagina,
		TamanioPagina: tamanioPagina,
	}, nil
}

func (s *UsuarioService) ActivarUsuario(
	ctx context.Context,
	usuarioID string,
	adminID string,
) error {

	_, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return ErrUsuarioInexistente
	}

	return s.usuarioRepositorio.ActivarUsuario(
		ctx,
		usuarioID,
		adminID,
	)
}

func (s *UsuarioService) InicializarAdmin(ctx context.Context) error {

	total, err := s.usuarioRepositorio.ContarUsuarios(ctx)
	if err != nil {
		return err
	}

	if total > 0 {
		return nil
	}

	// A propósito NO hay credenciales hardcodeadas de respaldo: un
	// email/contraseña conocidos de antemano dejarían a cualquier despliegue
	// con un administrador público y utilizable (mismo criterio que
	// JWT_SECRET, ver auth.RequerirSecretoConfigurado). Solo se exigen con la
	// base vacía; una vez creado el admin, estas variables ya no se usan.
	adminEmail := os.Getenv("INITIAL_ADMIN_EMAIL")
	adminPassword := os.Getenv("INITIAL_ADMIN_PASSWORD")
	if adminEmail == "" || adminPassword == "" {
		return ErrAdminInicialSinConfig
	}

	// Misma regla de longitud que RegistrarDTO (min=8, max=64; el validador
	// cuenta caracteres, no bytes).
	largoPassword := utf8.RuneCountInString(adminPassword)
	if largoPassword < 8 || largoPassword > 64 {
		return ErrAdminInicialPassword
	}

	hash, err := auth.HashPassword(adminPassword)
	if err != nil {
		return err
	}

	adminOID := bson.NewObjectID()
	ahora := time.Now()

	adminUser := Usuario{
		ID:                  adminOID,
		Correo:              adminEmail,
		ContraseniaHasheada: hash,
		Rol:                 Administrador,
		Activado:            true,
		Auditoria: auditoria.Auditoria{
			CreadoPor: adminOID,
			CreadoEn:  ahora,
		},
		HistorialAuditoria: []EventoAuditoriaUsuario{
			{
				Accion:       "alta_usuario",
				Detalle:      "rol=administrador (inicial)",
				RealizadoPor: adminOID,
				Fecha:        ahora,
			},
		},
	}

	_, err = s.usuarioRepositorio.CrearUsuario(
		ctx,
		adminUser,
	)

	return err
}

func esRolValido(rol Rol) bool {
	switch rol {
	case Administrador, Gerente, Operario, Auditor:
		return true
	default:
		return false
	}
}
