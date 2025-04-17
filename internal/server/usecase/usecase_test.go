package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/domain"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/server/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func TestNewUsecase(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	mockPostgresStorage := mocks.NewMockPostgresStorageInterface(ctrl)
	mockMinioStorage := mocks.NewMockMinioStorageInterface(ctrl)

	passwordKey := "12345"
	minLoginLength := 6
	minPasswordLength := 8

	usecase := &Usecase{
		postgresStorage:   mockPostgresStorage,
		minioStorage:      mockMinioStorage,
		passwordKey:       passwordKey,
		minLoginLength:    minLoginLength,
		minPasswordLength: minPasswordLength,
	}

	actualUsecase := NewUsecase(
		mockPostgresStorage,
		mockMinioStorage,
		passwordKey,
		minLoginLength,
		minPasswordLength,
	)

	assert.Equal(t, usecase, actualUsecase)
}

func TestUsecase_Register(t *testing.T) {
	login := "login"
	password := "password"
	invalidLogin := "invalid_login?"
	existedLogin := "existed_login"
	invalidPassword := "invalid_password?"
	shortLogin := "1"
	shortPassword := "1"

	passwordHash := []byte("password_hash")
	publicKey := []byte("public_key")
	privateKeyCipher := []byte("private_key_cipher")

	user := &domain.User{
		ID:               "1",
		Login:            login,
		PasswordHash:     passwordHash,
		PublicKey:        publicKey,
		PrivateKeyCipher: privateKeyCipher,
	}

	type args struct {
		login    string
		password string
	}

	type want struct {
		user *domain.User
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				login:    login,
				password: password,
			},
			want: want{
				user: user,
				err:  nil,
			},
		},
		{
			name: "invalid login",
			args: args{
				login:    invalidLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPasswordFormat,
			},
		},
		{
			name: "login taken",
			args: args{
				login:    existedLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  domain.ErrLoginTaken,
			},
		},
		{
			name: "invalid password",
			args: args{
				login:    login,
				password: invalidPassword,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPasswordFormat,
			},
		},
		{
			name: "short login",
			args: args{
				login:    shortLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPasswordFormat,
			},
		},
		{
			name: "short password",
			args: args{
				login:    login,
				password: shortPassword,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPasswordFormat,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockPostgresStorageInterface(ctrl)

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().CreateUser(
		gomock.Any(),
		login,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(user, nil).AnyTimes()

	m.EXPECT().CreateUser(
		gomock.Any(),
		existedLogin,
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(
		nil,
		&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "duplicate key value violates unique constraint \"user_login_key\""},
	).AnyTimes()

	passwordKey := ""
	minLoginLength := 3
	minPasswordLength := 3

	usecase := &Usecase{
		postgresStorage:   m,
		passwordKey:       passwordKey,
		minLoginLength:    minLoginLength,
		minPasswordLength: minPasswordLength,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := usecase.Register(
				context.Background(),
				tt.args.login,
				tt.args.password,
				publicKey,
				privateKeyCipher,
			)

			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.user, u)
			}
		})
	}
}

