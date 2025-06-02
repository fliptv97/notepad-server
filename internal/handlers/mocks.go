package handlers

import (
	"context"
	"slices"
	"time"

	"github.com/fliptv97/notepad-server/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteRepositoryMock struct {
	Notes []domain.Note
}

func (nrm *NoteRepositoryMock) Create(ctx context.Context, title, content string) (*domain.Note, error) {
	note := domain.Note{
		Id:        uuid.New(),
		Title:     title,
		Content:   content,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	nrm.Notes = append(nrm.Notes, note)

	return &note, nil
}

func (nrm *NoteRepositoryMock) GetAll(ctx context.Context) ([]domain.Note, error) {
	return nrm.Notes, nil
}

func (nrm *NoteRepositoryMock) GetById(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	note, err := nrm.find(id)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (nrm *NoteRepositoryMock) Update(ctx context.Context, id uuid.UUID, title, content string) error {
	note, err := nrm.find(id)
	if err != nil {
		return err
	}

	note.Title = title
	note.Content = content
	note.UpdatedAt = time.Now().UTC()

	return nil
}

func (nrm *NoteRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	idx := nrm.findIndex(id)
	if idx == -1 {
		return pgx.ErrNoRows
	}

	nrm.Notes = append(nrm.Notes[:idx], nrm.Notes[idx+1:]...)

	return nil
}

func (nrm *NoteRepositoryMock) findIndex(id uuid.UUID) int {
	return slices.IndexFunc(nrm.Notes, func(note domain.Note) bool {
		return note.Id == id
	})
}

func (nrm *NoteRepositoryMock) find(id uuid.UUID) (*domain.Note, error) {
	idx := nrm.findIndex(id)
	if idx == -1 {
		return nil, pgx.ErrNoRows
	}

	return &nrm.Notes[idx], nil
}
