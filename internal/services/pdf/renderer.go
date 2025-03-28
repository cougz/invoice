package pdf

import (
	"fmt"
	"image"
	"io"
	"os"
	"strconv"
	"strings"
	
	"invoice/internal/models"
	"invoice/internal/services/currency"
	
	"github.com/signintech/gopdf"
)

// Constants for PDF layout
const (
	quantityColumnOffset = 390
	rateColumnOffset     = 450
	amountColumnOffset   = 510
)

// German label translations
const (
	invoiceTitle   = "RECHNUNG"
	billToLabel    = "RECHNUNG AN"
	itemLabel      = "ARTIKEL UND BESCHREIBUNG"
	qtyLabel       = "MENGE"
	rateLabel      = "PREIS"
	amountLabel    = "BETRAG"
	notesLabel     = "HINWEISE"
	subtotalLabel  = "Zwischensumme"
	discountLabel  = "Rabatt"
	taxLabel       = "MwSt."
	totalLabel     = "Gesamt"
	dueDateLabel   = "Fälligkeitsdatum"
)

// Font paths for Inter fonts
const (
	InterRegularFont = "Inter/Inter Variable/Inter.ttf"
	InterBoldFont    = "Inter/Inter Hinted for Windows/Desktop/Inter-Bold.ttf"
)

// Renderer defines the interface for invoice rendering
type Renderer interface {
	Render(invoice *models.Invoice, w io.Writer) error
	RenderToFile(invoice *models.Invoice, filePath string) error
}

// PDFRenderer implements the Renderer interface for PDF output
type PDFRenderer struct {
	currencyService currency.Service
}

// NewPDFRenderer creates a new PDFRenderer instance
func NewPDFRenderer(currencyService currency.Service) *PDFRenderer {
	return &PDFRenderer{
		currencyService: currencyService,
	}
}

// Render renders an invoice as PDF and writes it to the provided writer
func (r *PDFRenderer) Render(invoice *models.Invoice, w io.Writer) error {
	pdf := r.createPDF()
	
	if err := r.setupFonts(pdf); err != nil {
		return err
	}
	
	// Combine ID and IdSuffix for the full invoice number
	fullInvoiceId := invoice.Id
	if invoice.IdSuffix != "" {
		fullInvoiceId = invoice.Id + invoice.IdSuffix
	}
	
	// Generate the content
	r.writeLogo(pdf, invoice.Logo, invoice.From)
	r.writeTitle(pdf, invoice.Title, fullInvoiceId, invoice.Date)
	r.writeBillTo(pdf, invoice.To)
	r.writeHeaderRow(pdf)
	
	subtotal := 0.0
	if len(invoice.Items) > 0 {
		for i := range invoice.Items {
			q := 1
			if len(invoice.Quantities) > i {
				q = invoice.Quantities[i]
			}
			
			rate := 0.0
			if len(invoice.Rates) > i {
				rate = invoice.Rates[i]
			}
			
			r.writeRow(pdf, invoice.Items[i], q, rate, invoice.Currency)
			subtotal += float64(q) * rate
		}
	}
	
	// Write notes first before totals
	if invoice.Note != "" {
		r.writeNotes(pdf, invoice.Note)
	}
	
	// Then write totals (will be positioned on the right side)
	r.writeTotals(pdf, subtotal, subtotal*invoice.Tax, subtotal*invoice.Discount, invoice.TaxExempt, invoice.Currency)
	
	if invoice.Due != "" {
		r.writeDueDate(pdf, invoice.Due)
	}
	
	r.writeFooter(pdf, fullInvoiceId, invoice.Footer)
	
	// Write the PDF bytes to the provided writer
	return pdf.Write(w)
}

