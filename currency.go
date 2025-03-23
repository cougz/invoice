package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Default built-in currency symbols
var defaultCurrencySymbols = map[string]string{
	"USD": "$",
	"EUR": "€",
	"GBP": "£",
	"JPY": "¥",
	"CNY": "¥",
	"INR": "₹",
	"RUB": "₽",
	"KRW": "₩",
	"BRL": "R$",
	"SGD": "S$",
	"AUD": "A$",
	"CAD": "C$",
	"CHF": "CHF",
	"HKD": "HK$",
	"NZD": "NZ$",
	"SEK": "kr",
	"NOK": "kr",
	"DKK": "kr",
	"ZAR": "R",
	"MXN": "Mex$",
	"AED": "د.إ",
	"THB": "฿",
	"PLN": "zł",
}

// Custom currency configuration that can be loaded from a file
type CurrencyConfig struct {
	Symbols map[string]string `json:"symbols"`
}

// Global variable to store the merged currency symbols (default + custom)
var currencySymbols = make(map[string]string)

// Initialize the currency symbols map with default values
func init() {
	// Start with default symbols
	for code, symbol := range defaultCurrencySymbols {
		currencySymbols[code] = symbol
	}

	// Look for currency configuration in standard locations
	configLocations := []string{
		"currency_config.json",
		filepath.Join("config", "currency.json"),
		filepath.Join(os.Getenv("HOME"), ".config", "invoice", "currency.json"),
	}

	for _, location := range configLocations {
		if loadCurrencyConfig(location) {
			break // Stop once we've successfully loaded a config
		}
	}
}

// Load custom currency configuration from a JSON file
func loadCurrencyConfig(configPath string) bool {
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist or can't be read - this is fine, just use defaults
		return false
	}

	var config CurrencyConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Error parsing currency config file %s: %v\n", configPath, err)
		return false
	}

	// Merge custom symbols with default ones
	for code, symbol := range config.Symbols {
		currencySymbols[strings.ToUpper(code)] = symbol
	}

	fmt.Printf("Loaded custom currency symbols from %s\n", configPath)
	return true
}

// Helper function to safely get currency symbol
func getCurrencySymbol(currency string) string {
	if currency == "" {
		return ""
	}
	
	// Normalize to uppercase
	currencyUpper := strings.ToUpper(currency)
	
	symbol, exists := currencySymbols[currencyUpper]
	if !exists {
		// If the currency doesn't exist in our map, return the currency code as fallback
		return currency + " "
	}
	return symbol
}

// Export the currency configuration to a JSON file
func exportCurrencyConfig(configPath string) error {
	config := CurrencyConfig{
		Symbols: currencySymbols,
	}
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling currency config: %v", err)
	}
	
	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}
	
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing currency config file: %v", err)
	}
	
	return nil
}
