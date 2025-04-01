package db

import (
	"context"
	"database/sql"
	"errors"
	"strconv"
	"time"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
)

const (
	DataSizeLimit = 15 * 1024 * 1024

	CreateUserQuery = `INSERT INTO "user" (login, password_hash, public_key, private_key_cipher) VALUES ($1, $2, $3, $4) RETURNING id;`
	GetUserQuery    = `SELECT id, login, password_hash, public_key, private_key_cipher FROM "user" WHERE login = $1 AND password_hash = $2;`

	CreateDataQuery = `INSERT INTO data (user_id, name, type, size_in_bytes, data) VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at, updated_at;`
	GetDataQuery    = `SELECT name, type, size_in_bytes, status, data, external_id, created_at, updated_at FROM data WHERE user_id = $1 AND id = $2;`

	// UpdateDataQuery update data. But cannot update `size_in_bytes` and `data` fields in "uploading" status
	UpdateDataQuery = `UPDATE 
    data 
SET 
    name = $1, 
    type = $2, 
    size_in_bytes = $3, 
    status = $4, 
    data = $5, 
    external_id = $6,
    updated_at = now(),
WHERE 
    user_id = $7 AND 
    id = $8 AND 
    (
        (
            size_in_bytes = $9 AND 
            data = $10
		) OR 
        status = 'uploaded'
	)
RETURNING 
	created_at, 
	updated_at;`

	GetDataBatchQuery = `SELECT id, name, type, size_in_bytes, status, data, external_id, created_at, updated_at FROM data WHERE user_id = $1;`
	DeleteDataQuery   = `DELETE FROM data WHERE user_id = $1 AND id = $2 RETURNING name, type, size_in_bytes, status, data, external_id, created_at, updated_at;`
)

var (
	ErrDataTooLarge = errors.New("data too large")
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	return &PostgresStorage{db: db}, nil
}

func (ps *PostgresStorage) CreateUser(ctx context.Context, login string, passwordHash, publicKey, privateKeyCipher []byte) (*domain.User, error) {
	row := ps.db.QueryRowContext(
		ctx,
		CreateUserQuery,
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
	row := ps.db.QueryRowContext(
		ctx,
		GetUserQuery,
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

func (ps *PostgresStorage) CreateData(ctx context.Context, userID, name, dataType string, sizeInBytes uint64, data []byte) (*domain.Data, error) {
	if len(data) > DataSizeLimit {
		return nil, ErrDataTooLarge
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	row := ps.db.QueryRowContext(
		ctx,
		CreateDataQuery,
		userIDUint,
		name,
		dataType,
		sizeInBytes,
		data,
	)

	var id uint
	d := &domain.Data{
		UserID:      userID,
		Name:        name,
		Type:        dataType,
		SizeInBytes: sizeInBytes,
		Status:      domain.StatusUploaded,
		Data:        data,
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

	row := ps.db.QueryRowContext(
		ctx,
		GetDataQuery,
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
		&d.SizeInBytes,
		&d.Status,
		&d.Data,
		&d.ExternalID,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (ps *PostgresStorage) UpdateData(ctx context.Context, userID, id string, name, dataType string, sizeInBytes uint64, data []byte, status string, externalID *string) (*domain.Data, error) {
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, err
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	row := ps.db.QueryRowContext(
		ctx,
		UpdateDataQuery,
		name,
		dataType,
		sizeInBytes,
		status,
		data,
		externalID,
		userIDUint,
		idUint,
		sizeInBytes,
		data,
	)

	d := &domain.Data{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Type:        dataType,
		SizeInBytes: sizeInBytes,
		Status:      domain.Status(status),
		Data:        data,
		ExternalID:  externalID,
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

func (ps *PostgresStorage) GetDataBatch(ctx context.Context, userID string) ([]*domain.Data, error) {
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, err
	}

	rows, err := ps.db.QueryContext(
		ctx,
		GetDataBatchQuery,
		userIDUint,
	)
	if err != nil {
		return nil, err
	}
	rows.Close()

	dataBatch := []*domain.Data{}
	for rows.Next() {
		var (
			id          uint
			name        string
			dataType    string
			sizeInBytes uint64
			status      string
			data        []byte
			externalID  *string
			createdAt   time.Time
			updatedAt   time.Time
		)

		err = rows.Scan(
			&id,
			&name,
			&dataType,
			&sizeInBytes,
			&status,
			&data,
			&externalID,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		dataBatch = append(dataBatch, &domain.Data{
			ID:          strconv.Itoa(int(id)),
			UserID:      userID,
			Name:        name,
			Type:        dataType,
			SizeInBytes: sizeInBytes,
			Status:      domain.Status(status),
			Data:        data,
			ExternalID:  externalID,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return dataBatch, nil
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

	row := ps.db.QueryRowContext(
		ctx,
		DeleteDataQuery,
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
		&d.SizeInBytes,
		&d.Status,
		&d.Data,
		&d.ExternalID,
		&d.CreatedAt,
		&d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return d, nil
}
