package depositoproducto

import (
	"api/internal/middleware"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NuevoHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func responderSegunError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrAccesoDenegado):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, ErrProductoInexistente):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
}

// ListarStock responde GET /depositos/:depositoId/stock, con filtros
// opcionales ?categoriaId= y ?bajoMinimo=true, y paginación ?pagina=&tamañoPagina=.
func (h *Handler) ListarStock(c *gin.Context) {
	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	depositoID := c.Param("depositoId")

	pagina, _ := strconv.Atoi(c.Query("pagina"))
	tamañoPagina, _ := strconv.Atoi(c.Query("tamañoPagina"))

	filtro := ListarStockFiltro{
		CategoriaID:    c.Query("categoriaId"),
		SoloBajoMinimo: c.Query("bajoMinimo") == "true",
		Pagina:         pagina,
		TamañoPagina:   tamañoPagina,
	}

	resultado, err := h.service.ListarStockDeposito(
		c.Request.Context(),
		usrCtx.UsuarioID,
		usrCtx.UsuarioRol,
		depositoID,
		filtro,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// ObtenerStock responde GET /depositos/:depositoId/stock/:productoId.
func (h *Handler) ObtenerStock(c *gin.Context) {
	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	dto, err := h.service.ObtenerStockProducto(
		c.Request.Context(),
		usrCtx.UsuarioID,
		usrCtx.UsuarioRol,
		c.Param("depositoId"),
		c.Param("productoId"),
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto)
}

// ConfigurarStockMinimo responde PUT /depositos/:depositoId/stock/:productoId/stock-minimo.
func (h *Handler) ConfigurarStockMinimo(c *gin.Context) {
	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	var dto ConfigurarStockMinimoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultado, err := h.service.ConfigurarStockMinimo(
		c.Request.Context(),
		usrCtx.UsuarioID,
		usrCtx.UsuarioRol,
		c.Param("depositoId"),
		c.Param("productoId"),
		dto,
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// QuitarStockMinimo responde DELETE /depositos/:depositoId/stock/:productoId/stock-minimo:
// el depósito vuelve a regirse por el stock mínimo general del producto.
func (h *Handler) QuitarStockMinimo(c *gin.Context) {
	usrCtx, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "usuario no autenticado"})
		return
	}

	resultado, err := h.service.QuitarStockMinimo(
		c.Request.Context(),
		usrCtx.UsuarioID,
		usrCtx.UsuarioRol,
		c.Param("depositoId"),
		c.Param("productoId"),
	)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// RegistrarRutas agrupa las rutas de stock bajo /depositos/:depositoId/stock.
// Van agrupadas ahí (y no bajo /productos) porque el mismo producto tiene
// UNA fila de stock distinta por cada depósito — el recurso es "el stock de
// este depósito", no "el producto".
func RegistrarRutas(
	router *gin.Engine,
	handler *Handler,
	authMiddleware gin.HandlerFunc,
	rolMiddleware func(...string) gin.HandlerFunc,
) {
	stock := router.Group("/depositos/:depositoId/stock")
	{
		stock.GET(
			"",
			authMiddleware,
			rolMiddleware("administrador", "gerente", "operario", "auditor"),
			handler.ListarStock,
		)

		stock.GET(
			"/:productoId",
			authMiddleware,
			rolMiddleware("administrador", "gerente", "operario", "auditor"),
			handler.ObtenerStock,
		)

		stock.PUT(
			"/:productoId/stock-minimo",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			handler.ConfigurarStockMinimo,
		)

		stock.DELETE(
			"/:productoId/stock-minimo",
			authMiddleware,
			rolMiddleware("administrador", "gerente"),
			handler.QuitarStockMinimo,
		)
	}
}
