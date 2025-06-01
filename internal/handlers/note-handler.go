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

func (nh *NoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	type RequestBody struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	var reqBody RequestBody

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		nh.respondWithError(w, http.StatusBadRequest)
		return
	}

	if reqBody.Title == "" {
		nh.respondWithError(w, http.StatusBadRequest, "'title' is required field")
		return
	}

	if reqBody.Content == "" {
		nh.respondWithError(w, http.StatusBadRequest, "'content' is required field")
		return
	}

	note, err := nh.repo.Create(r.Context(), reqBody.Title, reqBody.Content)
	if err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("POST /note: %s\n", err.Error()))
		return
	}

	nh.respondWithJSON(w, http.StatusCreated, note)
}

func (nh *NoteHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	notes, err := nh.repo.GetAll(r.Context())

	if err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("GET /note: %s\n", err.Error()))
		return
	}

	if err = nh.respondWithJSON(w, http.StatusOK, notes); err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("GET /note: %s\n", err.Error()))
		return
	}
}

func (nh *NoteHandler) GetById(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)

	if err != nil {
		nh.respondWithError(w, http.StatusBadRequest, "Provided id is invalid")
		return
	}

	note, err := nh.repo.GetById(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			nh.respondWithError(w, http.StatusNotFound)
			return
		}

		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("GET /note/%s: %s\n", rawId, err.Error()))
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(note); err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("GET /note/%s: %s\n", rawId, err.Error()))
		return
	}
}

func (nh *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)

	if err != nil {
		nh.respondWithError(w, http.StatusBadRequest, "Provided id is invalid")
		return
	}

	type requestBody struct {
		Title   *string `json:"title,omitempty"`
		Content *string `json:"content,omitempty"`
	}

	var reqBody requestBody

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("PATCH /note/%s: %s\n", rawId, err.Error()))
		return
	}

	if reqBody.Title == nil && reqBody.Content == nil {
		nh.respondWithError(w, http.StatusBadRequest, "Request should contain at least one property to update")
		return
	}

	note, err := nh.repo.GetById(r.Context(), id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			nh.respondWithError(w, http.StatusNotFound)
			return
		}

		nh.respondWithError(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("PATCH /note/%s: %s\n", rawId, err.Error()),
		)
		return
	}

	if reqBody.Title != nil {
		note.Title = *reqBody.Title
	}

	if reqBody.Content != nil {
		note.Content = *reqBody.Content
	}

	note.UpdatedAt = time.Now().UTC()
	err = nh.repo.Update(r.Context(), id, note.Title, note.Content)

	if err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("PATCH /note/%s: %s\n", rawId, err.Error()))
		return
	}

	if err := nh.respondWithJSON(w, http.StatusOK, note); err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("PATCH /note/%s: %s\n", rawId, err.Error()))
		return
	}
}

func (nh *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	rawId := r.PathValue("id")
	id, err := uuid.Parse(rawId)

	if err != nil {
		nh.respondWithError(w, http.StatusBadRequest, "Provided id is invalid")
		return
	}

	if _, err = nh.repo.GetById(r.Context(), id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			nh.respondWithError(w, http.StatusNotFound)
			return
		}

		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("DELETE /note/%s: %s\n", rawId, err.Error()))
		return
	}

	if err = nh.repo.Delete(r.Context(), id); err != nil {
		nh.respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("DELETE /note/%s: %s\n", rawId, err.Error()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (nh *NoteHandler) respondWithError(w http.ResponseWriter, status int, messages ...string) {
	w.WriteHeader(status)

	if len(messages) == 0 {
		return
	}

	if status >= http.StatusInternalServerError {
		fmt.Printf("[ERROR] %s\n", messages[0])
	} else if messages[0] != "" {
		w.Write([]byte(messages[0]))
	}
}

func (nh *NoteHandler) respondWithJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
