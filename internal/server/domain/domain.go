package domain

import (
	"errors"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"google.golang.org/grpc"
)

type Status string

const (
	StatusUploaded  Status = "UPLOADED"
	StatusUploading Status = "UPLOADING"
)

var (
	ErrLoginTaken                    = errors.New("login already taken")
	ErrInvalidLoginPassword          = errors.New("invalid login/password")
	ErrInvalidLoginPasswordFormat    = errors.New("invalid login/password format")
	ErrDataWithThisNameAndTypeExists = errors.New("data with this name and type already exists")
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
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (d Data) SerializeToProtobuf() *pb.GetDataBatchResponse_Data {
	return &pb.GetDataBatchResponse_Data{
		Id:   d.ID,
		Name: d.Name,
		Type: pb.Type(pb.Type_value[d.Type]),
		Data: d.Data,
	}
}

type StreamInterface interface {
	GetDataChunk() []byte
}

type ChunkedData struct {
	stream grpc.ClientStreamingServer[StreamInterface, any]
	buffer []byte
}

func NewChunkedData(stream grpc.ClientStreamingServer[StreamInterface, any]) *ChunkedData {
	return &ChunkedData{
		stream: stream,
		buffer: []byte{},
	}
}

func (cd *ChunkedData) Read(p []byte) (n int, err error) {
	if len(cd.buffer) == 0 {
		chunk, err := cd.stream.Recv()
		if err != nil {
			return 0, err
		}

		cd.buffer = (*chunk).GetDataChunk()
	}

	n = copy(p, cd.buffer)
	cd.buffer = cd.buffer[n:]

	return n, nil
}
