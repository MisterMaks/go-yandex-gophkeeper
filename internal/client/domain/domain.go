package domain

import (
	"time"
)

const (
	LoginPasswordDataType                     = "LOGIN_PASSWORD"
	BankCardDataType                          = "BANK_CARD"
	TextDataType                              = "TEXT"
	BinaryDataType                            = "BINARY"
	BankCardDataTypeExpirationDateFormat      = "01/06"
	ChunkSize                            uint = 100 * 1024
	DataSizeLimit                             = 15 * 1024 * 1024
)

var EncryptedDataTypes = map[string]struct{}{
	LoginPasswordDataType: {},
	BankCardDataType:      {},
}

type LoginPasswordType struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type BankCardType struct {
	Number         string    `json:"number"`
	ExpirationDate time.Time `json:"expiration_date"`
	SecurityCode   string    `json:"security_code"`
}

type TextType struct {
	Text string `json:"text"`
}

type BinaryType struct {
	Data []byte `json:"data"`
}

type Data struct {
	ID        string
	Name      string
	Type      string
	Data      []byte
	IsChunked bool
}
