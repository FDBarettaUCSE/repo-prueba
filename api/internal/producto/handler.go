package producto

import (
	"api/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProductoHandler struct {
	productoService *ProductoService
}

func NuevoProductoHandler(service *ProductoService) *ProductoHandler {
	return &ProductoHandler{
		productoService: service,
	}
}

func (h *ProductoHandler) ListarTodos(c *gin.Context) {

	productosDTO, err := h.productoService.ListarTodosProductos(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, productosDTO)
}

func (h *ProductoHandler) BuscarPorID(c *gin.Context) {

	id := c.Param("id")

	producto, err := h.productoService.BuscarProductoPorID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, producto)
}

func (h *ProductoHandler) CrearProducto(c *gin.Context) {

	var dto CrearProductoDTO

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

	productoCreado, err := h.productoService.CrearProducto(
		c.Request.Context(),
		adminID,
		dto,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, productoCreado)
}

func (h *ProductoHandler) ActualizarProducto(c *gin.Context) {

	id := c.Param("id")

	var dto ActualizarProductoDTO

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

	actualizado, err := h.productoService.ActualizarProducto(
		c.Request.Context(),
		id,
		adminID,
		dto,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, actualizado)
}

func (h *ProductoHandler) EliminarProducto(c *gin.Context) {

	id := c.Param("id")

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	adminID := usrCtx.UsuarioID

	if err := h.productoService.EliminarProducto(
		c.Request.Context(),
		id,
		adminID,
	); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProductoHandler) ConfigurarStockMinimo(c *gin.Context) {
	id := c.Param("id")

	var dto StockMinimoDTO

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

	err := h.productoService.ConfigurarStockMinimo(
		c.Request.Context(),
		id,
		adminID,
		dto,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func RegistrarRutas(
	router *gin.Engine,
	productoHandler *ProductoHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {
	productos := router.Group("/productos")
	{
		productos.GET(
			"",
			authMiddleware,
			productoHandler.ListarTodos,
		)

		productos.GET(
			"/:id",
			authMiddleware,
			productoHandler.BuscarPorID,
		)

		productos.POST(
			"",
			authMiddleware,
			rolMiddleware("administrador"),
			productoHandler.CrearProducto,
		)

		productos.PUT(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			productoHandler.ActualizarProducto,
		)

		productos.PUT(
			"/:id/stock-minimo",
			authMiddleware,
			rolMiddleware("administrador"),
			productoHandler.ConfigurarStockMinimo,
		)

		productos.DELETE(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador"),
			productoHandler.EliminarProducto,
		)
	}
}
