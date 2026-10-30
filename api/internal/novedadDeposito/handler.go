package novedaddeposito

import (
	"api/internal/middleware"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NuevoHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func responderSegunError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAccesoDenegado):
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrSinDestinatarios):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrNovedadNoEncontrada):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

// EnviarADeposito responde POST /depositos/:depositoId/novedades.
func (h *Handler) EnviarADeposito(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var dto EnviarNovedadDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	novedad, err := h.service.EnviarADeposito(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		c.Param("depositoId"),
		usr.UsuarioID,
		dto,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusCreated, novedad)
}

// EnviarDifusion responde POST /novedades/difusion.
func (h *Handler) EnviarDifusion(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var dto EnviarNovedadDifusionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	novedades, err := h.service.EnviarDifusion(
		c.Request.Context(),
		usr.UsuarioID,
		dto,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusCreated, novedades)
}

// ListarDeDeposito responde GET /depositos/:depositoId/novedades.
func (h *Handler) ListarDeDeposito(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	pagina, err := strconv.Atoi(
		c.DefaultQuery("pagina", "1"),
	)
	if err != nil || pagina < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pagina debe ser un entero >= 1",
		})
		return
	}

	tamanioStr := c.Query("tamanioPagina")
	tamanioPagina, _ := strconv.Atoi(tamanioStr)

	if tamanioPagina < 1 {
		tamanioPagina = 20
	} else if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	filtro := FiltroNovedadesDTO{
		SoloNoLeidas:  c.Query("solo_no_leidas"),
		Pagina:        pagina,
		TamanioPagina: tamanioPagina,
	}

	resultado, err := h.service.ListarDeDeposito(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		c.Param("depositoId"),
		usr.UsuarioID,
		filtro,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Obtener responde GET /depositos/:depositoId/novedades/:id.
func (h *Handler) Obtener(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	novedad, err := h.service.Obtener(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		c.Param("id"),
		usr.UsuarioID,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, novedad)
}

// MarcarLeida responde PUT /depositos/:depositoId/novedades/:id/leida.
func (h *Handler) MarcarLeida(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	novedad, err := h.service.MarcarLeida(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		c.Param("id"),
		usr.UsuarioID,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, novedad)
}

// Responder responde PUT /depositos/:depositoId/novedades/:id/responder.
func (h *Handler) Responder(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var dto ResponderNovedadDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	novedad, err := h.service.Responder(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		c.Param("id"),
		usr.UsuarioID,
		dto,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, novedad)
}

// RegistrarRutas registra las rutas de novedades.
func RegistrarRutas(
	router *gin.Engine,
	h *Handler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {
	novedades := router.Group(
		"/depositos/:depositoId/novedades",
		authMiddleware,
	)

	{
		novedades.POST(
			"",
			rolMiddleware("administrador", "gerente", "auditor"),
			h.EnviarADeposito,
		)

		novedades.GET(
			"",
			rolMiddleware("administrador", "operario", "gerente", "auditor"),
			h.ListarDeDeposito,
		)

		novedades.GET(
			"/:id",
			rolMiddleware("administrador", "operario", "gerente", "auditor"),
			h.Obtener,
		)

		novedades.PUT(
			"/:id/leida",
			rolMiddleware("operario"),
			h.MarcarLeida,
		)

		novedades.PUT(
			"/:id/responder",
			rolMiddleware("operario"),
			h.Responder,
		)
	}

	router.POST(
		"/novedades/difusion",
		authMiddleware,
		rolMiddleware("administrador", "auditor"),
		h.EnviarDifusion,
	)
}