// RenderToFile renders an invoice as PDF and saves it to the provided file path
func (r *PDFRenderer) RenderToFile(invoice *models.Invoice, filePath string) error {
	pdf := r.createPDF()
	
	if err := r.setupFonts(pdf); err != nil {
		return err
	}
	
	// Combine ID and IdSuffix for the full invoice number
	fullInvoiceId := invoice.Id
	if invoice.IdSuffix != "" {
		fullInvoiceId = invoice.Id + invoice.IdSuffix
	}
	
	// Generate the content
	r.writeLogo(pdf, invoice.Logo, invoice.From)
	r.writeTitle(pdf, invoice.Title, fullInvoiceId, invoice.Date)
	r.writeBillTo(pdf, invoice.To)
	r.writeHeaderRow(pdf)
	
	subtotal := 0.0
	if len(invoice.Items) > 0 {
		for i := range invoice.Items {
			q := 1
			if len(invoice.Quantities) > i {
				q = invoice.Quantities[i]
			}
			
			rate := 0.0
			if len(invoice.Rates) > i {
				rate = invoice.Rates[i]
			}
			
			r.writeRow(pdf, invoice.Items[i], q, rate, invoice.Currency)
			subtotal += float64(q) * rate
		}
	}
	
	// Write notes first before totals
	if invoice.Note != "" {
		r.writeNotes(pdf, invoice.Note)
	}
	
	// Then write totals (will be positioned on the right side)
	r.writeTotals(pdf, subtotal, subtotal*invoice.Tax, subtotal*invoice.Discount, invoice.TaxExempt, invoice.Currency)
	
	if invoice.Due != "" {
		r.writeDueDate(pdf, invoice.Due)
	}
	
	r.writeFooter(pdf, fullInvoiceId, invoice.Footer)
	
	// Write the PDF to the file
	return pdf.WritePdf(filePath)
}

// createPDF initializes a new GoPdf instance with correct page setup
func (r *PDFRenderer) createPDF() *gopdf.GoPdf {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})
	pdf.SetMargins(40, 40, 40, 40)
	pdf.AddPage()
	return pdf
}

// setupFonts loads the required fonts for the PDF
func (r *PDFRenderer) setupFonts(pdf *gopdf.GoPdf) error {
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
	
	return nil
}

// writeLogo adds the company logo and name to the PDF
func (r *PDFRenderer) writeLogo(pdf *gopdf.GoPdf, logo string, from string) {
	if logo != "" {
		width, height := r.getImageDimension(logo)
		
		// Increase the logo size
		scaledWidth := 150.0  // Increased from 100.0
		scaledHeight := float64(height) * scaledWidth / float64(width)
		
		// Set a reasonable maximum height while allowing larger logos
		maxHeight := 100.0  // Increased from 60.0
		
		// If logo is too tall, rescale it to the maximum height
		if scaledHeight > maxHeight {
			scaledHeight = maxHeight
			scaledWidth = float64(width) * maxHeight / float64(height)
		}
		
		err := pdf.Image(logo, pdf.GetX(), pdf.GetY(), &gopdf.Rect{W: scaledWidth, H: scaledHeight})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Unable to add logo to PDF: %v\n", err)
		} else {
			pdf.Br(scaledHeight + 10) // Space after logo
		}
	}
	
	pdf.SetTextColor(55, 55, 55)
	
	formattedFrom := strings.ReplaceAll(from, `\n`, "\n")
	fromLines := strings.Split(formattedFrom, "\n")
	
	for i := 0; i < len(fromLines); i++ {
		if i == 0 {
			_ = pdf.SetFont("Inter", "", 12)
			_ = pdf.Cell(nil, fromLines[i])
			pdf.Br(14)
		} else {
			_ = pdf.SetFont("Inter", "", 10)
			_ = pdf.Cell(nil, fromLines[i])
			pdf.Br(12)
		}
	}
	
	pdf.Br(15)
	pdf.SetStrokeColor(225, 225, 225)
	pdf.Line(pdf.GetX(), pdf.GetY(), 260, pdf.GetY())
	pdf.Br(20)
}

// writeTitle adds the invoice title and ID to the PDF
func (r *PDFRenderer) writeTitle(pdf *gopdf.GoPdf, title, id, date string) {
	_ = pdf.SetFont("Inter-Bold", "", 22)  // Slightly smaller font
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.Cell(nil, title)
	pdf.Br(24) // Reduced space
	_ = pdf.SetFont("Inter", "", 11) // Slightly smaller font
	pdf.SetTextColor(100, 100, 100)
	_ = pdf.Cell(nil, "#")
	_ = pdf.Cell(nil, id)
	pdf.SetTextColor(150, 150, 150)
	_ = pdf.Cell(nil, "  ·  ")
	pdf.SetTextColor(100, 100, 100)
	_ = pdf.Cell(nil, date)
	pdf.Br(32) // Reduced space
}

