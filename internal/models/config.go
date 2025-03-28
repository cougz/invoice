package models

// WebConfig holds the configuration for the web server
type WebConfig struct {
	Port           int    `json:"port" yaml:"port" env:"PORT"`
	NextcloudURL   string `json:"nextcloudUrl" yaml:"nextcloudUrl" env:"NEXTCLOUD_URL"`
	NextcloudShare string `json:"nextcloudShare" yaml:"nextcloudShare" env:"NEXTCLOUD_SHARE"`
	UploadScript   string `json:"uploadScript" yaml:"uploadScript" env:"UPLOAD_SCRIPT"`
	TemplateDir    string `json:"templateDir" yaml:"templateDir" env:"TEMPLATE_DIR"`
}

// CurrencyConfig represents the currency configuration
type CurrencyConfig struct {
	Symbols map[string]string `json:"symbols" yaml:"symbols"`
}

// AppConfig represents the complete application configuration
type AppConfig struct {
	Web      WebConfig      `json:"web" yaml:"web"`
	Currency CurrencyConfig `json:"currency" yaml:"currency"`
	Debug    bool           `json:"debug" yaml:"debug" env:"DEBUG"`
}

// DefaultWebConfig returns a WebConfig with default values
func DefaultWebConfig() WebConfig {
	return WebConfig{
		Port:           8080,
		NextcloudURL:   "https://cloud.example.com",
		NextcloudShare: "/s/share-id",
		UploadScript:   "/var/scripts/cloudsend.sh",
		TemplateDir:    "web/templates",
	}
}

// DefaultAppConfig returns an AppConfig with default values
func DefaultAppConfig() AppConfig {
	return AppConfig{
		Web:      DefaultWebConfig(),
		Currency: CurrencyConfig{Symbols: map[string]string{}},
		Debug:    false,
	}
}

// InvoiceRequest represents the form data from the web UI
type InvoiceRequest struct {
	From             string   `json:"from"`
	To               string   `json:"to"`
	Items            string   `json:"items"`
	Quantities       string   `json:"quantities"`
	Rates            string   `json:"rates"`
	Tax              float64  `json:"tax"`
	TaxExempt        bool     `json:"taxExempt"`
	Discount         float64  `json:"discount"`
	Currency         string   `json:"currency"`
	Note             string   `json:"note"`
	Id               string   `json:"id"`
	IdSuffix         string   `json:"idSuffix"`
	ConfigFile       string   `json:"configFile"`
	UseConfig        bool     `json:"useConfig"`
	ShowRegistration bool     `json:"showRegistration"`
	ShowVatId        bool     `json:"showVatId"`
	CompanyName      string   `json:"companyName"`
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Message string `json:"message"`
}