package movimientostock

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
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, ErrNoAutorizadoDeposito):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrTipoMovimientoInvalido),
			errors.Is(err, ErrCantidadInvalida):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, ErrProductoInexistente),
			errors.Is(err, ErrDepositoInexistente):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, creado)
}

func (h *Handler) Listar(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	pagina, err := strconv.Atoi(c.DefaultQuery("pagina", "1"))
	if err != nil || pagina < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pagina debe ser un entero >= 1"})
		return
	}

	var filtro FiltroMovimientosDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filtro.Pagina = pagina

	tamanioStr := c.Query("tamanioPagina")
	if tamanioStr != "" {
		if tamanio, err := strconv.Atoi(tamanioStr); err == nil && tamanio > 0 {
			filtro.TamanioPagina = tamanio
		}
	}
	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 20
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	resultado, err := h.service.Listar(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resultado)
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
		switch {
		case errors.Is(err, ErrNoAutorizadoDeposito):
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case errors.Is(err, ErrMovimientoInexistente):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, ErrIDMovimientoInvalido):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, m)
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
