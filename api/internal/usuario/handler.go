package usuario

import (
	"api/internal/middleware"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UsuarioHandler struct {
	usuarioService *UsuarioService
}

func NuevoUsuarioHandler(service *UsuarioService) *UsuarioHandler {
	return &UsuarioHandler{
		usuarioService: service,
	}
}

func (usuarioHandler *UsuarioHandler) Registrar(c *gin.Context) {

	//Valida el DTO recibido
	var dto RegistrarDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//Busca dentro del header del contexto si existe el id del admin
	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	adminID := usrCtx.UsuarioID

	// Paso 2: el service valida el email duplicado y recién ahí hashea con
	// bcrypt (ver Service.Registrar).
	creado, err := usuarioHandler.usuarioService.Registrar(c.Request.Context(), adminID, dto)
	if err != nil {
		if errors.Is(err, ErrEmailYaRegistrado) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Paso 3: responder con UsuarioDTO — nunca con el hash de la contraseña.
	c.JSON(http.StatusCreated, creado.ConvertirAUsuarioDTO(creado))
}

// Login responde POST /usuarios/login.
func (usuarioHandler *UsuarioHandler) Ingresar(c *gin.Context) {
	var dto IngresarDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := usuarioHandler.usuarioService.Ingresar(c.Request.Context(), dto)
	if err != nil {
		// Tanto "email no existe" como "contraseña incorrecta" llegan acá
		// como ErrCredencialesInvalidas — el mismo 401 y el mismo mensaje
		// para los dos casos (ver Service.Login).
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Perfil responde GET /usuarios/me. Esta ruta va DETRÁS de
// middleware.AuthMiddleware() (ver RegisterRoutes) — el ID del usuario sale
// del JWT, nunca de un parámetro de ruta, así cada cuenta solo puede pedir
// SU PROPIO perfil (ver el mismo criterio en CambiarPassword).
func (usuarioHandler *UsuarioHandler) Perfil(c *gin.Context) {
	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	usuario, err := usuarioHandler.usuarioService.ObtenerPerfil(c.Request.Context(), usuarioCtx.UsuarioID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, usuario.ConvertirAUsuarioDTO(usuario))
}

func (usuarioHandler *UsuarioHandler) CambiarContraseña(c *gin.Context) {
	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	// Paso 2: parsear y validar el body (passwordActual + passwordNueva).
	idUsuario := c.Param("id")
	var dto CambiarContraseñaDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Paso 3: el service verifica passwordActual contra el hash guardado y,
	// si coincide, guarda el hash de passwordNueva. El ID viene del JWT
	// (usuarioCtx), nunca del body — así nadie puede pedir cambiar la
	// contraseña de otra cuenta.
	err := usuarioHandler.usuarioService.CambiarContraseña(c.Request.Context(), idUsuario, usuarioCtx.UsuarioID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrPasswordActualInvalida):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (usuarioHandler *UsuarioHandler) CambiarRolUsuario(ctx *gin.Context) {
	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	idUsuario := ctx.Param("id")
	var dto CambiarRolDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := usuarioHandler.usuarioService.CambiarRolUsuario(
		ctx.Request.Context(),
		idUsuario,
		usuarioCtx.UsuarioID,
		dto)

	if err != nil {
		switch {
		case errors.Is(err, ErrRolInexistente):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, ErrUsuarioInexistente):
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (usuarioHandler *UsuarioHandler) CambiarEstadoUsuario(ctx *gin.Context) {
	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	idUsuario := ctx.Param("id")
	var dto CambiarEstadoUsuarioDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := usuarioHandler.usuarioService.CambiarEstadoUsuario(
		ctx.Request.Context(),
		idUsuario,
		usuarioCtx.UsuarioID,
		dto,
	)

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RegisterRoutes agrupa las rutas del dominio bajo /usuarios (ver Clase 5 —
// "router.Group() — organizar rutas por dominio"). authMiddleware se recibe
// por parámetro en vez de importarlo acá adentro: quién cablea las rutas
// (main.go) decide qué middleware aplicar, el paquete "usuario" no depende
// de una instancia global de "middleware".
func RegistrarRutas(
	router *gin.Engine,
	usuarioHandler *UsuarioHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc) {
	usuarios := router.Group("/usuarios")
	{
		usuarios.POST(
			"/registrar",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.Registrar)
		usuarios.POST(
			"/ingresar",
			usuarioHandler.Ingresar)
		// A partir de acá, ambas rutas necesitan saber quién está autenticado —
		// registro y login son, por definición, accesibles sin token todavía.
		usuarios.GET(
			"/perfil",
			authMiddleware,
			usuarioHandler.Perfil)
		usuarios.PUT(
			"/:id/cambiarcontraseña",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.CambiarContraseña)
		usuarios.PUT(
			"/:id/cambiarestado",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.CambiarEstadoUsuario)
		usuarios.PUT(
			"/:id/cambiarrol",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.CambiarRolUsuario)
	}
}
