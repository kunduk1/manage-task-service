package main

import (
	"context"
	"log"

	"github.com/kunduk1/manage-task-service/internal/app"
)

//go:generate go tool swag init -g cmd/main.go -o docs --v3.1 --parseDependency --parseInternal

// @title           Manage Task Service API
// @version         1.0
// @description     Auth/tasks/teams API for manage-task-service.
// Схема указана прямо в @host: swag v2 --v3.1 склеивает @host+@BasePath в
// servers[].url дословно и игнорирует @schemes, поэтому без http:// тут URL
// вышел бы относительным и импортёры (Postman) роняли бы префикс /api/v1.
// @host            http://localhost:8080
// @BasePath        /api/v1
// swag умеет только apiKey/basic/oauth2, поэтому JWT описан как apiKey-заголовок;
// tools/swagfix переписывает его в http/bearer-схему (иначе Postman импортит как
// "API Key" и шлёт токен без префикса "Bearer " → 401). См. tools/swagfix/main.go.
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
func main() {
	ctx := context.Background()

	a, err := app.NewApp(ctx)
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	if err := a.Run(); err != nil {
		log.Fatalf("failed to run app: %v", err)
	}
}
