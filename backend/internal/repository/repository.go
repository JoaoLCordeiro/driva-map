package repository

import (
	"database/sql"
	"fmt"

	"github.com/empresa/mercado/internal/models"
)

type Repository struct {
	db *sql.DB
}

func New(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ---- Users ----

func (r *Repository) GetUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

// ---- States ----

func (r *Repository) ListStates() ([]models.State, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.uf, s.nome, s.regiao, s.capital, s.populacao, s.pib_per_capita,
		       COUNT(b.id) AS filiais_ativas
		FROM states s
		LEFT JOIN branches b ON b.state_id = s.id AND b.deleted_at IS NULL
		GROUP BY s.id
		ORDER BY s.nome
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []models.State
	for rows.Next() {
		var s models.State
		var count int
		if err := rows.Scan(&s.ID, &s.UF, &s.Nome, &s.Regiao, &s.Capital,
			&s.Populacao, &s.PIBPerCapita, &count); err != nil {
			return nil, err
		}
		s.Filiais = &count
		states = append(states, s)
	}
	return states, rows.Err()
}

func (r *Repository) GetStateByUF(uf string) (*models.State, error) {
	s := &models.State{}
	var count int
	err := r.db.QueryRow(`
		SELECT s.id, s.uf, s.nome, s.regiao, s.capital, s.populacao, s.pib_per_capita,
		       COUNT(b.id) AS filiais_ativas
		FROM states s
		LEFT JOIN branches b ON b.state_id = s.id AND b.deleted_at IS NULL
		WHERE s.uf = $1
		GROUP BY s.id
	`, uf).Scan(&s.ID, &s.UF, &s.Nome, &s.Regiao, &s.Capital,
		&s.Populacao, &s.PIBPerCapita, &count)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Filiais = &count
	return s, nil
}

// ---- Branches ----

func (r *Repository) ListBranches(uf string, page, limit int) (*models.PaginatedBranches, error) {
	offset := (page - 1) * limit

	baseWhere := `WHERE b.deleted_at IS NULL`
	args := []any{}
	argN := 1

	if uf != "" {
		baseWhere += fmt.Sprintf(` AND s.uf = $%d`, argN)
		args = append(args, uf)
		argN++
	}

	var total int
	countQ := `SELECT COUNT(b.id) FROM branches b JOIN states s ON s.id = b.state_id ` + baseWhere
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, err
	}

	args = append(args, limit, offset)
	dataQ := fmt.Sprintf(`
		SELECT b.id, b.nome, b.cidade, b.state_id, s.uf, b.lat, b.lng,
		       TO_CHAR(b.opened_at, 'YYYY-MM-DD'), b.created_at
		FROM branches b
		JOIN states s ON s.id = b.state_id
		%s
		ORDER BY b.nome
		LIMIT $%d OFFSET $%d
	`, baseWhere, argN, argN+1)

	rows, err := r.db.Query(dataQ, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []models.Branch
	for rows.Next() {
		var b models.Branch
		if err := rows.Scan(&b.ID, &b.Nome, &b.Cidade, &b.StateID, &b.UF,
			&b.Lat, &b.Lng, &b.OpenedAt, &b.CreatedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	if branches == nil {
		branches = []models.Branch{}
	}
	return &models.PaginatedBranches{Data: branches, Page: page, Limit: limit, Total: total}, rows.Err()
}

func (r *Repository) GetBranchByID(id int) (*models.Branch, error) {
	b := &models.Branch{}
	err := r.db.QueryRow(`
		SELECT b.id, b.nome, b.cidade, b.state_id, s.uf, b.lat, b.lng,
		       TO_CHAR(b.opened_at, 'YYYY-MM-DD'), b.created_at
		FROM branches b JOIN states s ON s.id = b.state_id
		WHERE b.id = $1 AND b.deleted_at IS NULL
	`, id).Scan(&b.ID, &b.Nome, &b.Cidade, &b.StateID, &b.UF,
		&b.Lat, &b.Lng, &b.OpenedAt, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func (r *Repository) CreateBranch(req *models.BranchRequest) (*models.Branch, error) {
	var stateID int
	err := r.db.QueryRow(`SELECT id FROM states WHERE uf = $1`, req.UF).Scan(&stateID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("uf not found")
	}
	if err != nil {
		return nil, err
	}

	b := &models.Branch{}
	err = r.db.QueryRow(`
		INSERT INTO branches (nome, cidade, state_id, lat, lng, opened_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, nome, cidade, state_id, lat, lng,
		          TO_CHAR(opened_at, 'YYYY-MM-DD'), created_at
	`, req.Nome, req.Cidade, stateID, req.Lat, req.Lng, req.OpenedAt,
	).Scan(&b.ID, &b.Nome, &b.Cidade, &b.StateID, &b.Lat, &b.Lng, &b.OpenedAt, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	b.UF = req.UF
	return b, nil
}

func (r *Repository) UpdateBranch(id int, req *models.BranchRequest) (*models.Branch, error) {
	var stateID int
	err := r.db.QueryRow(`SELECT id FROM states WHERE uf = $1`, req.UF).Scan(&stateID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("uf not found")
	}
	if err != nil {
		return nil, err
	}

	b := &models.Branch{}
	err = r.db.QueryRow(`
		UPDATE branches SET nome=$1, cidade=$2, state_id=$3, lat=$4, lng=$5, opened_at=$6
		WHERE id=$7 AND deleted_at IS NULL
		RETURNING id, nome, cidade, state_id, lat, lng,
		          TO_CHAR(opened_at, 'YYYY-MM-DD'), created_at
	`, req.Nome, req.Cidade, stateID, req.Lat, req.Lng, req.OpenedAt, id,
	).Scan(&b.ID, &b.Nome, &b.Cidade, &b.StateID, &b.Lat, &b.Lng, &b.OpenedAt, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	b.UF = req.UF
	return b, nil
}

func (r *Repository) SoftDeleteBranch(id int) (bool, error) {
	// update avoids deleting a branch thats already deleted
	res, err := r.db.Exec(
		`UPDATE branches SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id,
	)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}