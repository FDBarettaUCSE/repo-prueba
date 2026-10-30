package alertastock

import (
	"errors"
	"net/http"
	"strconv"

	"api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Handler es la ÚNICA pieza del módulo que interactúa con Gin (HTTP, JSON,
// query params). Convierte los parámetros del request al formato que espera
// el Service y convierte la respuesta del Service a JSON.
type Handler struct {
	service *Service
}

func NuevoHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListarAlertas responde GET /alertas-stock — devuelve las alertas de stock
// paginadas, con filtros opcionales por depósito y por estado de resolución.
func (h *Handler) ListarAlertas(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	pagina, err := strconv.Atoi(c.DefaultQuery("pagina", "1"))
	if err != nil || pagina < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pagina debe ser un entero >= 1"})
		return
	}

	tamanioStr := c.Query("tamanioPagina")
	if tamanioStr == "" {
		tamanioStr = c.Query("tamaño_pagina")
	}

	tamanioPagina, _ := strconv.Atoi(tamanioStr)
	if tamanioPagina < 1 {
		tamanioPagina = 20
	} else if tamanioPagina > MaxTamanioPagina {
		tamanioPagina = MaxTamanioPagina
	}

	filtro := FiltroAlertasDTO{
		DepositoID:     c.Query("deposito_id"),
		SoloPendientes: c.Query("solo_pendientes"),
		Pagina:         pagina,
		TamanioPagina:  tamanioPagina,
	}

	res, err := h.service.ListarAlertas(
		c.Request.Context(),
		usr.UsuarioRol,
		usr.DepositoID,
		filtro,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func responderSegunError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAccesoDenegado):
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

// RegistrarRutas agrupa todas las rutas del módulo bajo /alertas-stock.
func RegistrarRutas(
	router *gin.Engine,
	handler *Handler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {
	alertas := router.Group("/alertas-stock")
	alertas.Use(authMiddleware)
	{
		// Consultar alertas — todos los roles.
		alertas.GET(
			"",
			rolMiddleware("administrador", "gerente", "operario", "auditor"),
			handler.ListarAlertas,
		)

	}
}