// writeDueDate adds the payment due date to the PDF
func (r *PDFRenderer) writeDueDate(pdf *gopdf.GoPdf, due string) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(350) // Fixed position for label
	_ = pdf.Cell(nil, dueDateLabel)
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.SetFontSize(11)
	pdf.SetX(470) // Fixed position for value
	_ = pdf.Cell(nil, due)
	pdf.Br(12)
}

// writeBillTo adds the recipient information to the PDF
func (r *PDFRenderer) writeBillTo(pdf *gopdf.GoPdf, to string) {
	pdf.SetTextColor(75, 75, 75)
	_ = pdf.SetFont("Inter", "", 9)
	_ = pdf.Cell(nil, billToLabel)
	pdf.Br(12) // Reduced space
	pdf.SetTextColor(75, 75, 75)
	
	formattedTo := strings.ReplaceAll(to, `\n`, "\n")
	toLines := strings.Split(formattedTo, "\n")
	
	for i := 0; i < len(toLines); i++ {
		if i == 0 {
			_ = pdf.SetFont("Inter", "", 15)
			_ = pdf.Cell(nil, toLines[i])
			pdf.Br(16) // Reduced space
		} else {
			_ = pdf.SetFont("Inter", "", 10)
			_ = pdf.Cell(nil, toLines[i])
			pdf.Br(12) // Reduced space
		}
	}
	pdf.Br(30) // Reduced space
}

// writeHeaderRow adds the column headers for invoice items to the PDF
func (r *PDFRenderer) writeHeaderRow(pdf *gopdf.GoPdf) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(55, 55, 55)
	_ = pdf.Cell(nil, itemLabel)
	pdf.SetX(quantityColumnOffset)
	_ = pdf.Cell(nil, qtyLabel)
	pdf.SetX(rateColumnOffset)
	_ = pdf.Cell(nil, rateLabel)
	pdf.SetX(amountColumnOffset)
	_ = pdf.Cell(nil, amountLabel)
	pdf.Br(24)
}

// writeMultilineText draws text with word wrapping
func (r *PDFRenderer) writeMultilineText(pdf *gopdf.GoPdf, text string, x, y, width float64, lineHeight float64) float64 {
	pdf.SetX(x)
	pdf.SetY(y)
	
	words := strings.Fields(text)
	currentLine := ""
	
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word
		
		// Measure the width of the test line
		textWidth, err := pdf.MeasureTextWidth(testLine)
		if err != nil {
			textWidth = float64(len(testLine) * 5) // rough estimate
		}
		
		// If adding the word exceeds available width, write the current line and start a new one
		if textWidth > width && currentLine != "" {
			pdf.SetX(x)
			_ = pdf.Cell(nil, currentLine)
			pdf.Br(lineHeight)
			currentLine = word
		} else {
			// Add the word to the current line
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}
	
	// Write the last line if any
	if currentLine != "" {
		pdf.SetX(x)
		_ = pdf.Cell(nil, currentLine)
		pdf.Br(lineHeight)
	}
	
	// Return the new Y position after writing all the text
	return pdf.GetY()
}

// writeNotes adds notes to the PDF
func (r *PDFRenderer) writeNotes(pdf *gopdf.GoPdf, notes string) {
	// Get the current Y position after writing all the invoice items
	currentY := pdf.GetY()
	
	// Add spacing after the items (reduced)
	currentY += 15
	
	// Set position for notes section
	pdf.SetY(currentY)
	
	// Write the "NOTES" header
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(55, 55, 55)
	_ = pdf.Cell(nil, notesLabel)
	pdf.Br(12) // Reduced space
	
	// Configure for the notes content
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(0, 0, 0)
	
	// Available width for text (leaving space for the totals column)
	availableWidth := 320.0
	
	// Format notes text
	formattedNotes := strings.ReplaceAll(notes, `\n`, "\n")
	
	// Write the notes with word wrapping
	r.writeMultilineText(pdf, formattedNotes, pdf.GetX(), pdf.GetY(), availableWidth, 12) // Reduced line height
}

