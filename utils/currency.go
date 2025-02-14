package utils

const (
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyCAD = "CAD"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case CurrencyUSD, CurrencyEUR, CurrencyCAD:
		return true
	default:
		return false
	}
}
