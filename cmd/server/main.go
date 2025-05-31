package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	conn, err := pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("POST /note", func(w http.ResponseWriter, r *http.Request) {
		type Note struct {
			Title   string `json:"title"`
			Content string `json:"content"`
		}
		var note Note
		if err := json.NewDecoder(r.Body).Decode(&note); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if note.Title == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("'title' is required field"))
			return
		}
		if note.Content == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("'content' is required field"))
			return
		}

		_, err := conn.Exec(
			r.Context(),
			"INSERT INTO note (id, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
			uuid.New(),
			note.Title,
			note.Content,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			fmt.Printf("[ERROR] POST /note: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /note", func(w http.ResponseWriter, r *http.Request) {
		rows, err := conn.Query(r.Context(), "SELECT * FROM note")
		if err != nil {
			fmt.Printf("[ERROR] GET /note: %s\n", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type Note struct {
			Id        string    `json:"id"`
			Title     string    `json:"title"`
			Content   string    `json:"content"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
		}
		var note Note
		notes := []Note{}
		for rows.Next() {
			err := rows.Scan(&note.Id, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
			if err != nil {
				fmt.Printf("[ERROR] GET /note: %s\n", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			notes = append(notes, note)
		}

		if rows.Err() != nil {
			fmt.Printf("[ERROR] GET /note: %s\n", rows.Err().Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(notes)
		if err != nil {
			fmt.Printf("[ERROR] GET /note: %s\n", rows.Err().Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

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
