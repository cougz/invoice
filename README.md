# German Localization Support

This version of the invoice generator now includes German localization for businesses operating in Germany or German-speaking regions.

## German Features

- All labels and text in German
- German date format (DD.MM.YYYY)
- Default German VAT rate (19%)
- Euro (€) as default currency
- German company information in footer
- Support for extended invoice numbers with suffix

## Usage Examples

Generate a basic German invoice:

```bash
invoice generate --from "Meine Firma GmbH" \
    --to "Kunde GmbH" \
    --item "Beratungsleistung" --quantity 10 --rate 120 \
    --tax 0.19 \
    --note "Zahlbar innerhalb von 14 Tagen ohne Abzug."
```

Using invoice number suffix:

```bash
invoice generate --id "2023001" --id-suffix "-R1" \
    --from "Meine Firma GmbH" \
    --to "Kunde GmbH" \
    --item "Software-Lizenz" --quantity 1 --rate 499 \
    --tax 0.19
```

## Configuration File Example

Save repeated information with JSON / YAML:

```json
{
    "logo": "/path/to/logo.png",
    "from": "Meine Firma GmbH\nMusterstraße 123\n10115 Berlin",
    "tax": 0.19,
    "currency": "EUR",
    "note": "Bitte überweisen Sie den Betrag innerhalb von 14 Tagen."
}
```

Generate a new invoice by importing the configuration file:

```bash
invoice generate --import path/to/data.json \
    --to "Kunde GmbH\nKundenweg 42\n80331 München" \
    --item "Support-Paket" --quantity 1 --rate 299
```

## Customization

You can adjust the footer information by modifying the `writeFooter` function in `pdf.go` to include your specific:

- Company name
- Registration court and number
- VAT ID number
- Address
- Contact information
- Bank details

## Currency Configuration

The invoice generator now supports custom currency configurations through JSON files. You can:

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
