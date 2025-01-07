# Parsenv

Parse OS environment into a Go struct.

```go
import (
	"os"
	"log"
	"github.com/cvanloo/parsenv"
)

type AppConfig struct {
	ApiKey    string  `cfg:"required"`
	IsDebug   bool    `cfg:"default=false"`
	Coolness  float64 `cfg:"name=COOL_NESS;default=0.951"`
	FrobCount int
}

func main() {
	os.Setenv("API_KEY", "xxXXXXxxXXXxxXXxXXx")
	os.Setenv("COOL_NESS", "0.69")
	os.Setenv("FROB_COUNT", "99")

	var cfg AppConfig
	if err := parsenv.Load(&cfg); err != nil {
		log.Fatal(err)
	}

	log.Println(cfg)
	// Output: {xxXXXXxxXXXxxXXxXXx false 0.69 99}
}
```
