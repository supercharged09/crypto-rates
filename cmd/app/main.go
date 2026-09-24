package main

import (
	"github.com/supercharged09/crypto-rates/internal/app"
	"github.com/supercharged09/crypto-rates/internal/config"
)

func main() {
	cfg := config.MustLoad()
	app.Start(cfg)
}
