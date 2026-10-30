package proveedor

import (
	"errors"
	"net/http"

	"api/internal/middleware"

	"github.com/gin-gonic/gin"
)

type ProveedorHandler struct {
	proveedorService *ProveedorService
}

func NuevoProveedorHandler(
	service *ProveedorService,
) *ProveedorHandler {
	return &ProveedorHandler{
		proveedorService: service,
	}
}

func obtenerAdminID(c *gin.Context) (string, bool) {

	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return "", false
	}

	return usrCtx.UsuarioID, true
}

func responderError(c *gin.Context, err error) {

	switch {

	case errors.Is(err, ErrNombreProveedorVacio),
		errors.Is(err, ErrEmailProveedorInvalido),
		errors.Is(err, ErrCUITInvalido),
		errors.Is(err, ErrCostoInvalido),
		errors.Is(err, ErrTiempoEntregaInvalido):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrCUITYaRegistrado),
		errors.Is(err, ErrProductoYaSuministrado):

		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrProveedorInexistente),
		errors.Is(err, ErrProductoInexistente),
		errors.Is(err, ErrProductoSuministradoInexistente):

		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

	default:

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

func (h *ProveedorHandler) ListarTodos(c *gin.Context) {

	proveedores, err :=
		h.proveedorService.ListarTodosProveedores(
			c.Request.Context(),
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, proveedores)
}

func (h *ProveedorHandler) BuscarPorID(c *gin.Context) {

	id := c.Param("id")

	proveedor, err :=
		h.proveedorService.BuscarProveedorPorID(
			c.Request.Context(),
			id,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, proveedor)
}

func (h *ProveedorHandler) CrearProveedor(c *gin.Context) {

	var dto CrearProveedorDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	creado, err :=
		h.proveedorService.CrearProveedor(
			c.Request.Context(),
			adminID,
			dto,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, creado)
}

func (h *ProveedorHandler) ActualizarProveedor(c *gin.Context) {

	id := c.Param("id")

	var dto ActualizarProveedorDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	actualizado, err :=
		h.proveedorService.ActualizarProveedor(
			c.Request.Context(),
			id,
			adminID,
			dto,
		)

	if err != nil {
		responderError(c, err)
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

	if err := h.proveedorService.EliminarProveedor(
		c.Request.Context(),
		id,
		adminID,
	); err != nil {

		responderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProveedorHandler) ListarProductosSuministrados(
	c *gin.Context,
) {

	id := c.Param("id")

	productos, err :=
		h.proveedorService.ListarProductosSuministrados(
			c.Request.Context(),
			id,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, productos)
}

func (h *ProveedorHandler) ListarProveedoresDeProducto(
	c *gin.Context,
) {

	productoID := c.Param("productoId")

	proveedores, err :=
		h.proveedorService.ListarProveedoresDeProducto(
			c.Request.Context(),
			productoID,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, proveedores)
}

func (h *ProveedorHandler) AgregarProductoSuministrado(c *gin.Context) {
	id := c.Param("id")

	var dto AgregarProductoSuministradoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	if err := h.proveedorService.AgregarProductoSuministrado(
		c.Request.Context(), adminID, id, dto,
	); err != nil {
		responderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProveedorHandler) ActualizarProductoSuministrado(c *gin.Context) {
	proveedorID := c.Param("id")
	productoID := c.Param("productoId")

	var dto ActualizarProductoSuministradoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	if err := h.proveedorService.ActualizarProductoSuministrado(
		c.Request.Context(),
		adminID,
		proveedorID,
		productoID,
		dto,
	); err != nil {
		responderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProveedorHandler) EliminarProductoSuministrado(
	c *gin.Context,
) {
	proveedorID := c.Param("id")
	productoID := c.Param("productoId")

	adminID, ok := obtenerAdminID(c)
	if !ok {
		return
	}

	if err := h.proveedorService.EliminarProductoSuministrado(
		c.Request.Context(),
		adminID,
		proveedorID,
		productoID,
	); err != nil {
		responderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func RegistrarRutas(
	router *gin.Engine,
	proveedorHandler *ProveedorHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {

	proveedores := router.Group("/proveedores")

	{
		proveedores.GET(
			"",
			authMiddleware,
			proveedorHandler.ListarTodos,
		)

		proveedores.GET(
			"/:id",
			authMiddleware,
			proveedorHandler.BuscarPorID,
		)

		proveedores.GET(
			"/:id/productos",
			authMiddleware,
			proveedorHandler.ListarProductosSuministrados,
		)

		proveedores.GET(
			"/producto/:productoId",
			authMiddleware,
			proveedorHandler.ListarProveedoresDeProducto,
		)

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
