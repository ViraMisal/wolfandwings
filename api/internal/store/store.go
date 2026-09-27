package store

import (
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound — запись не найдена.
var ErrNotFound = errors.New("not found")

// Store — единый слой доступа к данным. Портативный SQL для sqlite/postgres.
type Store struct {
	DB      *sql.DB
	Dialect string
}

func New(db *sql.DB, dialect string) *Store {
	return &Store{DB: db, Dialect: dialect}
}

// ph возвращает плейсхолдер для N-го аргумента (? для sqlite, $N для postgres).
func (s *Store) ph(n int) string {
	if s.Dialect == "postgres" {
		return fmt.Sprintf("$%d", n)
	}
	return "?"
}

// phList возвращает "?, ?, ?" / "$1, $2, $3" (без внешних скобок).
func (s *Store) phList(start, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		if i > 0 {
			out += ", "
		}
		out += s.ph(start + i)
	}
	return out
}

// jsonParam — плейсхолдер для JSONB-колонки: Postgres не кастит text→jsonb неявно,
// поэтому параметр нужно кастить явно; sqlite хранит JSON как текст.
func (s *Store) jsonParam(n int) string {
	if s.Dialect == "postgres" {
		return s.ph(n) + "::jsonb"
	}
	return s.ph(n)
}
