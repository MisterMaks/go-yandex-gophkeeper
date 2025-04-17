package domain

import (
	"testing"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/stretchr/testify/assert"
)

func TestData_SerializeToProtobuf(t *testing.T) {
	now := time.Now()

	data := &Data{
		ID:        "1",
		UserID:    "2",
		Name:      "name",
		Type:      pb.Type_LOGIN_PASSWORD.String(),
		Data:      []byte("Hello world!"),
		CreatedAt: now,
		UpdatedAt: now,
	}

	protobufData := &pb.GetDataBatchResponse_Data{
		Id:   "1",
		Name: "name",
		Type: pb.Type_LOGIN_PASSWORD,
		Data: []byte("Hello world!"),
	}

	actualProtobufData := data.SerializeToProtobuf()

	assert.Equal(t, protobufData, actualProtobufData)
}
