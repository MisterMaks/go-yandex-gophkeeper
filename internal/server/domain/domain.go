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

type LoginPassword struct {
	ID             string
	UserID         string
	Name           string
	LoginCipher    string
	PasswordCipher string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (lp LoginPassword) SerializeToProtobuf() *pb.GetLoginPasswordBatchResponse_LoginPassword {
	return &pb.GetLoginPasswordBatchResponse_LoginPassword{
		Id:             lp.ID,
		Name:           lp.Name,
		LoginCipher:    lp.LoginCipher,
		PasswordCipher: lp.PasswordCipher,
	}
}

type BankCard struct {
	ID                   string
	UserID               string
	Name                 string
	NumberCipher         string
	ExpirationDateCipher string
	SecurityCodeCipher   string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (bc BankCard) SerializeToProtobuf() *pb.GetBankCardBatchResponse_BankCard {
	return &pb.GetBankCardBatchResponse_BankCard{
		Id:                   bc.ID,
		Name:                 bc.Name,
		NumberCipher:         bc.NumberCipher,
		ExpirationDateCipher: bc.ExpirationDateCipher,
		SecurityCodeCipher:   bc.SecurityCodeCipher,
	}
}

type TextMeta struct {
	ID         string
	UserID     string
	Name       string
	TextPrefix string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (tm TextMeta) SerializeToProtobuf() *pb.GetTextMetaBatchResponse_TextMeta {
	return &pb.GetTextMetaBatchResponse_TextMeta{
		Id:         tm.ID,
		Name:       tm.Name,
		TextPrefix: tm.TextPrefix,
	}
}
