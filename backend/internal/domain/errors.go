package domain

import "errors"

var (
	ErrUserNotFound             = errors.New("usuário não encontrado")
	ErrUserAlreadyExists        = errors.New("usuário com este e-mail já cadastrado")
	ErrInvalidCredentials       = errors.New("e-mail ou senha inválidos")
	ErrInvalidAmount            = errors.New("valor da transação deve ser positivo")
	ErrAccountRequired          = errors.New("conta bancária é obrigatória para esta transação")
	ErrAccountNotFound          = errors.New("conta não encontrada")
	ErrCreditCardNotFound       = errors.New("cartão de crédito não encontrado")
	ErrCategoryNotFound         = errors.New("categoria não encontrada")
	ErrTransactionNotFound      = errors.New("transação não encontrada")
	ErrPortfolioNotFound        = errors.New("carteira de investimentos não encontrada")
	ErrAssetNotFound            = errors.New("ativo não encontrado")
	ErrInsufficientAssetQuantity = errors.New("quantidade insuficiente do ativo para venda")
	ErrInvalidInstallmentCount  = errors.New("número de parcelas deve ser maior que zero")
	ErrInvalidAssetTicker       = errors.New("ticker do ativo inválido")
)
