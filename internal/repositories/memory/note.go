package memory

import (
	"context"
	"slices"
	"time"

	"github.com/fliptv97/notepad-server/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteRepository struct {
	Notes []domain.Note
}

func (nr *NoteRepository) Create(ctx context.Context, title, content string) (*domain.Note, error) {
	note := domain.Note{
		Id:        uuid.New(),
		Title:     title,
		Content:   content,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	nr.Notes = append(nr.Notes, note)

	return &note, nil
}

func (nr *NoteRepository) GetAll(ctx context.Context) ([]domain.Note, error) {
	return nr.Notes, nil
}

func (nr *NoteRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	note, err := nr.find(id)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (nr *NoteRepository) Update(ctx context.Context, id uuid.UUID, title, content string) error {
	note, err := nr.find(id)
	if err != nil {
		return err
	}

	note.Title = title
	note.Content = content
	note.UpdatedAt = time.Now().UTC()

	return nil
}

func (nr *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	idx := nr.findIndex(id)
	if idx == -1 {
		return pgx.ErrNoRows
	}

	nr.Notes = append(nr.Notes[:idx], nr.Notes[idx+1:]...)

	return nil
}

func (nr *NoteRepository) findIndex(id uuid.UUID) int {
	return slices.IndexFunc(nr.Notes, func(note domain.Note) bool {
		return note.Id == id
	})
}

func (nr *NoteRepository) find(id uuid.UUID) (*domain.Note, error) {
	idx := nr.findIndex(id)
	if idx == -1 {
		return nil, pgx.ErrNoRows
	}

	return &nr.Notes[idx], nil
}
