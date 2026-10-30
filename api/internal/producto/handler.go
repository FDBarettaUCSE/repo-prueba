package producto

import (
	"api/internal/middleware"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductoHandler struct {
	productoService *ProductoService
}

func NuevoProductoHandler(
	service *ProductoService,
) *ProductoHandler {
	return &ProductoHandler{
		productoService: service,
	}
}

func responderError(c *gin.Context, err error) {

	switch {

	case errors.Is(err, ErrUnidadMedidaInvalida),
		errors.Is(err, ErrNombreProductoVacio),
		errors.Is(err, ErrCostoUnitarioInvalido),
		errors.Is(err, ErrStockMinimoInvalido):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrProductoInexistente),
		errors.Is(err, ErrCategoriaInexistente):

		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

	default:

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

func (h *ProductoHandler) ListarTodos(c *gin.Context) {

	pagina, err :=
		strconv.Atoi(
			c.DefaultQuery("pagina", "1"),
		)

	if err != nil || pagina < 1 {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "pagina debe ser un entero >= 1",
		})

		return
	}

	var filtro FiltroProductosDTO

	if err := c.ShouldBindQuery(&filtro); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	filtro.Pagina = pagina

	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 10
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	resultado, err :=
		h.productoService.ListarProductos(
			c.Request.Context(),
			filtro,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

func (h *ProductoHandler) BuscarPorID(c *gin.Context) {

	id := c.Param("id")

	producto, err :=
		h.productoService.BuscarProductoPorID(
			c.Request.Context(),
			id,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, producto)
}

func (h *ProductoHandler) CrearProducto(c *gin.Context) {

	var dto CrearProductoDTO

	if err := c.ShouldBindJSON(&dto); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	usrCtx, ok :=
		middleware.ObtenerUsuarioContexto(c)

	if !ok {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})

		return
	}

	productoCreado, err :=
		h.productoService.CrearProducto(
			c.Request.Context(),
			usrCtx.UsuarioID,
			dto,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(
		http.StatusCreated,
		productoCreado,
	)
}

func (h *ProductoHandler) ActualizarProducto(c *gin.Context) {

	id := c.Param("id")

	var dto ActualizarProductoDTO

	if err := c.ShouldBindJSON(&dto); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	usrCtx, ok :=
		middleware.ObtenerUsuarioContexto(c)

	if !ok {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})

		return
	}

	actualizado, err :=
		h.productoService.ActualizarProducto(
			c.Request.Context(),
			id,
			usrCtx.UsuarioID,
			dto,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		actualizado,
	)
}

func (h *ProductoHandler) EliminarProducto(c *gin.Context) {

	id := c.Param("id")

	usrCtx, ok :=
		middleware.ObtenerUsuarioContexto(c)

	if !ok {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})

		return
	}

	err :=
		h.productoService.EliminarProducto(
			c.Request.Context(),
			id,
			usrCtx.UsuarioID,
		)

	if err != nil {
		responderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ProductoHandler) ConfigurarStockMinimo(
	c *gin.Context,
) {

	id := c.Param("id")

	var dto StockMinimoDTO

	if err := c.ShouldBindJSON(&dto); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}

	usrCtx, ok :=
		middleware.ObtenerUsuarioContexto(c)

	if !ok {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})

		return
	}

	err :=
		h.productoService.ConfigurarStockMinimo(
			c.Request.Context(),
			id,
			usrCtx.UsuarioID,
			dto,
		)

	if err != nil {
		responderError(c, err)
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
			rolMiddleware(
				"administrador",
				"gerente",
			),
			productoHandler.CrearProducto,
		)

		productos.PUT(
			"/:id",
			authMiddleware,
			rolMiddleware(
				"administrador",
				"gerente",
			),
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
