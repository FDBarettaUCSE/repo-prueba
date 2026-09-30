package novedaddeposito

import (
	"api/internal/middleware"
	"errors"
	"net/http"

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
	case errors.Is(err, ErrSinDestinatarios):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	}
}

// EnviarADeposito responde POST /depositos/:depositoId/novedades.
func (h *Handler) EnviarADeposito(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var dto EnviarNovedadDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	novedad, err := h.service.EnviarADeposito(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, c.Param("depositoId"), usr.UsuarioID, dto)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusCreated, novedad)
}

// EnviarDifusion responde POST /novedades/difusion (Auditor/Administrador).
func (h *Handler) EnviarDifusion(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var dto EnviarNovedadDifusionDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	novedades, err := h.service.EnviarDifusion(c.Request.Context(), usr.UsuarioID, dto)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusCreated, novedades)
}

// ListarDeDeposito responde GET /depositos/:depositoId/novedades.
func (h *Handler) ListarDeDeposito(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var filtro FiltroNovedadesDTO
	if err := c.ShouldBindQuery(&filtro); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resultado, err := h.service.ListarDeDeposito(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, c.Param("depositoId"), filtro)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, resultado)
}

// Obtener responde GET /depositos/:depositoId/novedades/:id.
func (h *Handler) Obtener(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	novedad, err := h.service.Obtener(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, c.Param("id"))
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, novedad)
}

// MarcarLeida responde PUT /depositos/:depositoId/novedades/:id/leida.
func (h *Handler) MarcarLeida(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	novedad, err := h.service.MarcarLeida(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, c.Param("id"))
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, novedad)
}

// Responder responde PUT /depositos/:depositoId/novedades/:id/responder.
func (h *Handler) Responder(c *gin.Context) {
	usr, ok := middleware.ObtenerUsuarioContexto(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
		return
	}

	var dto ResponderNovedadDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	novedad, err := h.service.Responder(c.Request.Context(), usr.UsuarioRol, usr.DepositoID, c.Param("id"), usr.UsuarioID, dto)
	if err != nil {
		responderSegunError(c, err)
		return
	}

	c.JSON(http.StatusOK, novedad)
}

// RegistrarRutas registra las rutas de novedades. La difusión a varios
// depósitos (o a todos) va aparte, en /novedades/difusion, porque no cuelga
// de un :depositoId puntual — es al revés, un mismo envío se abre en varios.
func RegistrarRutas(router *gin.Engine, h *Handler, authMiddleware gin.HandlerFunc, rolMiddleware func(...string) gin.HandlerFunc) {
	novedades := router.Group("/depositos/:depositoId/novedades", authMiddleware)
	{
		// Gerente (solo su depósito, validado en el service) y
		// Auditor/Administrador (cualquier depósito) envían novedades.
		novedades.POST("", rolMiddleware("gerente", "auditor", "administrador"), h.EnviarADeposito)

		// El Operario consulta las novedades de su propio depósito;
		// Gerente/Administrador/Auditor pueden ver las de cualquiera que
		// les corresponda (ver Service.verificarAccesoDeposito).
		novedades.GET("", rolMiddleware("operario", "gerente", "administrador", "auditor"), h.ListarDeDeposito)
		novedades.GET("/:id", rolMiddleware("operario", "gerente", "administrador", "auditor"), h.Obtener)

		// Marcar como leída y responder son acciones del Operario
		// destinatario (ver enunciado): Administrador se agrega
		// defensivamente, igual que en el resto de los módulos.
		novedades.PUT("/:id/leida", rolMiddleware("operario", "administrador"), h.MarcarLeida)
		novedades.PUT("/:id/responder", rolMiddleware("operario", "administrador"), h.Responder)
	}

	router.POST(
		"/novedades/difusion",
		authMiddleware,
		rolMiddleware("auditor", "administrador"),
		h.EnviarDifusion,
	)
}
