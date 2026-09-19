package main

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/app"
)

var cfgPath string = "config/"

func main() {
	ctx := context.Background()
	app.Run(ctx)
}
