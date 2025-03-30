package domain

import (
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
)

type User struct {
	ID             string
	Login          string
	PasswordHash   string
	PublicKey      string
	PrivateKeyHash string
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
