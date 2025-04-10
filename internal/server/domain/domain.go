package domain

import (
	"errors"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
)

// Status is status type.
//type Status string

// Uploadingstatuses.
//const (
//	StatusUploaded  Status = "UPLOADED"
//	StatusUploading Status = "UPLOADING"
//)

// Server errors.
var (
	ErrLoginTaken                    = errors.New("login already taken")
	ErrInvalidLoginPassword          = errors.New("invalid login/password")
	ErrInvalidLoginPasswordFormat    = errors.New("invalid login/password format")
	ErrDataWithThisNameAndTypeExists = errors.New("data with this name and type already exists")
)

// User is user struct.
type User struct {
	ID               string
	Login            string
	PasswordHash     []byte
	PublicKey        []byte
	PrivateKeyCipher []byte
}

// Data is data struct.
type Data struct {
	ID        string
	UserID    string
	Name      string
	Type      string
	Data      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SerializeToProtobuf serializes data to protobuf format.
func (d Data) SerializeToProtobuf() *pb.GetDataBatchResponse_Data {
	return &pb.GetDataBatchResponse_Data{
		Id:   d.ID,
		Name: d.Name,
		Type: pb.Type(pb.Type_value[d.Type]),
		Data: d.Data,
	}
}

// StreamInterface contains needed funcs for saving chunked data.
//type StreamInterface interface {
//	GetDataChunk() []byte
//}

// ChunkedData implements the [io.Reader]
// using for saving chunk file from GRPC server stream.
//type ChunkedData struct {
//	stream grpc.ClientStreamingServer[StreamInterface, any]
//	buffer []byte
//}

// Read implements the [io.Reader] interface.
//func (cd *ChunkedData) Read(p []byte) (n int, err error) {
//	if len(cd.buffer) == 0 {
//		chunk, err := cd.stream.Recv()
//		if err != nil {
//			return 0, err
//		}
//
//		cd.buffer = (*chunk).GetDataChunk()
//	}
//
//	n = copy(p, cd.buffer)
//	cd.buffer = cd.buffer[n:]
//
//	return n, nil
//}
