package reporte

import (
	"errors"
	"net/http"

	"api/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NuevoHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func responderSegunError(c *gin.Context, err error) {

	switch {

	case errors.Is(err, ErrAccesoDenegado):
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrDepositoNoAsignado):
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrDepositoInvalido):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrCategoriaInvalida):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrTipoMovimientoInvalido):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrFechaDesdeInvalida):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrFechaHastaInvalida):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrRangoFechasInvalido):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

// validarPagina valida el parámetro ?pagina=.
// Si no se especifica, utiliza la página 1.
func validarPagina(pagina *int) error {

	if *pagina == 0 {
		*pagina = 1
		return nil
	}

	if *pagina < 1 {
		return errors.New(
			"la página debe ser mayor o igual a 1",
		)
	}

	return nil
}

// Inventario responde GET /reportes/inventario.
func (h *Handler) Inventario(c *gin.Context) {

	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var filtro FiltroInventarioDTO

	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validarPagina(&filtro.Pagina); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resultado, err := h.service.ReporteInventario(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		filtro,
	)

	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Movimientos responde GET /reportes/movimientos.
func (h *Handler) Movimientos(c *gin.Context) {

	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var filtro FiltroReporteMovimientosDTO

	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validarPagina(&filtro.Pagina); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resultado, err := h.service.ReporteMovimientos(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		filtro,
	)

	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Alertas responde GET /reportes/alertas.
func (h *Handler) Alertas(c *gin.Context) {

	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var filtro FiltroAlertasDTO

	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validarPagina(&filtro.Pagina); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resultado, err := h.service.ReporteAlertas(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		filtro,
	)

	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Pendientes responde GET /reportes/pendientes.
func (h *Handler) Pendientes(c *gin.Context) {

	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "no autenticado",
		})
		return
	}

	var filtro FiltroPendientesDTO

	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := validarPagina(&filtro.Pagina); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	resultado, err := h.service.ReportePendientes(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		filtro,
	)

	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// RegistrarRutas registra las rutas de reportes.
func RegistrarRutas(
	router *gin.Engine,
	h *Handler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {
	reportes := router.Group(
		"/reportes",
		authMiddleware,
	)

	{
		reportes.GET(
			"/inventario",
			h.Inventario,
		)

		reportes.GET(
			"/movimientos",
			h.Movimientos,
		)

		reportes.GET(
			"/alertas",
			h.Alertas,
		)

		reportes.GET(
			"/pendientes",
			h.Pendientes,
		)
	}
}
