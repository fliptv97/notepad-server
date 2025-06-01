package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/fliptv97/notepad-server/internal/repositories"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteHandler struct {
	repo *repositories.NoteRepository
}

func NewNoteHandler(repo *repositories.NoteRepository) *NoteHandler {
	return &NoteHandler{
		repo: repo,
	}
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	var reqBody RequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if reqBody.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("'title' is required field"))
		return
	}
	if reqBody.Content == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("'content' is required field"))
		return
	}

	_, err := h.repo.Create(r.Context(), reqBody.Title, reqBody.Content)
	if err != nil {
		fmt.Printf("[ERROR] POST /note: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	notes, err := h.repo.GetAll(r.Context())
	if err != nil {
		fmt.Printf("[ERROR] GET /note: %s\n", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(notes)
	if err != nil {
		fmt.Printf("[ERROR] GET /note: %s\n", err.Error())
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

	note, err := h.repo.GetById(r.Context(), id)
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

	note, err := h.repo.GetById(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
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
	err = h.repo.Update(r.Context(), id, note.Title, note.Content)
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

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Provided id is invalid"))
		return
	}

	if _, err = h.repo.GetById(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Printf("[ERROR] DELETE /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = h.repo.Delete(r.Context(), id); err != nil {
		fmt.Printf("[ERROR] DELETE /note/%s: %s\n", rawId, err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
