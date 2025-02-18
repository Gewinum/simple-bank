package random

import (
	"simple-bank/internal/utils"
)

func AccountBalance() int64 {
	return Int64(1, 1000)
}

func AccountCurrency() string {
	currencies := []string{
		utils.CurrencyUSD,
		utils.CurrencyEUR,
		utils.CurrencyCAD,
	}
	n := Int(0, len(currencies)-1)
	return currencies[n]
}
