package oopsrecoverygin

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/oops"
)

// GinOopsRecovery creates a Gin middleware that recovers from panics and
// converts them to structured oops.OopsError instances.
//
// This middleware wraps the Gin context's Next() method in a panic recovery
// mechanism. When a panic occurs in any subsequent middleware or handler,
// it is caught and converted to an oops.OopsError with appropriate context
// and stack trace information.
//
// The middleware automatically:
// - Catches any panic that occurs in the request processing chain
// - Converts the panic to an oops.OopsError with "gin: panic recovered" message
// - Adds the error to the Gin context for logging or response handling
// - Aborts the request with a 500 Internal Server Error status
//
// Performance: This middleware has minimal overhead when no panics occur.
// When a panic is caught, the overhead includes error creation and stack
// trace generation, which is negligible compared to the panic recovery cost.
//
// Example usage:
//
//	router := gin.New()
//	router.Use(oopsrecoverygin.GinOopsRecovery())
//
//	router.GET("/api/users", func(c *gin.Context) {
//	  // This panic will be caught and converted to an oops error
//	  panic("something went wrong")
//	})
//
//	// Handle the recovered error in your error middleware
//	router.Use(func(c *gin.Context) {
//	  c.Next()
//	  if len(c.Errors) > 0 {
//	    // Log or handle the recovered error
//	    for _, err := range c.Errors {
//	      log.Printf("Recovered error: %+v", err)
//	    }
//	  }
//	})
func GinOopsRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use oops.Recoverf to catch panics and convert them to structured errors
		// The callback function executes the rest of the middleware/handler chain
		err := oops.Recoverf(func() {
			c.Next() // Continue processing the request
		}, "gin: panic recovered")

		// If a panic was recovered, handle the resulting error
		if err != nil {
			// Add the error to the Gin context for logging or response handling
			// This allows other middleware to access and process the recovered error
			_ = c.Error(err)

			// Abort the request with a 500 Internal Server Error status
			// This prevents the request from continuing and ensures the client
			// receives an appropriate error response
			c.AbortWithStatus(500)
		}
	}
}