func TestUsecase_Login(t *testing.T) {
	login := "login"
	password := "password"
	invalidLogin := "invalid_login?"
	invalidPassword := "invalid_password?"
	incorrectPassword := "incorrect_password"

	passwordHash := []byte("password_hash")
	publicKey := []byte("public_key")
	privateKeyCipher := []byte("private_key_cipher")

	user := &domain.User{
		ID:               "1",
		Login:            login,
		PasswordHash:     passwordHash,
		PublicKey:        publicKey,
		PrivateKeyCipher: privateKeyCipher,
	}

	type args struct {
		login    string
		password string
	}

	type want struct {
		user *domain.User
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "valid data",
			args: args{
				login:    login,
				password: password,
			},
			want: want{
				user: user,
				err:  nil,
			},
		},
		{
			name: "invalid login",
			args: args{
				login:    invalidLogin,
				password: password,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPasswordFormat,
			},
		},
		{
			name: "incorrect password",
			args: args{
				login:    login,
				password: incorrectPassword,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPassword,
			},
		},
		{
			name: "invalid password",
			args: args{
				login:    login,
				password: invalidPassword,
			},
			want: want{
				user: nil,
				err:  domain.ErrInvalidLoginPasswordFormat,
			},
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockPostgresStorageInterface(ctrl)

	passwordKey := ""
	minLoginLength := 3
	minPasswordLength := 3

	usecase := &Usecase{
		postgresStorage:   m,
		passwordKey:       passwordKey,
		minLoginLength:    minLoginLength,
		minPasswordLength: minPasswordLength,
	}

	// гарантируем, что заглушка
	// при вызове с аргументом "Key" вернёт "Value"
	m.EXPECT().AuthUser(gomock.Any(), login, usecase.hashPassword(password)).Return(user, nil).AnyTimes()

	m.EXPECT().AuthUser(gomock.Any(), login, gomock.Not(usecase.hashPassword(password))).Return(nil, sql.ErrNoRows).AnyTimes()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := usecase.Login(context.Background(), tt.args.login, tt.args.password)

			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.user, u)
			}
		})
	}
}

func TestUsecase_CreateData(t *testing.T) {
	userID := "1"
	name := "data"
	existedNameWithTypeForUser := "bad_name"
	dataType := pb.Type_LOGIN_PASSWORD.String()
	dataBytes := []byte("data")
	now := time.Now()
	internalError := errors.New("internal error")

	data := &domain.Data{
		ID:        "1",
		UserID:    userID,
		Name:      name,
		Type:      dataType,
		Data:      dataBytes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		postgresStorageMock func() *mocks.MockPostgresStorageInterface
		userID              string
		name                string
		dataType            string
		dataBytes           []byte
	}

	type want struct {
		data *domain.Data
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().CreateData(
						gomock.Any(),
						userID,
						name,
						dataType,
						dataBytes,
					).Return(
						data,
						nil,
					)

					return m
				},
				userID:    userID,
				name:      name,
				dataType:  dataType,
				dataBytes: dataBytes,
			},
			want: want{
				data: data,
				err:  nil,
			},
		},
		{
			name: "existed name and type for user",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().CreateData(
						gomock.Any(),
						userID,
						existedNameWithTypeForUser,
						dataType,
						dataBytes,
					).Return(
						nil,
						sql.ErrNoRows,
					)

					return m
				},
				userID:    userID,
				name:      existedNameWithTypeForUser,
				dataType:  dataType,
				dataBytes: dataBytes,
			},
			want: want{
				data: nil,
				err:  domain.ErrDataWithThisNameAndTypeExists,
			},
		},
		{
			name: "internal error",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().CreateData(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Return(
						nil,
						internalError,
					)

					return m
				},
				userID:    userID,
				name:      name,
				dataType:  dataType,
				dataBytes: dataBytes,
			},
			want: want{
				data: nil,
				err:  internalError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &Usecase{
				postgresStorage:   tt.args.postgresStorageMock(),
				passwordKey:       "",
				minLoginLength:    0,
				minPasswordLength: 0,
			}

			data, err := usecase.CreateData(
				context.Background(),
				tt.args.userID,
				tt.args.name,
				tt.args.dataType,
				tt.args.dataBytes,
			)

			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.data, data)
			}
		})
	}
}

func TestUsecase_GetDataBatch(t *testing.T) {
	userID := "1"
	id := "1"
	name := "data"
	dataType := pb.Type_LOGIN_PASSWORD.String()
	dataBytes := []byte("data")
	now := time.Now()
	internalError := errors.New("internal error")

	dataBatch := []*domain.Data{
		{
			ID:        id,
			UserID:    userID,
			Name:      name,
			Type:      dataType,
			Data:      dataBytes,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		postgresStorageMock func() *mocks.MockPostgresStorageInterface
		id                  string
	}

	type want struct {
		dataBatch []*domain.Data
		err       error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().GetDataBatch(
						gomock.Any(),
						userID,
					).Return(
						dataBatch,
						nil,
					)

					return m
				},
				id: id,
			},
			want: want{
				dataBatch: dataBatch,
				err:       nil,
			},
		},
		{
			name: "internal error",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().GetDataBatch(
						gomock.Any(),
						gomock.Any(),
					).Return(
						nil,
						internalError,
					)

					return m
				},
				id: id,
			},
			want: want{
				dataBatch: nil,
				err:       internalError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &Usecase{
				postgresStorage:   tt.args.postgresStorageMock(),
				passwordKey:       "",
				minLoginLength:    0,
				minPasswordLength: 0,
			}

			dataBatch, err := usecase.GetDataBatch(
				context.Background(),
				tt.args.id,
			)

			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.dataBatch, dataBatch)
			}
		})
	}
}

