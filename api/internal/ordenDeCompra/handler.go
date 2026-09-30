package ordendecompra

import (
	"api/internal/middleware"
	ordendecompraitem "api/internal/ordenDeCompraItem"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrdenDeCompraHandler struct {
	ordenDeCompraService *OrdenDeCompraService
}

func NuevoHandler(ordenDeCompraService *OrdenDeCompraService) *OrdenDeCompraHandler {
	return &OrdenDeCompraHandler{
		ordenDeCompraService: ordenDeCompraService,
	}
}

func (h *OrdenDeCompraHandler) CrearOrdenCompra(c *gin.Context) {

	var dto CrearOrdenCompraDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
		})
		return
	}

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	orden, err := h.ordenDeCompraService.CrearOrdenCompra(
		c.Request.Context(),
		dto,
		usuario.UsuarioID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, orden)
}

func (h *OrdenDeCompraHandler) EncontrarOrdenCompraPorId(c *gin.Context) {

	id := c.Param("id")

	orden, err := h.ordenDeCompraService.EncontrarPorId(
		c.Request.Context(),
		id,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, orden)
}

func (h *OrdenDeCompraHandler) EncontrarOrdenesCompra(c *gin.Context) {

	ordenes, err := h.ordenDeCompraService.EncontrarTodos(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ordenes)
}

func (h *OrdenDeCompraHandler) ActualizarOrdenCompra(c *gin.Context) {

	var dto ActualizarOrdenCompraDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
		})
		return
	}

	dto.ID = c.Param("id")

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	orden, err := h.ordenDeCompraService.ActualizarOrdenCompra(
		c.Request.Context(),
		dto,
		usuario.UsuarioID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, orden)
}

func (h *OrdenDeCompraHandler) AgregarItem(c *gin.Context) {

	var dto ordendecompraitem.CrearOrdenCompraItemDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
		})
		return
	}

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	ordenCompraID := c.Param("id")

	item, err := h.ordenDeCompraService.AgregarItem(
		c.Request.Context(),
		ordenCompraID,
		dto,
		usuario.UsuarioID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *OrdenDeCompraHandler) ActualizarItem(c *gin.Context) {

	var dto ordendecompraitem.ActualizarOrdenCompraItemDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
		})
		return
	}

	dto.ID = c.Param("itemId")

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	ordenCompraID := c.Param("id")

	item, err := h.ordenDeCompraService.ActualizarItem(
		c.Request.Context(),
		ordenCompraID,
		dto,
		usuario.UsuarioID,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, item)
}

// Revisar caso de eliminacion fisica o logica
func (h *OrdenDeCompraHandler) EliminarItem(c *gin.Context) {

	_, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	ordenCompraID := c.Param("id")
	itemID := c.Param("itemId")

	err := h.ordenDeCompraService.EliminarItem(
		c.Request.Context(),
		ordenCompraID,
		itemID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *OrdenDeCompraHandler) ConfirmarOrdenCompra(c *gin.Context) {

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	id := c.Param("id")

	orden, err := h.ordenDeCompraService.ConfirmarOrdenCompra(
		c.Request.Context(),
		usuario.UsuarioID,
		id,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, orden)
}

func (h *OrdenDeCompraHandler) DescartarOrdenCompra(c *gin.Context) {

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	id := c.Param("id")

	err := h.ordenDeCompraService.DescartarOrdenCompra(
		c.Request.Context(),
		usuario.UsuarioID,
		id,
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
	handler *OrdenDeCompraHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {

	ordenes := router.Group("/ordenes-compra")

	ordenes.Use(authMiddleware)

	// Crear orden en estado borrador.
	ordenes.POST(
		"",
		rolMiddleware("administrador", "gerente"),
		handler.CrearOrdenCompra,
	)

	// Consultar órdenes.
	ordenes.GET(
		"",
		rolMiddleware("administrador", "gerente", "operario", "auditor"),
		handler.EncontrarOrdenesCompra,
	)

	ordenes.GET(
		"/:id",
		rolMiddleware("administrador", "gerente", "operario", "auditor"),
		handler.EncontrarOrdenCompraPorId,
	)

	// Actualizar datos de la orden.
	ordenes.PUT(
		"/:id",
		rolMiddleware("administrador", "gerente"),
		handler.ActualizarOrdenCompra,
	)

	// Agregar un producto a la orden.
	ordenes.POST(
		"/:id/items",
		rolMiddleware("administrador", "gerente"),
		handler.AgregarItem,
	)

	// Modificar un producto de la orden.
	ordenes.PUT(
		"/:id/items/:itemId",
		rolMiddleware("administrador", "gerente"),
		handler.ActualizarItem,
	)

	// Eliminar un producto de la orden.
	ordenes.DELETE(
		"/:id/items/:itemId",
		rolMiddleware("administrador", "gerente"),
		handler.EliminarItem,
	)

	// Confirmar orden.
	ordenes.PUT(
		"/:id/confirmar",
		rolMiddleware("administrador", "gerente"),
		handler.ConfirmarOrdenCompra,
	)

	// Descartar orden.
	ordenes.PUT(
		"/:id/descartar",
		rolMiddleware("administrador", "gerente"),
		handler.DescartarOrdenCompra,
	)
}
