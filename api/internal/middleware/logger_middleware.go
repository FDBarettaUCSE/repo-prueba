package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		inicio := time.Now()
		ctx.Next()
		duracion := time.Since(inicio)
		log.Printf("%s %s -> %d (%s)", ctx.Request.Method, ctx.Request.URL.Path, ctx.Writer.Status(), duracion)
	}
}
