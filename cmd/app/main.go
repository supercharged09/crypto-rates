package main

import (
	"github.com/supercharged09/crypto-rates/internal/app"
	"github.com/supercharged09/crypto-rates/internal/config"
)

// @title Crypto Rates API
// @version 1.0
// @description API для отслеживания курсов криптовалют
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
func main() {
	cfg := config.MustLoad()
	app.Start(cfg)
}
