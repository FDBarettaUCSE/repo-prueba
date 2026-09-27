package usuario

import (
	"api/internal/auth"
	"api/internal/deposito"
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrEmailYaRegistrado      = errors.New("ya existe un usuario registrado con ese email")
	ErrCredencialesInvalidas  = errors.New("credenciales inválidas")
	ErrPasswordActualInvalida = errors.New("la contraseña actual no es correcta")
	ErrDepositoInexistente    = errors.New("el depósito ingresado no existe")
	ErrRolInexistente         = errors.New("el rol ingresado no existe")
	ErrUsuarioInexistente     = errors.New("el usuario solicitado no existe")
)

type UsuarioService struct {
	usuarioRepositorio  UsuarioRepositorio
	depositoRepositorio deposito.DepositoRepositorio
}

func NuevoUsuarioService(repoUsuario UsuarioRepositorio, repoDeposito deposito.DepositoRepositorio) *UsuarioService {
	return &UsuarioService{usuarioRepositorio: repoUsuario, depositoRepositorio: repoDeposito}
}

func (s *UsuarioService) Registrar(ctx context.Context, adminId string, dto RegistrarDTO) (Usuario, error) {
	_, err := s.usuarioRepositorio.EncontarPorCorreo(ctx, dto.Correo)
	if err == nil {
		// Si err es nil, significa que encontró un usuario previo con ese correo
		return Usuario{}, ErrEmailYaRegistrado
	}

	//Los gerentes y operarios deben estar asociados a un deposito,
	//Administradores y auditores no.
	requiereDeposito := dto.Rol == Gerente || dto.Rol == Operario
	tieneDeposito := dto.DepositoAsociado != ""

	if requiereDeposito && !tieneDeposito {
		return Usuario{}, errors.New("los roles Gerente u Operario requieren un depósito asociado obligatorio")
	}

	if !requiereDeposito && tieneDeposito {
		return Usuario{}, errors.New("los roles Administrador u Auditor no deben estar asociados a un depósito")
	}

	// 1. Convertir el ID del admin que viene por parámetro (adminId)
	adId, err := bson.ObjectIDFromHex(adminId)
	if err != nil {
		return Usuario{}, err
	}

	// 2. Convertir el ID del depósito que viene en el DTO (dto.DepositoAsociado)
	depoId, err := bson.ObjectIDFromHex(dto.DepositoAsociado)
	if err != nil {
		return Usuario{}, err
	}

	hash, err := auth.HashPassword(dto.Contraseña)
	if err != nil {
		return Usuario{}, err
	}

	// 3. Crear la entidad a través de tu constructor y desreferenciar el puntero (*)
	nuevoUsuario := dto.NuevoUsuarioRegistrar(dto, hash, adId, depoId)

	return s.usuarioRepositorio.CrearUsuario(ctx, *nuevoUsuario)
}

func (s *UsuarioService) Ingresar(ctx context.Context, dto IngresarDTO) (string, error) {

	//Primero busca un usuario con correo, si no lo encuentra err va a ser distinto de nil
	u, err := s.usuarioRepositorio.EncontarPorCorreo(ctx, dto.Correo)
	if err != nil {
		return "", ErrCredencialesInvalidas
	}

	//Compara la contraseña enviada por el cliente con aquella almacenada en la base de datos
	if err := auth.CheckPassword(u.ContraseñaHasheada, dto.Contraseña); err != nil {
		return "", ErrCredencialesInvalidas
	}

	depoID := ""
	if !u.DepositoAsociado.IsZero() {
		depoID = u.DepositoAsociado.Hex()
	}

	return auth.GenerarToken(u.ID.Hex(), string(u.Rol), depoID)
}

// ObtenerPerfil devuelve el usuario autenticado. usuarioID llega desde el JWT
// ya validado por AuthMiddleware (ver internal/middleware), nunca de un
// parámetro de ruta — así una cuenta solo puede pedir SU PROPIO perfil, igual
// que en CambiarPassword.
func (s *UsuarioService) ObtenerPerfil(ctx context.Context, usuarioID string) (Usuario, error) {
	return s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
}

// Funcion de cambiar contraseña. No estoy seguro de si es requerido. Preguntar
func (s *UsuarioService) CambiarContraseña(ctx context.Context, usuarioID, adminID string, dto CambiarContraseñaDTO) error {
	u, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return err
	}

	if err := auth.CheckPassword(u.ContraseñaHasheada, dto.Contraseña); err != nil {
		return ErrPasswordActualInvalida
	}

	nuevoHash, err := auth.HashPassword(dto.ContraseñaNueva)
	if err != nil {
		return err
	}

	return s.usuarioRepositorio.ActualizarContrasenaHash(ctx, usuarioID, adminID, nuevoHash)
}

// El administrador modifica el rol de los distintos usuarios
// Se verifica si el rol está presente en el modelo
func (s *UsuarioService) CambiarRolUsuario(ctx context.Context, usuarioID, adminID string, dto CambiarRolDTO) error {

	//Valida que el rol ingresado se ajuste a algún valor interno.
	//Complementa al binding del DTO

	if !esRolValido(Rol(dto.Rol)) {
		return ErrRolInexistente
	}

	//Buscamos el usuario para ver si existe en la base de datos
	_, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return ErrUsuarioInexistente
	}

	//Los gerentes y operarios deben estar asociados a un deposito,
	//Administradores y auditores no.
	requiereDeposito := dto.Rol == Gerente || dto.Rol == Operario
	tieneDeposito := dto.DepositoAsociado != ""

	if requiereDeposito && !tieneDeposito {
		return errors.New("los roles Gerente u Operario requieren un depósito asociado obligatorio")
	}

	if !requiereDeposito && tieneDeposito {
		return errors.New("los roles Administrador u Auditor no deben estar asociados a un depósito")
	}

	return s.usuarioRepositorio.ActualizarRolUsuario(ctx, usuarioID, adminID, dto.DepositoAsociado, Rol(dto.Rol))
}

// El administrador puede activar y desactivar los distintos usuarios
// Se verifica únicamente que el usuario exista
func (s *UsuarioService) CambiarEstadoUsuario(ctx context.Context, usuarioID, adminID string, dto CambiarEstadoUsuarioDTO) error {
	_, err := s.usuarioRepositorio.EncontrarPorId(ctx, usuarioID)
	if err != nil {
		return ErrUsuarioInexistente
	}

	return s.usuarioRepositorio.ActivarUsuario(ctx, usuarioID, adminID, dto.estado)
}

// Funcion auxiliar, recibe un rol desde el cliente y revisa que exista en el model
func esRolValido(rol Rol) bool {
	switch rol {
	case Administrador, Gerente, Operario, Auditor:
		return true
	default:
		return false
	}
}
