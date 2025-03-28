package models

import (
	"time"
)

// Footer holds the invoice footer information
type Footer struct {
	CompanyName      string `json:"companyName" yaml:"companyName"`
	RegistrationInfo string `json:"registrationInfo" yaml:"registrationInfo"`
	ShowRegistration bool   `json:"showRegistration" yaml:"showRegistration"`
	VatId            string `json:"vatId" yaml:"vatId"`
	ShowVatId        bool   `json:"showVatId" yaml:"showVatId"`
	Address          string `json:"address" yaml:"address"`
	City             string `json:"city" yaml:"city"`
	Zip              string `json:"zip" yaml:"zip"`
	Phone            string `json:"phone" yaml:"phone"`
	Email            string `json:"email" yaml:"email"`
	Website          string `json:"website" yaml:"website"`
	BankName         string `json:"bankName" yaml:"bankName"`
	BankIban         string `json:"bankIban" yaml:"bankIban"`
	BankBic          string `json:"bankBic" yaml:"bankBic"`
}

// Invoice represents an invoice with all its data
type Invoice struct {
	Id            string  `json:"id" yaml:"id"`
	IdSuffix      string  `json:"idSuffix" yaml:"idSuffix"`
	Title         string  `json:"title" yaml:"title"`
	Logo          string  `json:"logo" yaml:"logo"`
	From          string  `json:"from" yaml:"from"`
	To            string  `json:"to" yaml:"to"`
	Date          string  `json:"date" yaml:"date"`
	Due           string  `json:"due" yaml:"due"`
	Items         []string  `json:"items" yaml:"items"`
	Quantities    []int     `json:"quantities" yaml:"quantities"`
	Rates         []float64 `json:"rates" yaml:"rates"`
	Tax           float64 `json:"tax" yaml:"tax"`
	TaxExempt     bool    `json:"taxExempt" yaml:"taxExempt"`
	Discount      float64 `json:"discount" yaml:"discount"`
	Currency      string  `json:"currency" yaml:"currency"`
	Note          string  `json:"note" yaml:"note"`
	Footer        Footer  `json:"footer" yaml:"footer"`
}

// DefaultFooter returns a new footer with default values
func DefaultFooter() Footer {
	return Footer{
		CompanyName:      "Firma GmbH",
		RegistrationInfo: "Registergericht München, HRB 123456",
		ShowRegistration: true,  // Default to showing registration info
		VatId:            "USt-IdNr. DE123456789",
		ShowVatId:        true,  // Default to showing VAT ID
		Address:          "Musterstraße 123",
		City:             "München",
		Zip:              "80331",
		Phone:            "+49 89 1234567",
		Email:            "info@firma.de",
		Website:          "www.firma.de",
		BankName:         "Sparkasse München",
		BankIban:         "DE12 3456 7890 1234 5678 90",
		BankBic:          "ABCDEFGHXXX",
	}
}

// DefaultInvoice returns a new invoice with default values
func DefaultInvoice() Invoice {
	return Invoice{
		Id:         time.Now().Format("20060102"),
		IdSuffix:   "",  // Default empty suffix
		Title:      "RECHNUNG", // Use German title
		Rates:      []float64{25},
		Quantities: []int{2},
		Items:      []string{"Dienstleistung"}, // Changed to German default
		From:       "Firma GmbH",  // Changed to German default
		To:         "Kunde GmbH",  // Changed to German default
		Date:       time.Now().Format("02.01.2006"), // German date format (day.month.year)
		Due:        time.Now().AddDate(0, 0, 14).Format("02.01.2006"), // German date format
		Tax:        0.19, // Default German VAT rate (19%)
		TaxExempt:  false, // Default to tax inclusion
		Discount:   0,
		Currency:   "EUR", // Default to Euro
		Footer:     DefaultFooter(), // Default footer information
	}
}

// InvoiceItem represents a single item in an invoice
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Rate        float64 `json:"rate"`
	Amount      float64 `json:"amount"`
}

// CalculateTotal calculates the total amount for the invoice
func CalculateTotal(invoice *Invoice) float64 {
	subtotal := 0.0
	
	// Calculate subtotal from items
	for i, item := range invoice.Items {
		quantity := 1
		if i < len(invoice.Quantities) {
			quantity = invoice.Quantities[i]
		}
		
		rate := 0.0
		if i < len(invoice.Rates) {
			rate = invoice.Rates[i]
		}
		
		subtotal += float64(quantity) * rate
	}
	
	// Apply discount if any
	discountAmount := subtotal * invoice.Discount
	afterDiscount := subtotal - discountAmount
	
	// Apply tax if not exempt
	if !invoice.TaxExempt {
		taxAmount := afterDiscount * invoice.Tax
		return afterDiscount + taxAmount
	}
	
	return afterDiscount
}