// writeFooter adds the footer information to the PDF
func (r *PDFRenderer) writeFooter(pdf *gopdf.GoPdf, id string, footer models.Footer) {
	// Set position for footer - moved higher up the page
	pdf.SetY(770)
	
	// Add a line above the footer
	pdf.SetStrokeColor(225, 225, 225)
	pdf.Line(40, pdf.GetY(), 550, pdf.GetY())
	pdf.Br(15)
	
	// Set font for footer text
	_ = pdf.SetFont("Inter", "", 8)
	pdf.SetTextColor(75, 75, 75)
	
	// Define column widths and positions
	leftColX := 40.0
	leftColWidth := 150.0
	
	middleColX := 215.0
	middleColWidth := 160.0
	
	rightColX := 400.0
	
	lineHeight := 10.0 // Space between lines
	
	// Column 1 - Left
	currentY := pdf.GetY()
	startY := currentY
	
	// Company name
	pdf.SetX(leftColX)
	_ = pdf.Cell(nil, footer.CompanyName)
	pdf.Br(lineHeight)
	
	// Registration info - only if it should be shown
	if footer.ShowRegistration && footer.RegistrationInfo != "" {
		pdf.SetX(leftColX)
		formattedRegInfo := strings.ReplaceAll(footer.RegistrationInfo, `\n`, "\n")
		if strings.Contains(formattedRegInfo, "\n") {
			// If it contains newlines, use multiline text
			currentY = pdf.GetY()
			newY := r.writeMultilineText(pdf, formattedRegInfo, leftColX, currentY, leftColWidth, lineHeight)
			// Set Y position after multiline text
			pdf.SetY(newY)
		} else {
			// Single line
			_ = pdf.Cell(nil, formattedRegInfo)
			pdf.Br(lineHeight)
		}
	}
	
	// VAT ID - only if it should be shown
	if footer.ShowVatId && footer.VatId != "" {
		pdf.SetX(leftColX)
		_ = pdf.Cell(nil, footer.VatId)
	}
	
	// Column 2 - Middle
	pdf.SetY(startY)
	currentY = startY
	
	// Address
	pdf.SetX(middleColX)
	_ = pdf.Cell(nil, footer.Address)
	pdf.Br(lineHeight)
	
	// Zip and City
	pdf.SetX(middleColX)
	zipCity := footer.Zip
	if zipCity != "" && footer.City != "" {
		zipCity += " " + footer.City
	} else if footer.City != "" {
		zipCity = footer.City
	}
	_ = pdf.Cell(nil, zipCity)
	pdf.Br(lineHeight)
	
	// Phone
	pdf.SetX(middleColX)
	if footer.Phone != "" {
		_ = pdf.Cell(nil, "Tel.: " + footer.Phone)
	}
	pdf.Br(lineHeight)
	
	// Email and Website
	pdf.SetX(middleColX)
	contactInfo := ""
	if footer.Email != "" {
		contactInfo = footer.Email
		if footer.Website != "" {
			contactInfo += " | " + footer.Website
		}
	} else if footer.Website != "" {
		contactInfo = footer.Website
	}
	
	// Check if contact info is long and needs wrapping
	if len(contactInfo) > 30 {
		currentY = pdf.GetY()
		r.writeMultilineText(pdf, contactInfo, middleColX, currentY, middleColWidth, lineHeight)
	} else {
		_ = pdf.Cell(nil, contactInfo)
	}
	
	// Column 3 - Right
	pdf.SetY(startY)
	
	// Bank header
	pdf.SetX(rightColX)
	_ = pdf.Cell(nil, "Bankverbindung:")
	pdf.Br(lineHeight)
	
	// Bank name
	pdf.SetX(rightColX)
	_ = pdf.Cell(nil, footer.BankName)
	pdf.Br(lineHeight)
	
	// IBAN
	pdf.SetX(rightColX)
	if footer.BankIban != "" {
		_ = pdf.Cell(nil, "IBAN: " + footer.BankIban)
	}
	pdf.Br(lineHeight)
	
	// BIC
	pdf.SetX(rightColX)
	if footer.BankBic != "" {
		_ = pdf.Cell(nil, "BIC: " + footer.BankBic)
	}
	
	// Add invoice number at the top of the page
	pdf.SetY(25)
	pdf.SetX(500)
	_ = pdf.Cell(nil, id + " · " + "1/1")
}

