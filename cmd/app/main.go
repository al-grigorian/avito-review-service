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

	teamRepo := repositories.NewTeamRepository(db)
	teamSvc := services.NewTeamService(teamRepo)
	teamHandler := handlers.NewTeamHandler(teamSvc)

	r := chi.NewRouter()
	r.Post("/team/add", teamHandler.AddTeam)
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
