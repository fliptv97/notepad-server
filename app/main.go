package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/fliptv97/notepad-server/internal/handlers"
	"github.com/fliptv97/notepad-server/internal/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	databaseUser := os.Getenv("DATABASE_USER")
	if databaseUser == "" {
		log.Fatal("DATABASE_USER environment variable is required")
	}

	databasePassword := os.Getenv("DATABASE_PASSWORD")

	databaseHost := os.Getenv("DATABASE_HOST")
	if databaseHost == "" {
		log.Fatal("DATABASE_HOST environment variable is required")
	}

	databasePort := os.Getenv("DATABASE_PORT")
	if databasePort == "" {
		log.Fatal("DATABASE_PORT environment variable is required")
	}

	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		log.Fatal("DATABASE_NAME environment variable is required")
	}

	connString := fmt.Sprintf(
		`postgres://%s:%s@%s:%s/%s`,
		databaseUser,
		databasePassword,
		databaseHost,
		databasePort,
		databaseName,
	)
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := conn.Close(context.Background())
		if err != nil {
			log.Fatal(err)
		}
	}()

	noteRepository := repositories.NewNoteRepository(conn)
	noteHandler := handlers.NewNoteHandler(noteRepository)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /note", noteHandler.Create)
	mux.HandleFunc("GET /note", noteHandler.GetAll)
	mux.HandleFunc("GET /note/{id}", noteHandler.GetById)
	mux.HandleFunc("PATCH /note/{id}", noteHandler.Update)
	mux.HandleFunc("DELETE /note/{id}", noteHandler.Delete)

	s := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	fmt.Printf("Starting server on :%s\n", port)
	err = s.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
