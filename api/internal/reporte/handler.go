package reporte

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

func responderSegunError(c *gin.Context, err error) {
	if errors.Is(err, ErrAccesoDenegado) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

// Inventario responde GET /reportes/inventario: stock y valorización por
// depósito y por categoría. Sin Operario (ver enunciado: "no visualiza la
// valorización económica del inventario").
func (h *Handler) Inventario(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroInventarioDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultado, err := h.service.ReporteInventario(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Movimientos responde GET /reportes/movimientos: historial agregado por
// fecha, depósito y tipo.
func (h *Handler) Movimientos(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroReporteMovimientosDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultado, err := h.service.ReporteMovimientos(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Alertas responde GET /reportes/alertas: productos bajo mínimo y próximos a
// vencer. Disponible para todos los roles autenticados (todos los
// dashboards, incluido el del Operario, muestran alertas).
func (h *Handler) Alertas(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroAlertasDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultado, err := h.service.ReporteAlertas(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Pendientes responde GET /reportes/pendientes: estado de órdenes de compra
// y transferencias, agrupado por depósito.
func (h *Handler) Pendientes(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroPendientesDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultado, err := h.service.ReportePendientes(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, filtro)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// RegistrarRutas registra los cuatro reportes del enunciado bajo /reportes.
// Ninguno admite escritura (todos son GET) — son justamente el tipo de
// endpoint que el Auditor puede consultar sin ninguna restricción extra,
// porque su rol ya es de solo lectura por definición.
func RegistrarRutas(router *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc, rolMiddleware func(...string) gin.HandlerFunc) {
	reportes := router.Group("/reportes", authMiddleware)
	{
		reportes.GET("/inventario", rolMiddleware("administrador", "gerente", "auditor"), h.Inventario)
		reportes.GET("/movimientos", rolMiddleware("administrador", "gerente", "auditor"), h.Movimientos)
		reportes.GET("/alertas", rolMiddleware("administrador", "gerente", "operario", "auditor"), h.Alertas)
		reportes.GET("/pendientes", rolMiddleware("administrador", "gerente", "auditor"), h.Pendientes)
	}
}
