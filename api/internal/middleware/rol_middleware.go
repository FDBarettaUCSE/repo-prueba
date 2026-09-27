package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequiereRolMiddleware(rolesPermitidos ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		usrCtx, existe := ObtenerUsuarioContexto(ctx)
		if !existe {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no autenticado"})
			return
		}

		for _, permitido := range rolesPermitidos {
			if usrCtx.UsuarioRol == permitido {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no tenés permiso para esta acción"})
	}
}
