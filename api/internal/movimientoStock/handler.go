package movimientostock

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

func (h *Handler) Registrar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var dto RegistrarMovimientoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creado, err := h.service.Registrar(c.Request.Context(), usr.UsuarioID, usr.UsuarioRol, usr.DepositoID, dto)
	if err != nil {
		switch {
		case errors.Is(err, ErrStockInsuficiente):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNoAutorizadoDeposito):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrTipoMovimientoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, creado.ConvertirADTO())
}

func (h *Handler) Listar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroMovimientosDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	movimientos, err := h.service.Listar(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dtos := make([]MovimientoStockDTO, 0, len(movimientos))
	for _, m := range movimientos {
		dtos = append(dtos, m.ConvertirADTO())
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
	m, err := h.service.ObtenerPorID(c.Request.Context(), id, usr.UsuarioRol, usr.DepositoID)
	if err != nil {
		if errors.Is(err, ErrNoAutorizadoDeposito) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m.ConvertirADTO())
}

func RegistrarRutas(router *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc, rolMiddleware func(...string) gin.HandlerFunc) {
	g := router.Group("/movimientos", authMiddleware)
	{
		// Operario, Gerente y Administrador pueden registrar movimientos
		g.POST("", rolMiddleware("operario", "gerente", "administrador"), h.Registrar)
		// Todos los roles autenticados pueden consultar movimientos según sus permisos de depósito
		g.GET("", h.Listar)
		g.GET("/:id", h.Obtener)
	}
}
