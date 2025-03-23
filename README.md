# Invoice Generator with German Localization and Web Interface

This invoice generator application includes German localization for businesses operating in Germany or German-speaking regions, and now features a web-based interface for easier invoice generation.

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

## Web Interface

The invoice generator now includes a web server that provides a browser-based interface for creating invoices.

### Starting the Web Server

```bash
invoice web [--config web_config.json]
```

The web server will start on port 8080 by default (configurable). Access the interface by opening `http://localhost:8080` in your browser.

### Web Configuration

Configure the web server by creating a `web_config.json` file:

```json
{
  "port": 8080,
  "nextcloudUrl": "https://your-nextcloud-server.com",
  "nextcloudShare": "/s/your-share-id",
  "uploadScript": "./cloudsend.sh"
}
```

### Nextcloud Integration

The web interface supports uploading generated invoices directly to a Nextcloud share. To use this feature:

1. Configure your Nextcloud settings in `web_config.json`
2. Make sure the `cloudsend.sh` script is executable (`chmod +x cloudsend.sh`)
3. After generating an invoice, click "Upload to Nextcloud" to share it

## Command-Line Usage

### Basic German Invoice

```bash
invoice generate --from "Meine Firma GmbH" \
    --to "Kunde GmbH" \
    --item "Beratungsleistung" --quantity 10 --rate 120 \
    --tax 0.19 \
    --note "Zahlbar innerhalb von 14 Tagen ohne Abzug."
```

### Using Invoice Number Suffix

```bash
invoice generate --id "2023001" --id-suffix "-R1" \
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
invoice generate --import path/to/data.json \
    --to "Kunde GmbH\nKundenweg 42\n80331 München" \
    --item "Support-Paket" --quantity 1 --rate 299
```

## Currency Management

The invoice generator now supports custom currency configurations through JSON files.

### List Available Currencies

View all available currencies and their symbols:

```bash
invoice currency list
```

### Export Currency Configuration

Export the current currency configuration to a JSON file:

```bash
invoice currency export my_currencies.json
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

If a currency is not found in the configuration, the application will use the currency code itself as a fallback.