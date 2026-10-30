package usuario

import (
	"api/internal/middleware"
	"errors"
	"net/http"
	"strconv"

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

	var dto RegistrarDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	adminID := usrCtx.UsuarioID

	creado, err := usuarioHandler.usuarioService.Registrar(
		c.Request.Context(),
		dto,
		adminID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrEmailYaRegistrado):
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrRolInexistente),
			errors.Is(err, ErrRolRequiereDeposito),
			errors.Is(err, ErrRolNoPermiteDeposito),
			errors.Is(err, ErrDepositoInexistente):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(
		http.StatusCreated,
		creado.ConvertirAUsuarioDTO(creado),
	)
}

func (usuarioHandler *UsuarioHandler) Ingresar(c *gin.Context) {

	var dto IngresarDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	token, err := usuarioHandler.usuarioService.Ingresar(
		c.Request.Context(),
		dto,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func (usuarioHandler *UsuarioHandler) Perfil(c *gin.Context) {

	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	usuario, err := usuarioHandler.usuarioService.ObtenerPerfil(
		c.Request.Context(),
		usuarioCtx.UsuarioID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrUsuarioInexistente):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(
		http.StatusOK,
		usuario.ConvertirAUsuarioDTO(usuario),
	)
}

func (usuarioHandler *UsuarioHandler) CambiarContrasenia(c *gin.Context) {

	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	idUsuario := c.Param("id")

	var dto CambiarContraseniaDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := usuarioHandler.usuarioService.CambiarContrasenia(
		c.Request.Context(),
		idUsuario,
		usuarioCtx.UsuarioID,
		dto,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrPasswordActualInvalida):
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrUsuarioInexistente):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.Status(http.StatusNoContent)
}

func (usuarioHandler *UsuarioHandler) CambiarRolUsuario(c *gin.Context) {

	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	idUsuario := c.Param("id")

	var dto CambiarRolDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := usuarioHandler.usuarioService.CambiarRolUsuario(
		c.Request.Context(),
		idUsuario,
		usuarioCtx.UsuarioID,
		dto,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrRolInexistente),
			errors.Is(err, ErrRolRequiereDeposito),
			errors.Is(err, ErrRolNoPermiteDeposito),
			errors.Is(err, ErrDepositoInexistente):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrUsuarioInexistente):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.Status(http.StatusNoContent)
}

func (usuarioHandler *UsuarioHandler) DesactivarUsuario(c *gin.Context) {

	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	idUsuario := c.Param("id")

	err := usuarioHandler.usuarioService.DesactivarUsuario(
		c.Request.Context(),
		idUsuario,
		usuarioCtx.UsuarioID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrUsuarioInexistente):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.Status(http.StatusNoContent)
}

func (usuarioHandler *UsuarioHandler) ActivarUsuario(c *gin.Context) {

	usuarioCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	idUsuario := c.Param("id")

	err := usuarioHandler.usuarioService.ActivarUsuario(
		c.Request.Context(),
		idUsuario,
		usuarioCtx.UsuarioID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrUsuarioInexistente):
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.Status(http.StatusNoContent)
}

func (usuarioHandler *UsuarioHandler) Listar(c *gin.Context) {

	pagina, err := strconv.Atoi(c.DefaultQuery("pagina", "1"))
	if err != nil || pagina < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pagina debe ser un entero >= 1",
		})
		return
	}

	tamanioPagina := 20
	if tamanioStr := c.Query("tamanioPagina"); tamanioStr != "" {
		if t, err := strconv.Atoi(tamanioStr); err == nil && t > 0 {
			tamanioPagina = t
		}
	}
	if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	resultado, err := usuarioHandler.usuarioService.ListarPaginado(
		c.Request.Context(),
		pagina,
		tamanioPagina,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resultado)
}

func RegistrarRutas(
	router *gin.Engine,
	usuarioHandler *UsuarioHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {

	usuarios := router.Group("/usuarios")

	{
		usuarios.POST(
			"/registrar",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.Registrar,
		)

		usuarios.POST(
			"/ingresar",
			usuarioHandler.Ingresar,
		)

		usuarios.GET(
			"",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.Listar,
		)

		usuarios.GET(
			"/perfil",
			authMiddleware,
			usuarioHandler.Perfil,
		)

		usuarios.PUT(
			"/:id/cambiarcontrasenia",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.CambiarContrasenia,
		)

		usuarios.PUT(
			"/:id/desactivar",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.DesactivarUsuario,
		)

		usuarios.PUT(
			"/:id/activar",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.ActivarUsuario,
		)

		usuarios.PUT(
			"/:id/cambiarrol",
			authMiddleware,
			rolMiddleware("administrador"),
			usuarioHandler.CambiarRolUsuario,
		)
	}
}
