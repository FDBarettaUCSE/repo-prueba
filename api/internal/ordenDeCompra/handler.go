package ordendecompra

import (
	"api/internal/middleware"
	ordendecompraitem "api/internal/ordenDeCompraItem"
	"errors"
	"net/http"
	"strconv"

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

func responderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrDepositoNoPermitido),
		errors.Is(err, ErrDepositoNoPermitidoActualizar):
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrEstadoOrdenInvalido),
		errors.Is(err, ErrOrdenSinProductos),
		errors.Is(err, ErrCantidadRecibidaExcedeSolicitada):
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrCantidadSolicitadaInvalida),
		errors.Is(err, ErrItemNoPerteneceOrden),
		errors.Is(err, ErrItemRecepcionNoPerteneceOrden),
		errors.Is(err, ErrIDOrdenCompraInvalido),
		errors.Is(err, ErrDatosOrdenInvalidos),
		errors.Is(err, ErrRecepcionInvalida),
		errors.Is(err, ErrProductoNoSuministrado),
		errors.Is(err, ordendecompraitem.ErrIDOrdenCompraItemInvalido):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrOrdenCompraInexistente),
		errors.Is(err, ErrProveedorInexistente),
		errors.Is(err, ErrDepositoInexistente),
		errors.Is(err, ordendecompraitem.ErrOrdenCompraItemInexistente):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
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
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, orden)
}

func (h *OrdenDeCompraHandler) EncontrarOrdenCompraPorId(c *gin.Context) {

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	id := c.Param("id")

	orden, err := h.ordenDeCompraService.EncontrarPorId(
		c.Request.Context(),
		id,
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, orden)
}

func (h *OrdenDeCompraHandler) EncontrarOrdenesCompra(c *gin.Context) {

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	pagina, err := strconv.Atoi(c.DefaultQuery("pagina", "1"))
	if err != nil || pagina < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "pagina debe ser un entero >= 1"})
		return
	}

	var filtro FiltroOrdenesCompraDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	filtro.Pagina = pagina

	tamanioStr := c.Query("tamanioPagina")
	if tamanioStr != "" {
		if tamanio, err := strconv.Atoi(tamanioStr); err == nil && tamanio > 0 {
			filtro.TamanioPagina = tamanio
		}
	}
	if filtro.TamanioPagina < 1 {
		filtro.TamanioPagina = 20
	} else if filtro.TamanioPagina > MaxTamanioPagina {
		filtro.TamanioPagina = MaxTamanioPagina
	}

	res, err := h.ordenDeCompraService.Listar(
		c.Request.Context(),
		usuario.UsuarioRol,
		usuario.DepositoID,
		filtro,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *OrdenDeCompraHandler) ActualizarOrdenCompra(c *gin.Context) {

	var dto ActualizarOrdenCompraDTO

	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "datos inválidos",
		})
		return
	}

	id := c.Param("id")
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
		id,
		dto,
		usuario.UsuarioID,
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
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
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
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
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, item)
}

// Revisar caso de eliminacion fisica o logica
func (h *OrdenDeCompraHandler) EliminarItem(c *gin.Context) {

	usuario, ok := middleware.ObtenerUsuarioContexto(c)
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
		usuario.UsuarioRol,
		usuario.DepositoID,
	)

	if err != nil {
		responderError(c, err)
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
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
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
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *OrdenDeCompraHandler) RegistrarRecepcion(c *gin.Context) {

	var dto RegistrarRecepcionDTO

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

	orden, err := h.ordenDeCompraService.RegistrarRecepcion(
		c.Request.Context(),
		ordenCompraID,
		dto,
		usuario.UsuarioID,
		usuario.UsuarioRol,
		usuario.DepositoID,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, orden)
}

func (h *OrdenDeCompraHandler) GenerarSugerencias(c *gin.Context) {
	usuario, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "usuario no autenticado",
		})
		return
	}

	depositoID := c.Query("deposito_id")

	sugerencias, err := h.ordenDeCompraService.GenerarSugerenciasPorStockBajoMinimo(
		c.Request.Context(),
		usuario.UsuarioRol,
		usuario.DepositoID,
		depositoID,
		usuario.UsuarioID,
	)
	if err != nil {
		responderError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mensaje":           "sugerencias generadas correctamente",
		"ordenes_generadas": sugerencias,
		"cantidad":          len(sugerencias),
	})
}

func RegistrarRutas(
	router *gin.Engine,
	handler *OrdenDeCompraHandler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {

	ordenes := router.Group("/ordenes-compra")

	ordenes.Use(authMiddleware)

	// Sugerencia automática de órdenes de compra por stock bajo mínimo.
	// Va antes de /:id para que Gin no interprete "sugerir" como un ID.
	ordenes.POST(
		"/sugerir",
		rolMiddleware("administrador", "gerente"),
		handler.GenerarSugerencias,
	)

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

	// Agregar items al stock
	ordenes.POST(
		"/:id/recepciones",
		rolMiddleware("administrador", "gerente"),
		handler.RegistrarRecepcion,
	)
}
