package categoria

import (
	"api/internal/middleware"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CategoriaHandler struct {
	categoriaService *CategoriaService
}

func NuevoCategoriaHandler(service *CategoriaService) *CategoriaHandler {
	return &CategoriaHandler{
		categoriaService: service,
	}
}

func (h *CategoriaHandler) ListarTodos(c *gin.Context) {
	// Esto parece de más
	usuarioContexto, _ := c.Get("usuarioContexto")
	fmt.Println(usuarioContexto)
	categoriasDTO, err := h.categoriaService.ListarTodasCategorias(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categoriasDTO)
}

func (h *CategoriaHandler) BuscarPorID(c *gin.Context) {
	id := c.Param("id")

	categoria, err := h.categoriaService.BuscarCategoriaPorID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, categoria)
}

func (h *CategoriaHandler) CrearCategoria(c *gin.Context) {
	var dto CrearCategoriaDTO
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

	categoriaCreado, err := h.categoriaService.CrearCategoria(c.Request.Context(), adminID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, categoriaCreado)
}

func (h *CategoriaHandler) ActualizarCategoria(c *gin.Context) {
	id := c.Param("id")

	var dto CategoriaDTO
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

	actualizado, err := h.categoriaService.ActualizarCategoria(c.Request.Context(), id, adminID, dto)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actualizado)
}

func (h *CategoriaHandler) EliminarCategoria(c *gin.Context) {
	id := c.Param("id")

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	adminID := usrCtx.UsuarioID

	if err := h.categoriaService.EliminarCategoria(c.Request.Context(), id, adminID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func RegistrarRutas(
	router *gin.Engine,
	categoriaHandler *CategoriaHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc) {
	categorias := router.Group("/categorias")
	{
		categorias.GET(
			"",
			authMiddleware,
			rolMiddleware("administrador"),
			categoriaHandler.ListarTodos)
		categorias.GET(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador"),
			categoriaHandler.BuscarPorID)
		categorias.POST(
			"",
			authMiddleware,
			rolMiddleware("administrador"),
			categoriaHandler.CrearCategoria)
		categorias.PUT(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador"),
			categoriaHandler.ActualizarCategoria)
		categorias.DELETE(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador"),
			categoriaHandler.EliminarCategoria)
	}
}