func TestUsecase_UpdateData(t *testing.T) {
	userID := "1"
	id := "1"
	name := "data"
	dataType := pb.Type_LOGIN_PASSWORD.String()
	dataBytes := []byte("data")
	now := time.Now()
	internalError := errors.New("internal error")

	data := &domain.Data{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Type:      dataType,
		Data:      dataBytes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		postgresStorageMock func() *mocks.MockPostgresStorageInterface
		id                  string
		userID              string
		name                string
		dataType            string
		dataBytes           []byte
	}

	type want struct {
		data *domain.Data
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().UpdateData(
						gomock.Any(),
						userID,
						id,
						name,
						dataType,
						dataBytes,
					).Return(
						data,
						nil,
					)

					return m
				},
				id:        id,
				userID:    userID,
				name:      name,
				dataType:  dataType,
				dataBytes: dataBytes,
			},
			want: want{
				data: data,
				err:  nil,
			},
		},
		{
			name: "internal error",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().UpdateData(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Return(
						nil,
						internalError,
					)

					return m
				},
				id:        id,
				userID:    userID,
				name:      name,
				dataType:  dataType,
				dataBytes: dataBytes,
			},
			want: want{
				data: nil,
				err:  internalError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &Usecase{
				postgresStorage:   tt.args.postgresStorageMock(),
				passwordKey:       "",
				minLoginLength:    0,
				minPasswordLength: 0,
			}

			data, err := usecase.UpdateData(
				context.Background(),
				tt.args.userID,
				tt.args.id,
				tt.args.name,
				tt.args.dataType,
				tt.args.dataBytes,
			)

			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.data, data)
			}
		})
	}
}

func TestUsecase_DeleteData(t *testing.T) {
	userID := "1"
	id := "1"
	name := "data"
	dataType := pb.Type_LOGIN_PASSWORD.String()
	dataBytes := []byte("data")
	now := time.Now()
	internalError := errors.New("internal error")

	data := &domain.Data{
		ID:        id,
		UserID:    userID,
		Name:      name,
		Type:      dataType,
		Data:      dataBytes,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		postgresStorageMock func() *mocks.MockPostgresStorageInterface
		id                  string
		userID              string
	}

	type want struct {
		data *domain.Data
		err  error
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ok",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().DeleteData(
						gomock.Any(),
						userID,
						id,
					).Return(
						data,
						nil,
					)

					return m
				},
				id:     id,
				userID: userID,
			},
			want: want{
				data: data,
				err:  nil,
			},
		},
		{
			name: "internal error",
			args: args{
				postgresStorageMock: func() *mocks.MockPostgresStorageInterface {
					m := mocks.NewMockPostgresStorageInterface(ctrl)

					m.EXPECT().DeleteData(
						gomock.Any(),
						gomock.Any(),
						gomock.Any(),
					).Return(
						nil,
						internalError,
					)

					return m
				},
				id:     id,
				userID: userID,
			},
			want: want{
				data: nil,
				err:  internalError,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			usecase := &Usecase{
				postgresStorage:   tt.args.postgresStorageMock(),
				passwordKey:       "",
				minLoginLength:    0,
				minPasswordLength: 0,
			}

			data, err := usecase.DeleteData(
				context.Background(),
				tt.args.userID,
				tt.args.id,
			)

			assert.ErrorIs(t, err, tt.want.err)
			if err == nil {
				assert.Equal(t, tt.want.data, data)
			}
		})
	}
}
