package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
)

const DataSizeLimit = 15 * 1024 * 1024

var (
	ErrDataTooLarge = errors.New("data too large")
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (ps *PostgresStorage) CreateUser(ctx context.Context, login string, passwordHash, publicKey, privateKeyCipher []byte) (*domain.User, error) {
	query := `INSERT INTO "user" (login, password_hash, public_key, private_key_cipher) VALUES ($1, $2, $3, $4) RETURNING id;`

	row := ps.db.QueryRowContext(
		ctx,
		query,
		login,
		passwordHash,
		publicKey,
		privateKeyCipher,
	)

	var id uint
	err := row.Scan(&id)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:               strconv.Itoa(int(id)),
		Login:            login,
		PasswordHash:     passwordHash,
		PublicKey:        publicKey,
		PrivateKeyCipher: privateKeyCipher,
	}, nil
}

func (ps *PostgresStorage) AuthUser(ctx context.Context, login string, passwordHash []byte) (*domain.User, error) {
	query := `SELECT id, login, password_hash, public_key, private_key_cipher FROM "user" WHERE login = $1 AND password_hash = $2;`

	row := ps.db.QueryRowContext(
		ctx,
		query,
		login,
		passwordHash,
	)

	var id uint
	user := &domain.User{}
	err := row.Scan(&id, &user.Login, &user.PasswordHash, &user.PublicKey, &user.PrivateKeyCipher)
	if err != nil {
		return nil, err
	}

	user.ID = strconv.Itoa(int(id))

	return user, nil
}

func (ps *PostgresStorage) CreateData(ctx context.Context, userID, name, dataType string, data []byte) (*domain.Data, error) {
	if len(data) > DataSizeLimit {
		return nil, ErrDataTooLarge
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO data (user_id, name, type, data) 
VALUES ($1, $2, $3, $4) 
ON CONFLICT (user_id, name, type) 
DO UPDATE SET
	data = EXCLUDED.data,
	updated_at = NOW()
RETURNING id, created_at, updated_at;`

	row := ps.db.QueryRowContext(
		ctx,
		query,
		userIDUint,
		name,
		dataType,
		data,
	)

	var id uint
	d := &domain.Data{
		UserID: userID,
		Name:   name,
		Type:   dataType,
		Data:   data,
	}

	err = row.Scan(&id, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}

	d.ID = strconv.Itoa(int(id))

	return d, nil
}

func (ps *PostgresStorage) GetData(ctx context.Context, userID, id string) (*domain.Data, error) {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	query := `SELECT name, type, data, created_at, updated_at FROM data WHERE user_id = $1 AND id = $2;`

	row := ps.db.QueryRowContext(
		ctx,
		query,
		userIDUint,
		idUint,
	)

	d := &domain.Data{
		ID:     id,
		UserID: userID,
	}

	err = row.Scan(
		&d.Name,
		&d.Type,
		&d.Data,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (ps *PostgresStorage) GetDataBatch(ctx context.Context, userID string) ([]*domain.Data, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	query := `SELECT id, name, type, data, created_at, updated_at FROM data WHERE user_id = $1 ORDER BY updated_at DESC;`

	rows, err := ps.db.QueryContext(
		ctx,
		query,
		userIDUint,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dataBatch := []*domain.Data{}
	for rows.Next() {
		var (
			id        uint
			name      string
			dataType  string
			data      []byte
			createdAt time.Time
			updatedAt time.Time
		)

		err = rows.Scan(
			&id,
			&name,
			&dataType,
			&data,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		dataBatch = append(dataBatch, &domain.Data{
			ID:        strconv.Itoa(int(id)),
			UserID:    userID,
			Name:      name,
			Type:      dataType,
			Data:      data,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dataBatch, nil
}

func (ps *PostgresStorage) UpdateData(ctx context.Context, userID, id string, name, dataType string, data []byte) (*domain.Data, error) {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	query := `UPDATE 
    data 
SET 
    name = $1, 
    type = $2, 
    data = $3, 
    updated_at = now(),
WHERE 
    user_id = $4 AND 
    id = $5 
RETURNING 
	created_at, 
	updated_at;`

	row := ps.db.QueryRowContext(
		ctx,
		query,
		name,
		dataType,
		data,
		userIDUint,
		idUint,
	)

	d := &domain.Data{
		ID:     id,
		UserID: userID,
		Name:   name,
		Type:   dataType,
		Data:   data,
	}

	err = row.Scan(
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (ps *PostgresStorage) DeleteData(ctx context.Context, userID, id string) (*domain.Data, error) {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	query := `DELETE FROM data WHERE user_id = $1 AND id = $2 RETURNING name, type, data, created_at, updated_at;`

	row := ps.db.QueryRowContext(
		ctx,
		query,
		userIDUint,
		idUint,
	)

	d := &domain.Data{
		ID:     id,
		UserID: userID,
	}

	err = row.Scan(
		&d.Name,
		&d.Type,
		&d.Data,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}
