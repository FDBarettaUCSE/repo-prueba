// Package middleware agrupa los gin.HandlerFunc que corren antes de que la
// request llegue al handler final (ver Clase 3 — "Middlewares en Gin"). Es un
// paquete transversal, igual que "auth" y "db": ningún dominio de negocio
// debería reimplementar su propia validación de JWT.
package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"api/internal/auth"
)

// ErrUsuarioNoHabilitado es el error que devuelve ValidadorUsuario
// cuando el usuario del token ya no existe o fue desactivado.
// Cualquier otro error se toma como una falla interna (500).
var ErrUsuarioNoHabilitado = errors.New("el usuario no existe o fue desactivado")

// ValidadorUsuario define el contrato para consultar el estado actual del
// usuario en la base de datos en cada request autenticada.
type ValidadorUsuario interface {
	ObtenerUsuarioContexto(ctx context.Context, usuarioID string) (UsuarioContexto, error)
}

// AuthMiddleware valida el JWT del header "Authorization: Bearer <token>" y,
// si es válido, consulta a ValidadorUsuario para confirmar que el usuario
// siga activo en la base y cargar su rol y depósito actuales en la request.
func AuthMiddleware(validador ValidadorUsuario) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Paso 1: el header tiene que estar presente. Sin él no hay forma de
		// saber quién es el cliente.
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "falta el header Authorization"})
			return
		}

		// Paso 2: el formato esperado es EXACTAMENTE "Bearer <token>" — dos
		// partes separadas por un espacio. strings.SplitN con límite 2 evita
		// problemas si el token en sí tuviera un espacio (no debería, pero no
		// cuesta nada ser explícito con el límite).
		partes := strings.SplitN(header, " ", 2)
		if len(partes) != 2 || partes[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato esperado: Bearer <token>"})
			return
		}

		// Paso 3: decodificar y validar la firma y expiración del JWT. Un
		// token alterado, firmado con otra clave, o vencido, cae en este
		// mismo error — el cliente no necesita (ni debería) distinguir cuál
		// de esos tres pasó.
		claims, err := auth.ValidarToken(partes[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
			return
		}

		// Paso 4: el claim "sub" es el ID del usuario (ver auth.GenerarToken).
		// Viene como interface{} porque jwt.MapClaims es un map[string]any —
		// hay que confirmar que efectivamente es un string antes de usarlo.
		usuarioID, ok := claims["sub"].(string)
		if !ok || usuarioID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido o expirado"})
			return
		}

		// Paso 5: un token con firma válida no alcanza. Se busca al usuario
		// en la base para rechazarlo si fue desactivado (o eliminado)
		// mientras el token seguía vigente, y para tomar su rol y depósito
		// ACTUALES: si un admin le cambió el rol o el depósito, el token
		// viejo no conserva los permisos anteriores.
		usuarioActual, err := validador.ObtenerUsuarioContexto(c.Request.Context(), usuarioID)
		if err != nil {
			if errors.Is(err, ErrUsuarioNoHabilitado) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Paso 6: trasladar el DTO al contexto de esta request.
		setUsuarioContexto(c, usuarioActual)

		// Paso 7: recién acá se le cede el control al siguiente eslabón
		// (otro middleware, o el handler final). Si alguno de los checks de
		// arriba hubiera fallado, c.Abort...() ya cortó la cadena y este
		// c.Next() nunca se alcanza.
		c.Next()
	}
}
