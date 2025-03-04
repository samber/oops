package oopsrecoverygin

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/oops"
)

func GinOopsRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := oops.Recoverf(func() {
			c.Next()
		}, "gin: panic recovered")
		if err != nil {
			_ = c.Error(err)
			c.AbortWithStatus(500)
		}
	}
}
