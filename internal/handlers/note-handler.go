package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fliptv97/notepad-server/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteHandler struct {
	conn *pgx.Conn
}

func NewNoteHandler(conn *pgx.Conn) *NoteHandler {
	return &NoteHandler{
		conn: conn,
	}
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
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

	_, err := h.conn.Exec(
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
}

func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	rows, err := h.conn.Query(r.Context(), "SELECT * FROM note")
	if err != nil {
		fmt.Printf("[ERROR] GET /note: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var note domain.Note
	notes := []domain.Note{}
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
}

func (h *NoteHandler) GetNoteById(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Provided id is invalid"))
		return
	}

	var note domain.Note
	row := h.conn.QueryRow(r.Context(), "SELECT * FROM note WHERE id=$1", id)
	err = row.Scan(&note.Id, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		fmt.Printf("[ERROR] GET /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(note); err != nil {
		fmt.Printf("[ERROR] GET /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Provided id is invalid"))
		return
	}

	type requestBody struct {
		Title   *string `json:"title,omitempty"`
		Content *string `json:"content,omitempty"`
	}
	var reqBody requestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		fmt.Printf("[ERROR] PATCH /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if reqBody.Title == nil && reqBody.Content == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Request should contain at least one property to update"))
		return
	}

	var note domain.Note
	queryStr := "SELECT id, title, content, created_at, updated_at FROM note WHERE id=$1"
	row := h.conn.QueryRow(r.Context(), queryStr, id)
	err = row.Scan(&note.Id, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Note with id '%s' doesn't exists", rawId)))
		}
		fmt.Printf("[ERROR] PATCH /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if reqBody.Title != nil {
		note.Title = *reqBody.Title
	}
	if reqBody.Content != nil {
		note.Content = *reqBody.Content
	}
	note.UpdatedAt = time.Now().UTC()
	_, err = h.conn.Exec(
		r.Context(),
		"UPDATE note SET title=$1, content=$2, updated_at=$3 WHERE id=$4",
		note.Title,
		note.Content,
		note.UpdatedAt,
		id,
	)
	if err != nil {
		fmt.Printf("[ERROR] PATCH /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(note); err != nil {
		fmt.Printf("[ERROR] PATCH /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
