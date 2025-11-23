// cmd/app/main.go
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/al-grigorian/avito-review-service/internal/http/handlers"
	"github.com/al-grigorian/avito-review-service/internal/repositories"
	"github.com/al-grigorian/avito-review-service/internal/services"
	"github.com/al-grigorian/avito-review-service/pkg/db"

	"github.com/go-chi/chi/v5"
)

func main() {
	db := db.New()
	defer db.Close()

	// Репозитории
	teamRepo := repositories.NewTeamRepository(db)
	prRepo := repositories.NewPRRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// Сервисы
	teamSvc := services.NewTeamService(teamRepo)
	prSvc := services.NewPRService(prRepo)
	userSvc := services.NewUserService(userRepo)

	// Хендлеры
	teamHandler := handlers.NewTeamHandler(teamSvc)
	prHandler := handlers.NewPRHandler(prSvc)
	userHandler := handlers.NewUserHandler(userSvc)

	r := chi.NewRouter()

	r.Post("/team/add", teamHandler.AddTeam)
	r.Post("/pullRequest/create", prHandler.CreatePR)
	r.Get("/team/get", teamHandler.GetTeam)
	r.Post("/pullRequest/merge", prHandler.MergePR)
	r.Post("/users/setIsActive", userHandler.SetIsActive)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
