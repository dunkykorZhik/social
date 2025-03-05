package storage

import (
	"context"
	"database/sql"
)

type Role struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RoleStorage struct {
	db *sql.DB
}

func (r *RoleStorage) GetByName(ctx context.Context, name string) (*Role, error) {
	query := `SELECT id, name, description FROM roles WHERE name = $1`

	role := &Role{}
	err := r.db.QueryRowContext(ctx, query, name).Scan(&role.ID, &role.Name, &role.Description)
	if err != nil {
		return nil, err
	}

	return role, nil
}
