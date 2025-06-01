package repositories

import (
	"context"
	"time"

	"github.com/fliptv97/notepad-server/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type NoteRepository struct {
	conn *pgx.Conn
}

func NewNoteRepository(conn *pgx.Conn) *NoteRepository {
	return &NoteRepository{
		conn: conn,
	}
}

func (nr *NoteRepository) Create(ctx context.Context, title, content string) (*domain.Note, error) {
	note := &domain.Note{
		Id:        uuid.New(),
		Title:     title,
		Content:   content,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	_, err := nr.conn.Exec(
		ctx,
		"INSERT INTO note (id, title, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
		note.Id, note.Title, note.Content, note.CreatedAt, note.UpdatedAt,
	)

	return note, err
}

func (nr *NoteRepository) GetAll(ctx context.Context) ([]domain.Note, error) {
	rows, err := nr.conn.Query(ctx, "SELECT id, title, content, created_at, updated_at FROM note")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var note domain.Note
	notes := []domain.Note{}
	for rows.Next() {
		err := rows.Scan(&note.Id, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return notes, nil
}

func (nr *NoteRepository) GetById(ctx context.Context, id uuid.UUID) (*domain.Note, error) {
	var note domain.Note
	row := nr.conn.QueryRow(ctx, "SELECT id, title, content, created_at, updated_at FROM note WHERE id=$1", id)
	err := row.Scan(&note.Id, &note.Title, &note.Content, &note.CreatedAt, &note.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &note, err
}

func (nr *NoteRepository) Update(ctx context.Context, id uuid.UUID, title, content string) error {
	_, err := nr.conn.Exec(
		ctx,
		"UPDATE note SET title=$1, content=$2, updated_at=$3 WHERE id=$4",
		title,
		content,
		time.Now().UTC(),
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (nr *NoteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := nr.conn.Exec(ctx, "DELETE FROM note WHERE id=$1", id)

	return err
}
