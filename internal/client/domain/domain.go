package domain

import (
	"time"
)

const (
	LoginPasswordDataType                = "LOGIN_PASSWORD"
	BankCardDataType                     = "BANK_CARD"
	TextDataType                         = "TEXT"
	BinaryDataType                       = "BINARY"
	BankCardDataTypeExpirationDateFormat = "01/06"
)

var EncryptedDataTypes = map[string]struct{}{
	LoginPasswordDataType: {},
	BankCardDataType:      {},
}

var DataTypeStructs = map[string]interface{}{
	LoginPasswordDataType: LoginPassword{},
	BankCardDataType:      BankCard{},
	TextDataType:          Text{},
	BinaryDataType:        Binary{},
}

type LoginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BankCard struct {
	Number         string    `json:"number"`
	ExpirationDate time.Time `json:"expiration_date"`
	SecurityCode   string    `json:"security_code"`
}

type Text struct {
	Text string `json:"text"`
}

type Binary struct {
	Data []byte `json:"data"`
}

type Data struct {
	ID   string
	Name string
	Type string
	Data []byte
}
