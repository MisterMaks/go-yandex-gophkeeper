package domain

import (
	"errors"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
)

// Status is status type.
type Status string

// Uploading statuses.
const (
	StatusUploaded  Status = "UPLOADED"
	StatusUploading Status = "UPLOADING"
	StatusDeleting  Status = "DELETING"
)

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
	ID         string
	UserID     string
	Name       string
	Type       string
	Data       []byte
	TaskID     *int64
	ExternalID *int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
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
type StreamInterface interface {
	GetDataChunk() []byte
}

// ChunkedData implements the [io.Reader]
// using for saving chunk file from GRPC server stream.
type ChunkedData struct {
	Name        string
	Type        string
	SizeInBytes int64
	stream      pb.GoYandexGophkeeper_CreateChunkedDataServer
	buffer      []byte
}

// Read implements the [io.Reader] interface.
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

func NewChunkedDataFromStream(stream pb.GoYandexGophkeeper_CreateChunkedDataServer) (*ChunkedData, error) {
	chunk, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	name := chunk.GetMetaData().GetName()
	dataType := chunk.GetMetaData().GetType()
	size := chunk.GetMetaData().GetSizeInBytes()
	buffer := chunk.GetDataChunk()

	return &ChunkedData{
		Name:        name,
		Type:        dataType,
		SizeInBytes: size,
		stream:      stream,
		buffer:      buffer,
	}, nil
}