// writeRow adds an invoice item row to the PDF
func (r *PDFRenderer) writeRow(pdf *gopdf.GoPdf, item string, quantity int, rate float64, currency string) {
	_ = pdf.SetFont("Inter", "", 10) // Slightly smaller font
	pdf.SetTextColor(0, 0, 0)
	
	total := float64(quantity) * rate
	amount := strconv.FormatFloat(total, 'f', 2, 64)
	
	// For article/description column, use text wrapping if it's too long
	if len(item) > 40 {
		availableWidth := float64(quantityColumnOffset - 60)
		r.writeMultilineText(pdf, item, pdf.GetX(), pdf.GetY(), availableWidth, 12) // Reduced line height
		// Reset Y position for quantity, rate, and amount
		pdf.SetY(pdf.GetY() - 12)
	} else {
		_ = pdf.Cell(nil, item)
	}
	
	// Get currency symbol from the service
	currencySymbol := r.currencyService.GetSymbol(currency)
	
	pdf.SetX(quantityColumnOffset)
	_ = pdf.Cell(nil, strconv.Itoa(quantity))
	pdf.SetX(rateColumnOffset)
	_ = pdf.Cell(nil, currencySymbol+strconv.FormatFloat(rate, 'f', 2, 64))
	pdf.SetX(amountColumnOffset)
	_ = pdf.Cell(nil, currencySymbol+amount)
	pdf.Br(20) // Reduced row spacing
}

// writeTotals adds the invoice totals to the PDF
func (r *PDFRenderer) writeTotals(pdf *gopdf.GoPdf, subtotal float64, tax float64, discount float64, taxExempt bool, currency string) {
	// Get the current Y position - use dynamic positioning instead of fixed position
	currentY := pdf.GetY() + 20
	
	// Set X position for the totals section (using absolute positioning)
	pdf.SetX(350) // Fixed position for labels
	pdf.SetY(currentY)
	
	// Get currency symbol from the service
	currencySymbol := r.currencyService.GetSymbol(currency)
	
	r.writeTotal(pdf, subtotalLabel, subtotal, currencySymbol)
	
	// Only show tax if not exempt
	if !taxExempt && tax > 0 {
		r.writeTotal(pdf, taxLabel, tax, currencySymbol)
	} else if taxExempt {
		// Add a note about tax exemption (Kleinunternehmer-Regelung)
		pdf.SetX(350)
		_ = pdf.SetFont("Inter", "", 9)
		pdf.SetTextColor(75, 75, 75)
		_ = pdf.Cell(nil, "Gemäß § 19 UStG wird keine Umsatzsteuer berechnet.")
		pdf.Br(24)
	}
	
	if discount > 0 {
		r.writeTotal(pdf, discountLabel, discount, currencySymbol)
	}
	
	// Calculate total - only add tax if not exempt
	total := subtotal - discount
	if !taxExempt {
		total += tax
	}
	
	r.writeTotal(pdf, totalLabel, total, currencySymbol)
}

// writeTotal adds a single total line to the PDF
func (r *PDFRenderer) writeTotal(pdf *gopdf.GoPdf, label string, total float64, currencySymbol string) {
	_ = pdf.SetFont("Inter", "", 9)
	pdf.SetTextColor(75, 75, 75)
	pdf.SetX(350) // Fixed position for labels
	_ = pdf.Cell(nil, label)
	pdf.SetTextColor(0, 0, 0)
	_ = pdf.SetFontSize(12)
	pdf.SetX(470) // Fixed position for values
	if label == totalLabel {
		_ = pdf.SetFont("Inter-Bold", "", 11.5)
	}
	_ = pdf.Cell(nil, currencySymbol+strconv.FormatFloat(total, 'f', 2, 64))
	pdf.Br(24)
}

// getImageDimension returns the width and height of an image
func (r *PDFRenderer) getImageDimension(imagePath string) (int, int) {
	// If image path is empty, return zero dimensions
	if imagePath == "" {
		return 0, 0
	}
	
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening image %s: %v\n", imagePath, err)
		return 0, 0
	}
	defer file.Close()
	
	image, _, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding image %s: %v\n", imagePath, err)
		return 0, 0
	}
	return image.Width, image.Height
}