package main

var currencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"CNY": "¥",
	"INR": "₹",
	"RUB": "₽",
	"KRW": "₩",
	"BRL": "R$",
	"SGD": "SGD$",
}

// Helper function to safely get currency symbol
func getCurrencySymbol(currency string) string {
	symbol, exists := currencySymbols[currency]
	if !exists {
		// If the currency doesn't exist in our map, return the currency code as fallback
		return currency + " "
	}
	return symbol
}
