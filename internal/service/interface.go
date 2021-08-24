package service

type AccountInterface interface {
	// AddFunds Позволяет внести на счёт сумму sum
	AddFunds(sum float64) error
	// sumProfit Рассчитывает процент по вкладу и полученные деньги вносит на счёт
	sumProfit() error
	// Withdraw Производит списание со счёта по указанным правилам. Если списание выходит за рамки правил, выдаёт ошибку
	Withdraw(sum float64) error
	// GetCurrency Выдаёт валюту счёта
	GetCurrency() (string, error)
	// GetAccountCurrencyRate Выдаёт курс валюты счёта к передаваемой валюте cur
	GetAccountCurrencyRate(cur string) (float64, error)
	// GetBalance Выдаёт баланс счёта в указанной валюте
	GetBalance(cur string) (float64, error)
}
