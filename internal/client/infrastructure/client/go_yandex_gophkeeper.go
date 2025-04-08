package client

import (
	"context"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"google.golang.org/grpc/metadata"
)

type GophkeeperClient struct {
	client      pb.GoYandexGophkeeperClient
	accessToken string
}

func NewGophkeeperClient(client pb.GoYandexGophkeeperClient) *GophkeeperClient {
	return &GophkeeperClient{
		client: client,
	}
}

func (c *GophkeeperClient) Login(ctx context.Context, login, password string) (*pb.LoginResponse, error) {
	in := &pb.LoginRequest{
		Login:    login,
		Password: password,
	}

	out, err := c.client.Login(ctx, in)
	if err != nil {
		return nil, err
	}

	c.accessToken = out.AccessToken
	return out, err
}

func (c *GophkeeperClient) Register(
	ctx context.Context,
	login, password string,
	publicKey, privateKeyCipher []byte,
) (*pb.RegisterResponse, error) {
	in := &pb.RegisterRequest{
		Login:            login,
		Password:         password,
		PublicKey:        publicKey,
		PrivateKeyCipher: privateKeyCipher,
	}

	out, err := c.client.Register(ctx, in)
	if err != nil {
		return nil, err
	}

	c.accessToken = out.AccessToken
	return out, err
}

func (c *GophkeeperClient) CreateData(
	ctx context.Context,
	name, dataType string,
	data []byte,
) (*pb.CreateDataResponse, error) {
	ctx = c.getAuthCtx(ctx)

	in := &pb.CreateDataRequest{
		Name: name,
		Type: pb.Type(pb.Type_value[dataType]),
		Data: data,
	}

	out, err := c.client.CreateData(ctx, in)
	return out, err
}

func (c *GophkeeperClient) GetDataBatch(ctx context.Context) (*pb.GetDataBatchResponse, error) {
	ctx = c.getAuthCtx(ctx)
	in := &pb.GetDataBatchRequest{}
	out, err := c.client.GetDataBatch(ctx, in)
	return out, err
}

func (c *GophkeeperClient) UpdateData(
	ctx context.Context,
	id, name, dataType string,
	data []byte,
) (*pb.UpdateDataResponse, error) {
	ctx = c.getAuthCtx(ctx)

	in := &pb.UpdateDataRequest{
		Id:   id,
		Name: name,
		Type: pb.Type(pb.Type_value[dataType]),
		Data: data,
	}

	out, err := c.client.UpdateData(ctx, in)
	return out, err
}

func (c *GophkeeperClient) DeleteData(ctx context.Context, id string) (*pb.DeleteDataResponse, error) {
	ctx = c.getAuthCtx(ctx)
	in := &pb.DeleteDataRequest{Id: id}
	out, err := c.client.DeleteData(ctx, in)
	return out, err
}

func (c *GophkeeperClient) getAuthCtx(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("Authorization", "Bearer "+c.accessToken))
}
