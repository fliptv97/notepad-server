package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/fliptv97/notepad-server/domain"
	"github.com/google/uuid"
)

func getMockRepo() *NoteRepositoryMock {
	return &NoteRepositoryMock{
		Notes: []domain.Note{
			{Id: uuid.New(), Title: "1st note", Content: "1st content", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{Id: uuid.New(), Title: "2nd note", Content: "2nd content", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
			{Id: uuid.New(), Title: "3rd note", Content: "3rd content", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		},
	}
}

func TestCreate(t *testing.T) {
	repo := &NoteRepositoryMock{}
	handler := NewNoteHandler(repo)

	body := `{ "title": "note title", "content": "note content" }`
	req := httptest.NewRequest("POST", "/note", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	statusCode := w.Result().StatusCode
	if statusCode != http.StatusCreated {
		t.Errorf("Expect HTTP status 201, but got %d", statusCode)
	}

	notesCount := len(repo.Notes)
	if notesCount != 1 {
		t.Errorf("Repository should contain one note, but contains %d", notesCount)
	}

	note := repo.Notes[0]
	if note.Title != "note title" {
		t.Errorf(`Expected note title to be "note title", but got "%s"`, note.Title)
	}
	if note.Content != "note content" {
		t.Errorf(`Expected note content to be "note content", but got "%s"`, note.Content)
	}
}

func TestGetAll(t *testing.T) {
	repo := getMockRepo()
	handler := NewNoteHandler(repo)

	req := httptest.NewRequest("GET", "/note", nil)
	w := httptest.NewRecorder()
	handler.GetAll(w, req)

	statusCode := w.Result().StatusCode
	if statusCode != http.StatusOK {
		t.Errorf("Expected status code to be %d, got %d", http.StatusOK, statusCode)
	}

	var notes []domain.Note
	if err := json.NewDecoder(w.Result().Body).Decode(&notes); err != nil {
		t.Fatalf("Error while trying to decode: %s", err.Error())
	}

	if len(repo.Notes) != len(notes) {
		t.Errorf("Expected %d notes, got %d", len(repo.Notes), len(notes))
	}

	for idx, note := range repo.Notes {
		if note.Id != notes[idx].Id || note.Title != notes[idx].Title || note.Content != notes[idx].Content {
			t.Errorf("Note %d doesn't match. Expected %+v, got %+v", idx, note, notes[idx])
		}
	}
}

func TestGetById(t *testing.T) {
	repo := getMockRepo()
	handler := NewNoteHandler(repo)

	expectedNote := repo.Notes[2]
	id := expectedNote.Id.String()
	req := httptest.NewRequest("GET", "/note/"+id, nil)
	req.SetPathValue("id", id)
	w := httptest.NewRecorder()
	handler.GetById(w, req)

	statusCode := w.Result().StatusCode
	if statusCode != http.StatusOK {
		t.Errorf("Expected status code to be %d, got %d", http.StatusOK, statusCode)
	}

	var note domain.Note
	if err := json.NewDecoder(w.Result().Body).Decode(&note); err != nil {
		t.Fatalf("Error while trying to decode: %s", err.Error())
	}

	if note.Id != expectedNote.Id || note.Title != expectedNote.Title || note.Content != expectedNote.Content {
		t.Errorf("Expected %+v, but got %+v", expectedNote, note)
	}
}

func TestUpdate(t *testing.T) {
	repo := getMockRepo()
	handler := NewNoteHandler(repo)

	id := repo.Notes[0].Id.String()
	body := `{ "title": "updated title" }`
	req := httptest.NewRequest("GET", "/note/"+id, strings.NewReader(body))
	req.SetPathValue("id", id)
	w := httptest.NewRecorder()
	handler.Update(w, req)

	statusCode := w.Result().StatusCode
	if statusCode != http.StatusOK {
		t.Errorf("Expected status code to be %d, got %d", http.StatusOK, statusCode)
	}

	if repo.Notes[0].Title != "updated title" {
		t.Errorf(`Expected updated note title to be "updated title", got %s`, repo.Notes[0].Title)
	}
}

func TestDelete(t *testing.T) {
	repo := getMockRepo()
	handler := NewNoteHandler(repo)

	initialNotesCount := len(repo.Notes)
	noteToRemove := repo.Notes[0]
	id := noteToRemove.Id.String()
	req := httptest.NewRequest("GET", "/note/"+id, nil)
	req.SetPathValue("id", id)
	w := httptest.NewRecorder()
	handler.Delete(w, req)

	statusCode := w.Result().StatusCode
	if statusCode != http.StatusNoContent {
		t.Errorf("Expected status code to be %d, got %d", http.StatusNoContent, statusCode)
	}

	notesCount := len(repo.Notes)
	if initialNotesCount-1 != notesCount {
		t.Errorf("Expected %d notes, got %d", initialNotesCount-1, notesCount)
	}

	result := slices.IndexFunc(repo.Notes, func(note domain.Note) bool {
		return note.Id == noteToRemove.Id
	})
	if result != -1 {
		t.Errorf("Expect note to be removed, but still exists")
	}
}
