package reposytory

import "my-bank-service/internal/data"

type BalanceRepository interface {
	Create(balance *data.Balance) error
}
