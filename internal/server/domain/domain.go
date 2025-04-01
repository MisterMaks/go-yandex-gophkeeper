package domain

import (
	"bytes"
	"errors"
	"io"
	"sync"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
)

type Status string

const (
	StatusUploaded  Status = "UPLOADED"
	StatusUploading Status = "UPLOADING"
)

var (
	ErrLoginTaken                 = errors.New("login already taken")
	ErrInvalidLoginPassword       = errors.New("invalid login/password")
	ErrInvalidLoginPasswordFormat = errors.New("invalid login/password format")
)

type User struct {
	ID               string
	Login            string
	PasswordHash     []byte
	PublicKey        []byte
	PrivateKeyCipher []byte
}

type Data struct {
	ID          string
	UserID      string
	Name        string
	Type        string
	SizeInBytes uint64
	Status      Status
	Data        []byte
	ExternalID  *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (d Data) SerializeToProtobuf() *pb.GetDataBatchResponse_Data {
	return &pb.GetDataBatchResponse_Data{
		Id:          d.ID,
		Name:        d.Name,
		Type:        d.Type,
		SizeInBytes: d.SizeInBytes,
		Data:        d.Data,
		IsChunked:   d.ExternalID != nil,
	}
}

type DataChunkBuffer struct {
	buf   bytes.Buffer
	isEOF bool
	mu    sync.Mutex
}

func NewDataChunkBuffer() *DataChunkBuffer {
	return &DataChunkBuffer{
		buf:   bytes.Buffer{},
		isEOF: false,
		mu:    sync.Mutex{},
	}
}

func (dcb *DataChunkBuffer) Read(p []byte) (n int, err error) {
	dcb.mu.Lock()

	n, err = dcb.buf.Read(p)

	if n == 0 || err == io.EOF {
		if dcb.isEOF {
			return n, err
		}

		dcb.mu.Unlock()

		n, err = dcb.Read(p)
	}

	dcb.mu.Unlock()
	return n, err
}

func (dcb *DataChunkBuffer) Write(p []byte) (n int, err error) {
	dcb.mu.Lock()
	defer dcb.mu.Unlock()

	return dcb.buf.Write(p)
}

func (dcb *DataChunkBuffer) SetIsEOF() {
	dcb.mu.Lock()
	defer dcb.mu.Unlock()

	dcb.isEOF = true
}
