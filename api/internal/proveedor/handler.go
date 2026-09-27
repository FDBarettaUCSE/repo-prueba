package proveedor

import (
	"api/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProveedorHandler struct {
	proveedorService *ProveedorService
}

func NuevoProveedorHandler(service *ProveedorService) *ProveedorHandler {
	return &ProveedorHandler{proveedorService: service}
}

func obtenerAdminID(c *gin.Context) (string, bool) {
	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return "", false
	}
	return usrCtx.UsuarioID, true
}

func (h *ProveedorHandler) ListarTodos(c *gin.Context) {
	proveedores, err := h.proveedorService.ListarTodosProveedores(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proveedores)
}

func (h *ProveedorHandler) BuscarPorID(c *gin.Context) {
	id := c.Param("id")

	proveedor, err := h.proveedorService.BuscarProveedorPorID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proveedor)
}

func (h *ProveedorHandler) CrearProveedor(c *gin.Context) {
	var dto CrearProveedorDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	creado, err := h.proveedorService.CrearProveedor(c.Request.Context(), adminID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, creado)
}

func (h *ProveedorHandler) ActualizarProveedor(c *gin.Context) {
	id := c.Param("id")

	var dto ActualizarProveedorDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	actualizado, err := h.proveedorService.ActualizarProveedor(c.Request.Context(), id, adminID, dto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, actualizado)
}

func (h *ProveedorHandler) EliminarProveedor(c *gin.Context) {
	id := c.Param("id")

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	if err := h.proveedorService.EliminarProveedor(c.Request.Context(), id, adminID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProveedorHandler) ListarProductosSuministrados(c *gin.Context) {
	id := c.Param("id")

	productos, err := h.proveedorService.ListarProductosSuministrados(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, productos)
}

// ListarProveedoresDeProducto vive en este handler (no en el de producto)
// porque el dato que arma la respuesta —costo y tiempo de entrega— es propio
// del vínculo producto-proveedor, no del producto en sí.
func (h *ProveedorHandler) ListarProveedoresDeProducto(c *gin.Context) {
	productoID := c.Param("productoId")

	proveedores, err := h.proveedorService.ListarProveedoresDeProducto(c.Request.Context(), productoID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, proveedores)
}

func (h *ProveedorHandler) AgregarProductoSuministrado(c *gin.Context) {
	id := c.Param("id")

	var dto AgregarProductoSuministradoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.proveedorService.AgregarProductoSuministrado(c.Request.Context(), id, dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProveedorHandler) ActualizarProductoSuministrado(c *gin.Context) {
	id := c.Param("id")
	productoID := c.Param("productoId")

	var dto ActualizarProductoSuministradoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.proveedorService.ActualizarProductoSuministrado(c.Request.Context(), id, productoID, dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *ProveedorHandler) EliminarProductoSuministrado(c *gin.Context) {
	id := c.Param("id")
	productoID := c.Param("productoId")

	if err := h.proveedorService.EliminarProductoSuministrado(c.Request.Context(), id, productoID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// RegistrarRutas: alta/edición/baja de proveedores y de los productos que
// suministran quedan para administrador y gerente (el enunciado les da a
// ambos roles la gestión de proveedores); la consulta queda abierta a
// cualquier usuario autenticado, incluido el auditor.
func RegistrarRutas(
	router *gin.Engine,
	proveedorHandler *ProveedorHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {
	proveedores := router.Group("/proveedores")
	{
		proveedores.GET("", authMiddleware, proveedorHandler.ListarTodos)
		proveedores.GET("/:id", authMiddleware, proveedorHandler.BuscarPorID)
		proveedores.GET("/:id/productos", authMiddleware, proveedorHandler.ListarProductosSuministrados)
		proveedores.GET("/producto/:productoId", authMiddleware, proveedorHandler.ListarProveedoresDeProducto)

		proveedores.POST(
			"",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			proveedorHandler.CrearProveedor,
		)
		proveedores.PUT(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			proveedorHandler.ActualizarProveedor,
		)
		proveedores.DELETE(
			"/:id",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			proveedorHandler.EliminarProveedor,
		)

		proveedores.POST(
			"/:id/productos",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			proveedorHandler.AgregarProductoSuministrado,
		)
		proveedores.PUT(
			"/:id/productos/:productoId",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			proveedorHandler.ActualizarProductoSuministrado,
		)
		proveedores.DELETE(
			"/:id/productos/:productoId",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			proveedorHandler.EliminarProductoSuministrado,
		)
	}
}
