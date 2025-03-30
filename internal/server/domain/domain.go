package domain

import (
	"errors"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
)

var (
	ErrLoginTaken                 = errors.New("login already taken")
	ErrInvalidLoginPassword       = errors.New("invalid login/password")
	ErrInvalidLoginPasswordFormat = errors.New("invalid login/password format")

	ErrOrderUploaded              = errors.New("order number has already been uploaded by this user")
	ErrOrderUploadedByAnotherUser = errors.New("order number has already been uploaded by another user")

	ErrInsufficientFunds  = errors.New("there are insufficient funds in the account")
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

type User struct {
	ID               string
	Login            string
	PasswordHash     []byte
	PublicKey        []byte
	PrivateKeyCipher []byte
}

type Data struct {
	ID        string
	UserID    string
	Name      string
	Type      string
	Data      []byte
	IsChunked bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d Data) SerializeToProtobuf() *pb.GetDataBatchResponse_Data {
	return &pb.GetDataBatchResponse_Data{
		Id:        d.ID,
		Name:      d.Name,
		Type:      d.Type,
		Data:      d.Data,
		IsChunked: d.IsChunked,
	}
}
