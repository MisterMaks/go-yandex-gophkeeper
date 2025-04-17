package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"errors"
	"io"
	"regexp"
	"time"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/minio/minio-go/v7"
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
		data []byte,
	) (*domain.Data, error)
	GetDataBatch(
		ctx context.Context,
		userID string,
	) ([]*domain.Data, error)
	GetData(
		ctx context.Context,
		userID, id string,
	) (*domain.Data, error)
	UpdateData(
		ctx context.Context,
		userID, id, name, dataType string,
		data []byte,
	) (*domain.Data, error)
	DeleteData(
		ctx context.Context,
		userID, id string,
	) (*domain.Data, error)

	CreateChunkedData(
		ctx context.Context,
		userID, name, dataType string,
		externalID string,
	) (*domain.Data, error)
	SetUploaded(
		ctx context.Context,
		userID, id string,
		taskID int64,
	) (time.Time, error)
	SetDeleting(
		ctx context.Context,
		userID, id string,
	) (*domain.Data, error)
	DeleteChunkedData(
		ctx context.Context,
		userID, id string,
		taskID int64,
	) (*domain.Data, error)
}

// MinioStorageInterface contains the necessary functions for the S3 storage logic of app.
type MinioStorageInterface interface {
	GetObject(ctx context.Context, objectName string) (*minio.Object, error)
	PutObject(ctx context.Context, objectName string, reader io.Reader, objectSize int64) (minio.UploadInfo, error)
	RemoveObject(ctx context.Context, objectName string) error
}

// Usecase is the business logic of app.
type Usecase struct {
	postgresStorage PostgresStorageInterface
	minioStorage    MinioStorageInterface

	passwordKey       string
	minLoginLength    int
	minPasswordLength int
}

// NewUsecase creates new Usecase.
func NewUsecase(
	postgresStorage PostgresStorageInterface,
	minioStorage MinioStorageInterface,
	passwordKey string,
	minLoginLength int,
	minPasswordLength int,
) *Usecase {
	return &Usecase{
		postgresStorage: postgresStorage,
		minioStorage:    minioStorage,

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
	if len(login) < u.minLoginLength {
		return false, nil
	}
	okInvalidSymbols, err := regexp.MatchString(`[^\w\.\-]+`, login)
	return !okInvalidSymbols, err
}

func (u *Usecase) checkPassword(password string) (bool, error) {
	if len(password) < u.minPasswordLength {
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
		case pgErr.Code == pgerrcode.UniqueViolation:
			return nil, domain.ErrLoginTaken
		case pgErr.Code == pgerrcode.CheckViolation:
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
func (u *Usecase) CreateData(ctx context.Context, userID string, name string, dataType string, data []byte) (*domain.Data, error) {
	d, err := u.postgresStorage.CreateData(ctx, userID, name, dataType, data)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrDataWithThisNameAndTypeExists
		}
		return nil, err
	}

	return d, nil
}

// GetDataBatch gets data batch.
func (u *Usecase) GetDataBatch(ctx context.Context, userID string) ([]*domain.Data, error) {
	return u.postgresStorage.GetDataBatch(ctx, userID)
}

// UpdateData updates data.
func (u *Usecase) UpdateData(ctx context.Context, userID string, id string, name string, dataType string, data []byte) (*domain.Data, error) {
	return u.postgresStorage.UpdateData(ctx, userID, id, name, dataType, data)
}

// DeleteData deletes data.
func (u *Usecase) DeleteData(ctx context.Context, userID string, id string) (*domain.Data, error) {
	d, err := u.postgresStorage.DeleteData(ctx, userID, id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == pgerrcode.NoData:
			d, err = u.postgresStorage.SetDeleting(ctx, userID, id)
			if err != nil {
				return nil, err
			}

			err = u.minioStorage.RemoveObject(ctx, *d.ExternalID)
			if err != nil {
				return nil, err
			}

			d, err = u.postgresStorage.DeleteChunkedData(ctx, userID, id, *d.TaskID)
			if err != nil {
				return nil, err
			}
		default:
			return nil, err
		}
	}

	return d, nil
}

// CreateChunkedData creates chunked data
func (u *Usecase) CreateChunkedData(ctx context.Context, userID string, chunkedData *domain.ChunkedData) (*domain.Data, error) {
	data, err := u.postgresStorage.CreateChunkedData(ctx, userID, chunkedData.Name, chunkedData.Type, userID+"_"+chunkedData.Type+"_"+chunkedData.Name)
	if err != nil {
		return nil, err
	}

	_, err = u.minioStorage.PutObject(ctx, data.ID, chunkedData, chunkedData.SizeInBytes)
	if err != nil {
		return nil, err
	}

	updatedAt, err := u.postgresStorage.SetUploaded(
		ctx,
		userID,
		data.ID,
		*data.TaskID,
	)

	if err != nil {
		return nil, err
	}

	data.UpdatedAt = updatedAt

	return data, nil
}

// GetChunkedDataReader gets chunked data reader.
func (u *Usecase) GetChunkedDataReader(ctx context.Context, userID string, id string) (io.Reader, error) {
	d, err := u.postgresStorage.GetData(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	return u.minioStorage.GetObject(ctx, *d.ExternalID)
}
