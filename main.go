package main

import (
        _ "embed"
        "flag"
        "fmt"
        "log"
        "os"
        "strings"
        "sort"
        "time"

        "github.com/signintech/gopdf"
        "github.com/spf13/cobra"
        "github.com/spf13/viper"
)

// Font paths for Inter fonts
const (
    InterRegularFont = "Inter/Inter Variable/Inter.ttf"
    InterBoldFont    = "Inter/Inter Hinted for Windows/Desktop/Inter-Bold.ttf"
)

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

type Invoice struct {
        Id            string `json:"id" yaml:"id"`
        IdSuffix      string `json:"idSuffix" yaml:"idSuffix"` // New field for invoice number suffix
        Title         string `json:"title" yaml:"title"`

        Logo string `json:"logo" yaml:"logo"`
        From string `json:"from" yaml:"from"`
        To   string `json:"to" yaml:"to"`
        Date string `json:"date" yaml:"date"`
        Due  string `json:"due" yaml:"due"`

        Items      []string  `json:"items" yaml:"items"`
        Quantities []int     `json:"quantities" yaml:"quantities"`
        Rates      []float64 `json:"rates" yaml:"rates"`

        Tax           float64 `json:"tax" yaml:"tax"`
        TaxExempt     bool    `json:"taxExempt" yaml:"taxExempt"` // Tax exemption (Kleinunternehmer-Regelung)
        Discount      float64 `json:"discount" yaml:"discount"`
        Currency      string  `json:"currency" yaml:"currency"` 

        Note string `json:"note" yaml:"note"`

        // Footer information
        Footer Footer `json:"footer" yaml:"footer"`
}

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

var (
        importPath     string
        output         string
        file           = Invoice{}
        defaultInvoice = DefaultInvoice()
)

func init() {
        viper.AutomaticEnv()

        generateCmd.Flags().StringVar(&importPath, "import", "", "Imported file (.json/.yaml)")
        generateCmd.Flags().StringVar(&file.Id, "id", time.Now().Format("20060102"), "ID")
        generateCmd.Flags().StringVar(&file.IdSuffix, "id-suffix", "", "Invoice Number Suffix (e.g. -R1, -A, etc.)")
        generateCmd.Flags().StringVar(&file.Title, "title", "RECHNUNG", "Title")

        generateCmd.Flags().Float64SliceVarP(&file.Rates, "rate", "r", defaultInvoice.Rates, "Rates")
        generateCmd.Flags().IntSliceVarP(&file.Quantities, "quantity", "q", defaultInvoice.Quantities, "Quantities")
        generateCmd.Flags().StringSliceVarP(&file.Items, "item", "i", defaultInvoice.Items, "Items")

        generateCmd.Flags().StringVarP(&file.Logo, "logo", "l", defaultInvoice.Logo, "Company logo")
        generateCmd.Flags().StringVarP(&file.From, "from", "f", defaultInvoice.From, "Issuing company")
        generateCmd.Flags().StringVarP(&file.To, "to", "t", defaultInvoice.To, "Recipient company")
        generateCmd.Flags().StringVar(&file.Date, "date", defaultInvoice.Date, "Date")
        generateCmd.Flags().StringVar(&file.Due, "due", defaultInvoice.Due, "Payment due date")

        generateCmd.Flags().Float64Var(&file.Tax, "tax", defaultInvoice.Tax, "Tax")
        generateCmd.Flags().BoolVar(&file.TaxExempt, "tax-exempt", defaultInvoice.TaxExempt, "Tax exemption (Kleinunternehmer-Regelung)")
        generateCmd.Flags().Float64VarP(&file.Discount, "discount", "d", defaultInvoice.Discount, "Discount")
        generateCmd.Flags().StringVarP(&file.Currency, "currency", "c", defaultInvoice.Currency, "Currency")

        generateCmd.Flags().StringVarP(&file.Note, "note", "n", "", "Note")
        generateCmd.Flags().StringVarP(&output, "output", "o", "invoice.pdf", "Output file (.pdf)")

        flag.Parse()
}

var rootCmd = &cobra.Command{
        Use:   "invoice",
        Short: "Invoice generates invoices from the command line.",
        Long:  `Invoice generates invoices from the command line.`,
}

