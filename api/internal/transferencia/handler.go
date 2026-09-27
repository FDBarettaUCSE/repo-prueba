package transferencia

import (
	"api/internal/middleware"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NuevoHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Solicitar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var dto SolicitarTransferenciaDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creada, err := h.service.Solicitar(c.Request.Context(), usr.UsuarioID, usr.UsuarioRol, usr.DepositoID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrMismoDeposito):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNoAutorizadoDeposito):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrStockInsuficiente):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, creada.ConvertirADTO())
}

func (h *Handler) Listar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroTransferenciasDTO
	_ = c.ShouldBindQuery(&filtro)

	items, err := h.service.Listar(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]TransferenciaDTO, 0, len(items))
	for _, it := range items {
		dtos = append(dtos, it.ConvertirADTO())
	}
	c.JSON(http.StatusOK, dtos)
}

func (h *Handler) Obtener(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	id := c.Param("id")
	t, err := h.service.ObtenerPorID(c.Request.Context(), id, usr.UsuarioRol, usr.DepositoID)
	if err != nil {
		if errors.Is(err, ErrNoAutorizadoDeposito) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, t.ConvertirADTO())
}

func (h *Handler) Aprobar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}
	id := c.Param("id")

	if err := h.service.Aprobar(c.Request.Context(), id, usr.UsuarioID, usr.UsuarioRol, usr.DepositoID); err != nil {
		switch {
		case errors.Is(err, ErrEstadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrSoloGerenteDestinoAdmin):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrStockInsuficiente):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensaje": "transferencia aprobada y stock de origen descontado"})
}

func (h *Handler) Rechazar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}
	id := c.Param("id")

	var dto ResolverTransferenciaDTO
	_ = c.ShouldBindJSON(&dto)

	if err := h.service.Rechazar(c.Request.Context(), id, usr.UsuarioID, usr.UsuarioRol, usr.DepositoID, dto.Motivo); err != nil {
		switch {
		case errors.Is(err, ErrEstadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrSoloGerenteDestinoAdmin):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensaje": "transferencia rechazada"})
}

func (h *Handler) Completar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}
	id := c.Param("id")

	if err := h.service.Completar(c.Request.Context(), id, usr.UsuarioID, usr.UsuarioRol, usr.DepositoID); err != nil {
		switch {
		case errors.Is(err, ErrEstadoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNoAutorizadoDeposito):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"mensaje": "transferencia completada y stock acreditado en destino"})
}

func RegistrarRutas(router *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc, rolMiddleware func(...string) gin.HandlerFunc) {
	g := router.Group("/transferencias", authMiddleware)
	{
		g.GET("", h.Listar)
		g.GET("/:id", h.Obtener)
		g.POST("", rolMiddleware("operario", "gerente", "administrador"), h.Solicitar)
		g.PUT("/:id/aprobar", rolMiddleware("gerente", "administrador"), h.Aprobar)
		g.PUT("/:id/rechazar", rolMiddleware("gerente", "administrador"), h.Rechazar)
		g.PUT("/:id/completar", rolMiddleware("operario", "gerente", "administrador"), h.Completar)
	}
}
