package ui

import (
	"context"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type UsecaseInterface interface {
	Register(ctx context.Context, login, password string) error
	Login(ctx context.Context, login, password string) error
	GetDataBatch(ctx context.Context) ([]domain.Data, error)
	DeleteData(ctx context.Context, id string) error
	CreateLoginPassword(ctx context.Context, name, login, password string) error
	CreateBankCard(ctx context.Context, name, number, expirationDate, securityCode string) error
	CreateText(ctx context.Context, name, text string) error
	CreateBinary(ctx context.Context, filePath string) error
}

type TUI struct {
	usecase UsecaseInterface
}

func NewTUI(usecase UsecaseInterface) *TUI {
	return &TUI{usecase: usecase}
}

func (tui *TUI) Run() error {
	loginPasswordTypeModel := NewLoginPasswordTypeModel(tui.usecase)
	bankCardTypeModel := NewBankCardTypeModel(tui.usecase)
	textTypeModel := NewTextTypeModel(tui.usecase)
	binaryTypeModel := NewBinaryTypeModel(tui.usecase)

	typeListModel := NewTypeListModel(
		loginPasswordTypeModel,
		bankCardTypeModel,
		textTypeModel,
		binaryTypeModel,
	)
	homeModel := NewHomeModel(tui.usecase, typeListModel)

	loginPasswordTypeModel.homeModel = homeModel
	loginPasswordTypeModel.typeListModel = typeListModel

	bankCardTypeModel.homeModel = homeModel
	bankCardTypeModel.typeListModel = typeListModel

	textTypeModel.homeModel = homeModel
	textTypeModel.typeListModel = typeListModel

	binaryTypeModel.homeModel = homeModel
	binaryTypeModel.typeListModel = typeListModel

	typeListModel.homeModel = homeModel

	loginFormModel := NewLoginFormModel(tui.usecase, homeModel)
	loginModel := NewLoginModel(loginFormModel)

	p := tea.NewProgram(loginModel)
	_, err := p.Run()

	return err
}
