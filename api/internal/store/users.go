package store

import (
	"context"
	"database/sql"

	"github.com/wolfandwings/api/internal/model"
)

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := s.DB.QueryRowContext(ctx, `SELECT id, email, password_hash, name, role
		FROM users WHERE email=`+s.ph(1), email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	err := s.DB.QueryRowContext(ctx, `SELECT id, email, password_hash, name, role
		FROM users WHERE id=`+s.ph(1), id).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role)
	if err != nil {
		return nil, ErrNotFound
	}
	return &u, nil
}
