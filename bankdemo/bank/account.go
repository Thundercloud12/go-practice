package bank

import(
	"errors"
)

type Account struct {
	Owner string
	balance float64
}

func NewAccount(owner string, initialBalance float64) *Account {
	return &Account{Owner: owner, balance: initialBalance}
}

func (a *Account) Deposit(amount float64) {
    a.balance += amount
}

func (a *Account) Withdraw(amount float64) error {
    if a.balance <amount {
		return errors.New("Insufficient Balance")
	}

	a.balance = a.balance-amount
	return nil
}

func (a *Account) Balance() float64 {
    return a.balance
}