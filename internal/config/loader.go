package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	
	"invoice/internal/models"
	
	"encoding/json"
	"gopkg.in/yaml.v3"
)

// ConfigLoader defines the interface for configuration loading
type ConfigLoader interface {
	Load(path string) (*models.AppConfig, error)
	LoadWeb(path string) (models.WebConfig, error)
	LoadInvoice(path string) (*models.Invoice, error)
	ApplyEnvironmentVariables(config interface{}) error
}

// FileConfigLoader implements the ConfigLoader interface
type FileConfigLoader struct{}

// NewConfigLoader creates a new FileConfigLoader
func NewConfigLoader() ConfigLoader {
	return &FileConfigLoader{}
}

// Load loads the application configuration from a file
func (l *FileConfigLoader) Load(path string) (*models.AppConfig, error) {
	config := models.DefaultAppConfig()
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &config, fmt.Errorf("config file does not exist: %s", path)
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		return &config, fmt.Errorf("unable to read config file: %v", err)
	}
	
	// Remove UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	
	// Check file type and parse accordingly
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	default:
		return &config, fmt.Errorf("unsupported file type: %s", ext)
	}
	
	if err != nil {
		return &config, fmt.Errorf("error parsing config file: %v", err)
	}
	
	// Apply environment variables
	if err := l.ApplyEnvironmentVariables(&config); err != nil {
		return &config, fmt.Errorf("error applying environment variables: %v", err)
	}
	
	return &config, nil
}

// LoadWeb loads the web configuration from a file
func (l *FileConfigLoader) LoadWeb(path string) (models.WebConfig, error) {
	config := models.DefaultWebConfig()
	
	// If path is empty, just use the default config with environment variables
	if path == "" {
		l.ApplyEnvironmentVariables(&config)
		return config, nil
	}
	
	// Check if path doesn't have a directory prefix, assume it's in config dir
	if filepath.Dir(path) == "." {
		path = filepath.Join("config", path)
	}
	
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		l.ApplyEnvironmentVariables(&config)
		return config, nil
	}
	
	data, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("unable to read web config: %v", err)
	}
	
	// Remove UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &config)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	default:
		return config, fmt.Errorf("unsupported file type: %s", ext)
	}
	
	if err != nil {
		return config, fmt.Errorf("invalid config file: %v", err)
	}
	
	// Apply environment variables
	if err := l.ApplyEnvironmentVariables(&config); err != nil {
		return config, fmt.Errorf("error applying environment variables: %v", err)
	}
	
	return config, nil
}

// LoadInvoice loads an invoice template from a file
func (l *FileConfigLoader) LoadInvoice(path string) (*models.Invoice, error) {
	// Create a default invoice
	invoice := models.DefaultInvoice()
	
	// Check if path doesn't have a directory prefix, assume it's in config dir
	if filepath.Dir(path) == "." {
		path = filepath.Join("config", path)
	}
	
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return &invoice, fmt.Errorf("unable to read file: %v", err)
	}
	
	// Remove UTF-8 BOM if present
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	
	// Check file type and parse accordingly
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &invoice)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &invoice)
	default:
		return &invoice, fmt.Errorf("unsupported file type: %s", ext)
	}
	
	if err != nil {
		return &invoice, fmt.Errorf("error parsing invoice file: %v", err)
	}
	
	return &invoice, nil
}

// ApplyEnvironmentVariables applies environment variables to config structs
func (l *FileConfigLoader) ApplyEnvironmentVariables(config interface{}) error {
	// Get the value and type of the input interface
	val := reflect.ValueOf(config)
	
	// If it's a pointer, get the value it points to
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	
	// Only work on structs
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("config must be a struct or pointer to struct")
	}
	
	typ := val.Type()
	
	// Iterate over all fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		
		// Get the env tag value
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			// If this field is a struct or pointer to struct, recurse
			if field.Kind() == reflect.Struct {
				l.ApplyEnvironmentVariables(field.Addr().Interface())
			} else if field.Kind() == reflect.Ptr && field.Elem().Kind() == reflect.Struct {
				l.ApplyEnvironmentVariables(field.Interface())
			}
			continue
		}
		
		// Get environment variable
		envValue := os.Getenv(envTag)
		if envValue == "" {
			// Also check for prefixed versions (INVOICE_PORT, etc.)
			envValue = os.Getenv("INVOICE_" + envTag)
			if envValue == "" {
				continue
			}
		}
		
		// Set the field value based on its type
		if field.IsValid() && field.CanSet() {
			switch field.Kind() {
			case reflect.String:
				field.SetString(envValue)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if intVal, err := strconv.ParseInt(envValue, 10, 64); err == nil {
					field.SetInt(intVal)
				}
			case reflect.Bool:
				if boolVal, err := strconv.ParseBool(envValue); err == nil {
					field.SetBool(boolVal)
				}
			case reflect.Float32, reflect.Float64:
				if floatVal, err := strconv.ParseFloat(envValue, 64); err == nil {
					field.SetFloat(floatVal)
				}
			}
		}
	}
	
	return nil
}