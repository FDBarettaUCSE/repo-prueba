package deposito

import (
	"api/internal/middleware"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DepositoHandler struct {
	depositoService *DepositoService
}

func NuevoDepositoHandler(service *DepositoService) *DepositoHandler {
	return &DepositoHandler{
		depositoService: service,
	}
}

func (h *DepositoHandler) ListarTodos(c *gin.Context) {
	// Esto parece de más
	usuarioContexto, _ := c.Get("usuarioContexto")
	fmt.Println(usuarioContexto)
	depositosDTO, err := h.depositoService.ListarTodosDepositos(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, depositosDTO)
}

func (h *DepositoHandler) BuscarPorID(c *gin.Context) {
	id := c.Param("id")

	deposito, err := h.depositoService.BuscarDepositoPorID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deposito)
}

func (h *DepositoHandler) CrearDeposito(c *gin.Context) {
	var dto CrearDepositoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	adminID := usrCtx.UsuarioID

	depositoCreado, err := h.depositoService.CrearDeposito(c.Request.Context(), adminID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, depositoCreado)
}

func (h *DepositoHandler) ActualizarDeposito(c *gin.Context) {
	id := c.Param("id")

	var dto DepositoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	adminID := usrCtx.UsuarioID

	actualizado, err := h.depositoService.ActualizarDeposito(c.Request.Context(), id, adminID, dto)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actualizado)
}

func (h *DepositoHandler) EliminarDeposito(c *gin.Context) {
	id := c.Param("id")

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	adminID := usrCtx.UsuarioID

	if err := h.depositoService.EliminarDeposito(c.Request.Context(), id, adminID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func RegistrarRutas(
	router *gin.Engine,
	depositoHandler *DepositoHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc) {
	depositos := router.Group("/depositos")
	{
		depositos.GET(
			"",
			authMiddleware,
			depositoHandler.ListarTodos)
		depositos.GET(
			"/:id",
			authMiddleware,
			depositoHandler.BuscarPorID)
		depositos.POST(
			"",
			authMiddleware,
			rolMiddleware("administrador"),
			depositoHandler.CrearDeposito)
		depositos.PUT(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador"),
			depositoHandler.ActualizarDeposito)
		depositos.DELETE(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador"),
			depositoHandler.EliminarDeposito)
	}
}