var generateCmd = &cobra.Command{
        Use:   "generate",
        Short: "Generate an invoice",
        Long:  `Generate an invoice`,
        RunE: func(cmd *cobra.Command, args []string) error {
                if importPath != "" {
                        err := importData(importPath, &file, cmd.Flags())
                        if err != nil {
                                return fmt.Errorf("import failed: %v", err)
                        }
                }

                // Combine ID and IdSuffix for the full invoice number
                fullInvoiceId := file.Id
                if file.IdSuffix != "" {
                        fullInvoiceId = file.Id + file.IdSuffix
                }

                pdf := gopdf.GoPdf{}
                pdf.Start(gopdf.Config{
                        PageSize: *gopdf.PageSizeA4,
                })
                pdf.SetMargins(40, 40, 40, 40)
                pdf.AddPage()
                // Check if font files exist before attempting to load them
                if _, err := os.Stat(InterRegularFont); os.IsNotExist(err) {
                        return fmt.Errorf("Error: The Inter fonts are missing. Please download and restore the Inter font files.\n"+
                                "You can download them from: https://github.com/rsms/inter\n"+
                                "Directories needed:\n"+
                                "- %s\n"+
                                "- %s", InterRegularFont, InterBoldFont)
                }
                
                if _, err := os.Stat(InterBoldFont); os.IsNotExist(err) {
                        return fmt.Errorf("Error: The Inter fonts are missing. Please download and restore the Inter font files.\n"+
                                "You can download them from: https://github.com/rsms/inter\n"+
                                "Directories needed:\n"+
                                "- %s\n"+
                                "- %s", InterRegularFont, InterBoldFont)
                }
                
                // Load the Inter font from file
                err := pdf.AddTTFFont("Inter", InterRegularFont)
                if err != nil {
                        return fmt.Errorf("failed to load Inter font: %v", err)
                }
                
                // Load the Inter-Bold font from file
                err = pdf.AddTTFFont("Inter-Bold", InterBoldFont)
                if err != nil {
                        return fmt.Errorf("failed to load Inter-Bold font: %v", err)
                }

                writeLogo(&pdf, file.Logo, file.From)
                writeTitle(&pdf, file.Title, fullInvoiceId, file.Date) // Use full invoice ID with suffix
                writeBillTo(&pdf, file.To)
                writeHeaderRow(&pdf)
                subtotal := 0.0
                // Check if we have any items
                if len(file.Items) > 0 {
                    for i := range file.Items {
                        q := 1
                        if len(file.Quantities) > i {
                                q = file.Quantities[i]
                        }

                        r := 0.0
                        if len(file.Rates) > i {
                                r = file.Rates[i]
                        }

                        writeRow(&pdf, file.Items[i], q, r)
                        subtotal += float64(q) * r
                    }
                }

                // Write notes first before totals
                if file.Note != "" {
                        writeNotes(&pdf, file.Note)
                }

                // Then write totals (will be positioned on the right side)
                writeTotals(&pdf, subtotal, subtotal*file.Tax, subtotal*file.Discount)

                if file.Due != "" {
                        writeDueDate(&pdf, file.Due)
                }
                writeFooter(&pdf, fullInvoiceId) // Use full invoice ID with suffix in footer
                
                // Always use invoice ID for the filename, unless an explicit output is provided
                outputFile := fullInvoiceId + ".pdf"
                if output != "invoice.pdf" {
                    // User specified a custom output filename
                    outputFile = strings.TrimSuffix(output, ".pdf") + ".pdf"
                }
                
                err = pdf.WritePdf(outputFile)
                if err != nil {
                        return err
                }

                fmt.Printf("Generated %s\n", outputFile)
                
                // Set the output variable to the actual file path used
                output = outputFile

                return nil
        },
}

// Currency command definitions
var currencyCmd = &cobra.Command{
	Use:   "currency",
	Short: "Manage currency settings",
	Long:  `Manage currency settings for invoice generation.`,
}

// Web server command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start the web server",
	Long:  `Start a web server for creating invoices through a browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		webConfigPath := cmd.Flag("config").Value.String()
		webConfig := DefaultWebConfig()
		
		if webConfigPath != "" {
			var err error
			webConfig, err = loadWebConfig(webConfigPath)
			if err != nil {
				return fmt.Errorf("failed to load web config: %v", err)
			}
		}
		
		fmt.Printf("Starting invoice web server on port %d...\n", webConfig.Port)
		fmt.Printf("To access the web interface, open http://localhost:%d in your browser\n", webConfig.Port)
		
		return runWebServer(webConfig)
	},
}

var listCurrenciesCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available currencies and their symbols",
	Long:  `List all available currencies and their symbols.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available currencies and their symbols:")
		fmt.Println("---------------------------------------")
		
		// Get all currency codes sorted alphabetically
		var codes []string
		for code := range currencySymbols {
			codes = append(codes, code)
		}
		sort.Strings(codes)
		
		// Print each currency code and symbol
		for _, code := range codes {
			symbol := currencySymbols[code]
			fmt.Printf("%-5s : %s\n", code, symbol)
		}
	},
}

var exportConfigCmd = &cobra.Command{
	Use:   "export [path]",
	Short: "Export the current currency configuration",
	Long:  `Export the current currency configuration to a JSON file.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := "currency_config.json"
		if len(args) > 0 {
			configPath = args[0]
		}
		
		err := exportCurrencyConfig(configPath)
		if err != nil {
			return err
		}
		
		fmt.Printf("Currency configuration exported to %s\n", configPath)
		return nil
	},
}

func init() {
	// Add web server flags
	webCmd.Flags().String("config", "config/web_config.json", "Path to web server configuration file")
}

func main() {
	// Add currency subcommands
	currencyCmd.AddCommand(listCurrenciesCmd)
	currencyCmd.AddCommand(exportConfigCmd)
	
	// Add main commands
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(currencyCmd)
	rootCmd.AddCommand(webCmd)
	
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
