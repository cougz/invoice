package main

import (
        "fmt"
        "image"
        "os"
        "strconv"
        "strings"

        "github.com/signintech/gopdf"
)

// Further adjusted column positions to fix all overflow issues
const (
        quantityColumnOffset = 390
        rateColumnOffset     = 450
        amountColumnOffset   = 510
)

const (
        // German translations for labels
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

func writeLogo(pdf *gopdf.GoPdf, logo string, from string) {
        if logo != "" {
                width, height := getImageDimension(logo)

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

func writeTitle(pdf *gopdf.GoPdf, title, id, date string) {
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

func writeDueDate(pdf *gopdf.GoPdf, due string) {
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

func writeBillTo(pdf *gopdf.GoPdf, to string) {
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

func writeHeaderRow(pdf *gopdf.GoPdf) {
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

// Function to draw text with word wrapping
func writeMultilineText(pdf *gopdf.GoPdf, text string, x, y, width float64, lineHeight float64) float64 {
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

func writeNotes(pdf *gopdf.GoPdf, notes string) {
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
        writeMultilineText(pdf, formattedNotes, pdf.GetX(), pdf.GetY(), availableWidth, 12) // Reduced line height
}

func writeFooter(pdf *gopdf.GoPdf, id string) {
    // Set position for footer - moved higher up the page
    pdf.SetY(770)

    // Add a line above the footer
    pdf.SetStrokeColor(225, 225, 225)
    pdf.Line(40, pdf.GetY(), 550, pdf.GetY())
    pdf.Br(15)

    // Set font for footer text
    _ = pdf.SetFont("Inter", "", 8)
    pdf.SetTextColor(75, 75, 75)

    // Get the footer values from the invoice
    footer := file.Footer

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
            newY := writeMultilineText(pdf, formattedRegInfo, leftColX, currentY, leftColWidth, lineHeight)
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
        writeMultilineText(pdf, contactInfo, middleColX, currentY, middleColWidth, lineHeight)
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

func writeRow(pdf *gopdf.GoPdf, item string, quantity int, rate float64) {
        _ = pdf.SetFont("Inter", "", 10) // Slightly smaller font
        pdf.SetTextColor(0, 0, 0)

        total := float64(quantity) * rate
        amount := strconv.FormatFloat(total, 'f', 2, 64)

        // For article/description column, use text wrapping if it's too long
        if len(item) > 40 {
                availableWidth := float64(quantityColumnOffset - 60)
                writeMultilineText(pdf, item, pdf.GetX(), pdf.GetY(), availableWidth, 12) // Reduced line height
                // Reset Y position for quantity, rate, and amount
                pdf.SetY(pdf.GetY() - 12)
        } else {
                _ = pdf.Cell(nil, item)
        }

        // Get currency symbol safely using getCurrencySymbol function
        currencySymbol := getCurrencySymbol(file.Currency)

        pdf.SetX(quantityColumnOffset)
        _ = pdf.Cell(nil, strconv.Itoa(quantity))
        pdf.SetX(rateColumnOffset)
        _ = pdf.Cell(nil, currencySymbol+strconv.FormatFloat(rate, 'f', 2, 64))
        pdf.SetX(amountColumnOffset)
        _ = pdf.Cell(nil, currencySymbol+amount)
        pdf.Br(20) // Reduced row spacing
}

func writeTotals(pdf *gopdf.GoPdf, subtotal float64, tax float64, discount float64) {
        // Get the current Y position - use dynamic positioning instead of fixed position
        currentY := pdf.GetY() + 20

        // Set X position for the totals section (using absolute positioning)
        pdf.SetX(350) // Fixed position for labels
        pdf.SetY(currentY)

        // Get currency symbol safely using the dedicated function from currency.go
        currencySymbol := getCurrencySymbol(file.Currency)

        writeTotal(pdf, subtotalLabel, subtotal, currencySymbol)
        
        // Only show tax if not exempt
        if !file.TaxExempt && tax > 0 {
                writeTotal(pdf, taxLabel, tax, currencySymbol)
        } else if file.TaxExempt {
                // Add a note about tax exemption (Kleinunternehmer-Regelung)
                pdf.SetX(350)
                _ = pdf.SetFont("Inter", "", 9)
                pdf.SetTextColor(75, 75, 75)
                _ = pdf.Cell(nil, "Gemäß § 19 UStG wird keine Umsatzsteuer berechnet.")
                pdf.Br(24)
        }
        
        if discount > 0 {
                writeTotal(pdf, discountLabel, discount, currencySymbol)
        }
        
        // Calculate total - only add tax if not exempt
        total := subtotal - discount
        if !file.TaxExempt {
                total += tax
        }
        
        writeTotal(pdf, totalLabel, total, currencySymbol)
}

// Updated to accept currency symbol as parameter
func writeTotal(pdf *gopdf.GoPdf, label string, total float64, currencySymbol string) {
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


func getImageDimension(imagePath string) (int, int) {
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