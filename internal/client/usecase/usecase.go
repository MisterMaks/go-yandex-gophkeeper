package usecase

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	pb "github.com/MisterMaks/go-yandex-gophkeeper/api/proto/service"
	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/domain"
)

var NumberStringReg = regexp.MustCompile(`^[0-9]+$`)

type GophkeeperClientInterface interface {
	Login(ctx context.Context, login, password string) (*pb.LoginResponse, error)
	Register(
		ctx context.Context,
		login, password string,
		publicKey, privateKeyCipher []byte,
	) (*pb.RegisterResponse, error)
	CreateData(ctx context.Context, name string, dataType string, data []byte) (*pb.CreateDataResponse, error)
	GetDataBatch(ctx context.Context) (*pb.GetDataBatchResponse, error)
	UpdateData(ctx context.Context, id, name, dataType string, data []byte) (*pb.UpdateDataResponse, error)
	DeleteData(ctx context.Context, id string) (*pb.DeleteDataResponse, error)
}

type Usecase struct {
	gophkeeperClient GophkeeperClientInterface
	publicKey        *rsa.PublicKey
	privateKey       *rsa.PrivateKey
}

func NewUsecase(gophkeeperClient GophkeeperClientInterface) *Usecase {
	return &Usecase{
		gophkeeperClient: gophkeeperClient,
	}
}

func (u *Usecase) Register(ctx context.Context, login, password string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return err
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&privateKey.PublicKey),
	})
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(password))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]
	privateKeyCipher := aesgcm.Seal(nil, nonce, privateKeyPEM.Bytes(), nil)
	_, err = u.gophkeeperClient.Register(ctx, login, password, publicKeyPEM.Bytes(), privateKeyCipher)
	if err != nil {
		return err
	}

	u.privateKey = privateKey
	u.publicKey = &privateKey.PublicKey

	return nil
}

func (u *Usecase) Login(ctx context.Context, login, password string) error {
	out, err := u.gophkeeperClient.Login(ctx, login, password)
	if err != nil {
		return err
	}

	key := sha256.Sum256([]byte(password))

	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return err
	}

	nonce := key[len(key)-aesgcm.NonceSize():]

	privateKeyBytes, err := aesgcm.Open(nil, nonce, out.PrivateKeyCipher, nil)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(privateKeyBytes)

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return err
	}

	block, _ = pem.Decode(out.PublicKey)

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return err
	}

	u.privateKey = privateKey
	u.publicKey = publicKey

	return nil
}

func (u *Usecase) GetDataBatch(ctx context.Context) ([]domain.Data, error) {
	dataBatch, err := u.gophkeeperClient.GetDataBatch(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]domain.Data, 0, len(dataBatch.DataBatch))

	for _, data := range dataBatch.DataBatch {
		dataType := data.Type
		d := data.Data
		_, ok := domain.EncryptedDataTypes[dataType.String()]
		if ok {
			d, err = rsa.DecryptPKCS1v15(rand.Reader, u.privateKey, d)
			if err != nil {
				return nil, err
			}
		}

		out = append(out, domain.Data{
			ID:   data.Id,
			Name: data.Name,
			Type: dataType.String(),
			Data: d,
		})
	}

	return out, nil
}

func (u *Usecase) DeleteData(ctx context.Context, id string) error {
	_, err := u.gophkeeperClient.DeleteData(ctx, id)
	return err
}

func (u *Usecase) CreateLoginPassword(ctx context.Context, name, login, password string) error {
	loginPassword := domain.LoginPasswordType{
		Login:    login,
		Password: password,
	}

	data, err := json.Marshal(loginPassword)
	if err != nil {
		return err
	}

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, u.publicKey, data)
	if err != nil {
		return err
	}

	_, err = u.gophkeeperClient.CreateData(ctx, name, domain.LoginPasswordDataType, encryptedData)
	return err
}

func (u *Usecase) CreateBankCard(ctx context.Context, name, number, expirationDate, securityCode string) error {
	expirationDateTime, err := time.Parse(domain.BankCardDataTypeExpirationDateFormat, expirationDate)
	if err != nil {
		return err
	}

	number = strings.ReplaceAll(number, " ", "")
	if len(number) != 16 || !NumberStringReg.MatchString(number) {
		return fmt.Errorf("invalid number")
	}

	if len(securityCode) != 3 || !NumberStringReg.MatchString(securityCode) {
		return fmt.Errorf("invalid security code")
	}

	bankCard := domain.BankCardType{
		Number:         number,
		ExpirationDate: expirationDateTime,
		SecurityCode:   securityCode,
	}

	data, err := json.Marshal(bankCard)
	if err != nil {
		return err
	}

	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, u.publicKey, data)
	if err != nil {
		return err
	}

	_, err = u.gophkeeperClient.CreateData(ctx, name, domain.BankCardDataType, encryptedData)
	return err
}

func (u *Usecase) CreateText(ctx context.Context, name, text string) error {
	t := domain.TextType{Text: text}

	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	_, err = u.gophkeeperClient.CreateData(ctx, name, domain.TextDataType, data)
	return err
}

func (u *Usecase) CreateBinary(ctx context.Context, filePath string) error {
	name := filepath.Base(filePath)

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	_, err = u.gophkeeperClient.CreateData(ctx, name, domain.BinaryDataType, fileBytes)
	return err
}
