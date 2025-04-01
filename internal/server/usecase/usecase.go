package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"io"
	"regexp"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresStorageInterface contains the necessary functions for the postgres storage logic of app.
type PostgresStorageInterface interface {
	CreateUser(
		ctx context.Context,
		login string,
		passwordHash, publicKey, privateKeyCipher []byte,
	) (*domain.User, error)
	AuthUser(
		ctx context.Context,
		login string,
		passwordHash []byte,
	) (*domain.User, error)
	CreateData(
		ctx context.Context,
		userID, name, dataType string,
		sizeInBytes uint64,
		data []byte,
	) (*domain.Data, error)
	GetData(
		ctx context.Context,
		userID, id string,
	) (*domain.Data, error)
	UpdateData(
		ctx context.Context,
		userID, id string,
		name, dataType string,
		sizeInBytes uint64,
		data []byte,
		status string,
		externalID *string,
	) (*domain.Data, error)
	GetDataBatch(
		ctx context.Context,
		userID string,
	) ([]*domain.Data, error)
	DeleteData(
		ctx context.Context,
		userID, id string,
	) (*domain.Data, error)
}

// S3StorageInterface contains the necessary functions for the S3 storage logic of app.
type S3StorageInterface interface {
	GetObject(ctx context.Context, objectName string) (io.Reader, error)
	PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64) error
	RemoveObject(ctx context.Context, objectName string) error
}

// Usecase is the business logic of app.
type Usecase struct {
	postgresStorage PostgresStorageInterface
	s3Storage       S3StorageInterface

	passwordKey       string
	minLoginLength    uint
	minPasswordLength uint
}

// NewUsecase creates new Usecase.
func NewUsecase(
	postgresStorage PostgresStorageInterface,
	s3Storage S3StorageInterface,
	passwordKey string,
	minLoginLength uint,
	minPasswordLength uint,
) *Usecase {
	return &Usecase{
		postgresStorage: postgresStorage,
		s3Storage:       s3Storage,

		passwordKey:       passwordKey,
		minLoginLength:    minLoginLength,
		minPasswordLength: minPasswordLength,
	}
}

func (u *Usecase) hashPassword(password string) []byte {
	// подписываем алгоритмом HMAC, используя SHA-256
	h := hmac.New(sha256.New, []byte(u.passwordKey))
	h.Write([]byte(password))
	passwordHash := h.Sum(nil)

	return passwordHash
}

func (u *Usecase) checkLogin(login string) (bool, error) {
	if uint(len(login)) < u.minLoginLength {
		return false, nil
	}
	okInvalidSymbols, err := regexp.MatchString(`[^\w\.\-]+`, login)
	return !okInvalidSymbols, err
}

func (u *Usecase) checkPassword(password string) (bool, error) {
	if uint(len(password)) < u.minPasswordLength {
		return false, nil
	}
	okInvalidSymbols, err := regexp.MatchString(`[^\w\.\-]+`, password)
	return !okInvalidSymbols, err
}

// Register creates new user.
func (u *Usecase) Register(ctx context.Context, login, password string, publicKey, privateKeyCipher []byte) (*domain.User, error) {
	ok, err := u.checkLogin(login)
	if !ok || err != nil {
		return nil, domain.ErrInvalidLoginPasswordFormat
	}

	ok, err = u.checkPassword(password)
	if !ok || err != nil {
		return nil, domain.ErrInvalidLoginPasswordFormat
	}

	passwordHash := u.hashPassword(password)
	user, err := u.postgresStorage.CreateUser(ctx, login, passwordHash, publicKey, privateKeyCipher)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == pgerrcode.UniqueViolation && pgErr.Message == "duplicate key value violates unique constraint \"user_login_key\"":
			return nil, domain.ErrLoginTaken
		case pgErr.Code == pgerrcode.CheckViolation && pgErr.Message == "new row for relation \"user\" violates check constraint \"user_login_check\"":
			return nil, domain.ErrInvalidLoginPasswordFormat
		default:
			return nil, err
		}
	}

	return user, err
}

// Login auths user.
func (u *Usecase) Login(ctx context.Context, login, password string) (*domain.User, error) {
	ok, err := u.checkLogin(login)
	if !ok || err != nil {
		return nil, domain.ErrInvalidLoginPasswordFormat
	}

	ok, err = u.checkPassword(password)
	if !ok || err != nil {
		return nil, domain.ErrInvalidLoginPasswordFormat
	}

	passwordHash := u.hashPassword(password)
	user, err := u.postgresStorage.AuthUser(ctx, login, passwordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvalidLoginPassword
	}

	return user, err
}

// CreateData creates data.
func (u *Usecase) CreateData(ctx context.Context, userID string, name string, dataType string, sizeInBytes uint64, data []byte) (*domain.Data, error) {
	d, err := u.postgresStorage.CreateData(ctx, userID, name, dataType, sizeInBytes, data)
	if err != nil {
		return nil, err
	}

	return d, nil
}

// CreateChunkedData creates chunked data
func (u *Usecase) CreateChunkedData(ctx context.Context, userID string, id string, dataChunkBuffer *domain.DataChunkBuffer) error {
	data, err := u.postgresStorage.GetData(ctx, userID, id)
	if err != nil {
		return err
	}

	_, err = u.postgresStorage.UpdateData(
		ctx,
		userID,
		id,
		data.Name,
		data.Type,
		data.SizeInBytes,
		data.Data,
		string(domain.StatusUploading),
		nil,
	)
	if err != nil {
		return err
	}

	err = u.s3Storage.PutObject(ctx, id, dataChunkBuffer, int64(data.SizeInBytes))
	if err != nil {
		return err
	}

	_, err = u.postgresStorage.UpdateData(
		ctx,
		userID,
		id,
		data.Name,
		data.Type,
		data.SizeInBytes,
		data.Data,
		string(domain.StatusUploading),
		&id,
	)

	if err != nil {
		return err
	}

	_, err = u.postgresStorage.UpdateData(
		ctx,
		userID,
		id,
		data.Name,
		data.Type,
		data.SizeInBytes,
		data.Data,
		string(domain.StatusUploaded),
		&id,
	)
	if err != nil {
		return err
	}

	return nil
}

// GetDataBatch gets data batch.
func (u *Usecase) GetDataBatch(ctx context.Context, userID string) ([]*domain.Data, error) {
	return u.postgresStorage.GetDataBatch(ctx, userID)
}

// GetDataChunkReader gets chunked data.
func (u *Usecase) GetDataChunkReader(ctx context.Context, userID string, id string) (io.Reader, error) {
	d, err := u.postgresStorage.GetData(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return u.s3Storage.GetObject(ctx, d.ID)
}

// UpdateData updates data.
func (u *Usecase) UpdateData(ctx context.Context, userID string, id string, name string, dataType string, sizeInBytes uint64, data []byte) (*domain.Data, error) {
	d, err := u.postgresStorage.GetData(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return u.postgresStorage.UpdateData(ctx, userID, id, name, dataType, sizeInBytes, data, string(d.Status), d.ExternalID)
}

// DeleteData deletes data.
func (u *Usecase) DeleteData(ctx context.Context, userID string, id string) (*domain.Data, error) {
	d, err := u.postgresStorage.DeleteData(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if d.ExternalID != nil {
		err = u.s3Storage.RemoveObject(ctx, *d.ExternalID)
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}
