package main

import (
	"context"

	"github.com/maxon2034/trainee-go-cart-api/internal/app"
)

func main() {
	ctx := context.Background()
	app.Run(ctx)
}
