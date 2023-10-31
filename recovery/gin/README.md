# Logrus formatter for Oops

```go
import oopsrecoverygin "github.com/samber/oops/recovery/gin"

func main() {
	router = gin.New()
	router.Use(gin.Logger())
	router.Use(oopsrecoverygin.GinOopsRecovery())

    // ...
}
```
