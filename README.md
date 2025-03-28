# Invoice Generator with German Localization and Web Interface

This invoice generator application includes German localization for businesses operating in Germany or German-speaking regions, and features a web-based interface for easier invoice generation.

## Key Updates

### Major Architectural Improvements
- **Reorganized Codebase**: The entire codebase has been restructured following proper Go project practices with cmd, internal, and pkg directories
- **Decoupled Dependencies**: Interfaces for all services to improve testability and maintainability
- **Fixed WebUI Issues**: Resolved issues with tax exemption checkbox, footer settings, and company name display

## Features

- Command-line and web-based interfaces
- German localization (labels, date format, VAT)
- Euro (€) as default currency with support for other currencies
- Configurable invoice details via JSON/YAML files
- PDF generation with proper layout
- Dynamic positioning to handle various content lengths
- Support for extended invoice numbers with suffix
- Automatic file naming based on invoice ID
- Nextcloud integration for file sharing
- Tax exemption support (Kleinunternehmer-Regelung)
- Environment variable configuration support
- Proper service/handler separation for business logic

## Installation

### Build from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/cougz/invoice.git
   cd invoice
   ```

2. Build the application:
   ```bash
   go build -o invoice ./cmd/invoice
   ```

3. (Optional) Install to your PATH:
   ```bash
   go install ./cmd/invoice
   ```

This will create the `invoice` executable that you can run from the command line.

## Web Interface

The invoice generator includes a web server that provides a browser-based interface for creating invoices.

### Starting the Web Server

```bash
./invoice web [--config config/web_config.json]
```

The web server will start on port 8080 by default (configurable). Access the interface by opening `http://localhost:8080` in your browser.

### Web Configuration

Configure the web server by creating a `config/web_config.json` file:

```json
{
  "port": 8080,
  "nextcloudUrl": "https://your-nextcloud-server.com",
  "nextcloudShare": "/s/your-share-id",
  "uploadScript": "./cloudsend.sh"
}
```

### Environment Variables

You can also configure the application using environment variables:

```bash
INVOICE_WEB_PORT=8080 INVOICE_TAX_RATE=0.19 ./invoice web
```

### Nextcloud Integration

The web interface supports uploading and viewing generated invoices directly to a Nextcloud share. To use this feature:

1. Configure your Nextcloud settings in `config/web_config.json`
2. Make sure the `cloudsend.sh` script is executable (`chmod +x cloudsend.sh`)
3. After generating an invoice, click "Upload to Nextcloud" to share it

## Command-Line Usage

### Basic German Invoice

```bash
./invoice generate --from "Meine Firma GmbH" \
    --to "Kunde GmbH" \
    --item "Beratungsleistung" --quantity 10 --rate 120 \
    --tax 0.19 \
    --note "Zahlbar innerhalb von 14 Tagen ohne Abzug."
```

### Using Invoice Number Suffix

```bash
./invoice generate --id "2023001" --id-suffix "-R1" \
    --from "Meine Firma GmbH" \
    --to "Kunde GmbH" \
    --item "Software-Lizenz" --quantity 1 --rate 499 \
    --tax 0.19
```

### Using Configuration Files

Save repeated information with JSON / YAML:

```json
{
    "logo": "/path/to/logo.png",
    "from": "Meine Firma GmbH\nMusterstraße 123\n10115 Berlin",
    "tax": 0.19,
    "currency": "EUR",
    "note": "Bitte überweisen Sie den Betrag innerhalb von 14 Tagen.",
    "footer": {
      "companyName": "Meine Firma GmbH",
      "registrationInfo": "Amtsgericht Berlin, HRB 123456",
      "vatId": "USt-IdNr. DE123456789",
      "address": "Musterstraße 123",
      "city": "Berlin",
      "zip": "10115",
      "phone": "+49 30 1234567",
      "email": "info@meinefirma.de",
      "website": "www.meinefirma.de",
      "bankName": "Sparkasse Berlin",
      "bankIban": "DE12 3456 7890 1234 5678 90",
      "bankBic": "BELADEBEXXX"
    }
}
```

Generate a new invoice by importing the configuration file:

```bash
./invoice generate --import config/data.json \
    --to "Kunde GmbH\nKundenweg 42\n80331 München" \
    --item "Support-Paket" --quantity 1 --rate 299
```

## Currency Management

The invoice generator supports custom currency configurations through JSON files.

### List Available Currencies

View all available currencies and their symbols:

```bash
./invoice currency list
```

### Export Currency Configuration

Export the current currency configuration to a JSON file:

```bash
./invoice currency export my_currencies.json
```

### Custom Currency Configuration

Create or modify a currency configuration file (`currency_config.json` or `config/currency.json`):

```json
{
  "symbols": {
    "USD": "$",
    "EUR": "€",
    "GBP": "£",
    "CHF": "CHF",
    "CAD": "C$",
    "AUD": "A$",
    "CUSTOM": "¤"
  }
}
```

The application will automatically load currency symbols from any of these locations:
- `currency_config.json` in the current directory
- `config/currency.json` in the current directory
- `~/.config/invoice/currency.json` in the user's home directory

## Code Structure

The application follows a clean architecture pattern:

```
/invoice
├── cmd/                      # Command-line applications
│   └── invoice/              # Main application
│       └── main.go           # Entry point
├── internal/                 # Private application code
│   ├── models/               # Data models
│   │   ├── invoice.go        # Invoice data structures
│   │   └── config.go         # Configuration models
│   ├── handlers/             # Request handlers
│   │   └── web.go            # Web interface handlers
│   ├── services/             # Business logic
│   │   ├── pdf/              # PDF generation
│   │   │   └── renderer.go   # PDF rendering logic
│   │   ├── currency/         # Currency handling
│   │   │   └── service.go    # Currency service
│   │   ├── invoice/          # Invoice generation
│   │   │   └── service.go    # Invoice service
│   │   └── storage/          # Data persistence (future use)
│   └── config/               # Config handling
│       └── loader.go         # Configuration loading
├── pkg/                      # Public libraries
│   └── templates/            # Invoice templates (future use)
├── web/                      # Web assets and templates
│   ├── templates/            # HTML templates (future use)
│   └── static/               # Static assets
│       └── css/              # CSS stylesheets
└── config/                   # Configuration files
    ├── currency.json         # Currency configuration
    └── web_config.json       # Web server configuration
```

## Future Improvements

The codebase now supports further improvements:

1. Adding proper unit and integration tests
2. Implementing database storage for invoices
3. Adding authentication and user management
4. Creating customizable PDF templates
5. Implementing a RESTful API
6. Internationalization and additional localization support