package currency

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Service defines the interface for currency operations
type Service interface {
	GetSymbol(currency string) string
	GetAvailableCurrencies() map[string]string
	LoadConfig(configPath string) error
	ExportConfig(configPath string) error
}

// DefaultCurrencyService implements the Service interface
type DefaultCurrencyService struct {
	symbols map[string]string
}

// NewCurrencyService creates a new DefaultCurrencyService instance
func NewCurrencyService() Service {
	service := &DefaultCurrencyService{
		symbols: make(map[string]string),
	}
	
	// Initialize with default symbols
	service.initDefaultSymbols()
	
	// Try to load configuration from standard locations
	service.loadConfigFromStandardLocations()
	
	return service
}

// initDefaultSymbols initializes the service with default currency symbols
func (s *DefaultCurrencyService) initDefaultSymbols() {
	// Default built-in currency symbols
	defaultSymbols := map[string]string{
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
	
	// Copy default symbols to service
	for code, symbol := range defaultSymbols {
		s.symbols[code] = symbol
	}
}

// loadConfigFromStandardLocations tries to load configuration from standard locations
func (s *DefaultCurrencyService) loadConfigFromStandardLocations() {
	// Look for currency configuration in standard locations
	configLocations := []string{
		filepath.Join("config", "currency.json"),
		filepath.Join(os.Getenv("HOME"), ".config", "invoice", "currency.json"),
	}
	
	for _, location := range configLocations {
		if err := s.LoadConfig(location); err == nil {
			// Successfully loaded
			return
		}
	}
}

// GetSymbol returns the symbol for the given currency
func (s *DefaultCurrencyService) GetSymbol(currency string) string {
	if currency == "" {
		return ""
	}
	
	// Normalize to uppercase
	currencyUpper := strings.ToUpper(currency)
	
	symbol, exists := s.symbols[currencyUpper]
	if !exists {
		// If the currency doesn't exist in our map, return the currency code as fallback
		return currency + " "
	}
	return symbol
}

// GetAvailableCurrencies returns all available currencies and their symbols
func (s *DefaultCurrencyService) GetAvailableCurrencies() map[string]string {
	// Create a copy to avoid modifying the internal map
	result := make(map[string]string)
	for code, symbol := range s.symbols {
		result[code] = symbol
	}
	return result
}

// LoadConfig loads currency configuration from a JSON file
func (s *DefaultCurrencyService) LoadConfig(configPath string) error {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", configPath)
	}
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("unable to read config file: %v", err)
	}
	
	// Try to unmarshal directly into a map
	var symbols map[string]string
	err = json.Unmarshal(data, &symbols)
	if err == nil {
		// If successful, merge with existing symbols
		for code, symbol := range symbols {
			s.symbols[strings.ToUpper(code)] = symbol
		}
		return nil
	}
	
	// If direct unmarshaling failed, try with a structured format
	var config struct {
		Symbols map[string]string `json:"symbols"`
	}
	
	err = json.Unmarshal(data, &config)
	if err != nil {
		return fmt.Errorf("invalid JSON format: %v", err)
	}
	
	// Merge with existing symbols
	for code, symbol := range config.Symbols {
		s.symbols[strings.ToUpper(code)] = symbol
	}
	
	return nil
}

// ExportConfig exports currency configuration to a JSON file
func (s *DefaultCurrencyService) ExportConfig(configPath string) error {
	// Create a config struct
	config := struct {
		Symbols map[string]string `json:"symbols"`
	}{
		Symbols: s.symbols,
	}
	
	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}
	
	// If no directory specified, use config directory
	if filepath.Dir(configPath) == "." {
		configPath = filepath.Join("config", configPath)
	}
	
	// Ensure directory exists
	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}
	
	// Write to file
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}
	
	return nil
}