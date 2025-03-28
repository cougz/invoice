# Strategic Code Improvement Recommendations

This document outlines strategic recommendations for improving the invoice generator codebase based on analysis of its current structure and implementation needs.

## 1. CODE ORGANIZATION AND ARCHITECTURE

### Recommended Package Structure
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
│   │   ├── generate.go       # Invoice generation logic
│   │   └── web.go            # Web interface handlers
│   ├── services/             # Business logic
│   │   ├── pdf/              # PDF generation
│   │   │   └── renderer.go   # PDF rendering logic
│   │   ├── currency/         # Currency handling
│   │   └── storage/          # Data persistence
│   └── config/               # Config handling
│       └── loader.go         # Configuration loading
├── pkg/                      # Public libraries
│   └── templates/            # Invoice templates
└── web/                      # Web assets and templates
    ├── templates/            # HTML templates
    └── static/               # Static assets
```

### Interface Design
Create interfaces to decouple dependencies and improve testability:

```go
// Renderer defines the interface for invoice rendering services
type Renderer interface {
    Render(invoice *models.Invoice, w io.Writer) error
}

// Storage defines the interface for invoice storage
type Storage interface {
    Save(invoice *models.Invoice) (string, error)
    Get(id string) (*models.Invoice, error)
    List() ([]*models.Invoice, error)
}

// ConfigLoader defines the interface for loading configurations
type ConfigLoader interface {
    Load(path string) (*models.Config, error)
    Save(config *models.Config, path string) error
}
```

### Business Logic Extraction
Extract core business logic from handlers to improve testability:

- Move PDF generation from command handlers to a dedicated service
- Extract invoice validation to a separate service
- Implement a InvoiceService to encapsulate all invoice operations

## 2. CONFIGURATION MANAGEMENT

### Unified Configuration System

Create a unified configuration system with the following hierarchy:

1. Default values (hard-coded)
2. Configuration files (local JSON/YAML)
3. Environment variables
4. Command-line flags

```go
// Configuration management using viper
func initConfig() {
    viper.SetDefault("port", 8080)
    viper.SetDefault("tax.rate", 0.19)
    viper.SetDefault("tax.exempt", false)
    
    // Load from file
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    
    // Load environment variables
    viper.AutomaticEnv()
    viper.SetEnvPrefix("INVOICE")
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    
    // Load command line flags
    pflag.CommandLine.AddFlagSet(rootCmd.PersistentFlags())
    pflag.Parse()
    viper.BindPFlags(pflag.CommandLine)
    
    // Read config
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            log.Fatalf("Error reading config file: %v", err)
        }
    }
}
```

### Environment Variable Integration

Map key configuration options to environment variables:

| Environment Variable     | Configuration Key       | Description                          |
|--------------------------|-------------------------|--------------------------------------|
| INVOICE_PORT             | port                    | Web server port                      |
| INVOICE_TAX_RATE         | tax.rate                | Default tax rate                     |
| INVOICE_TAX_EXEMPT       | tax.exempt              | Tax exemption (Kleinunternehmer)     |
| INVOICE_FOOTER_COMPANY   | footer.company_name     | Company name in footer               |
| INVOICE_FOOTER_SHOW_VAT  | footer.show_vat_id      | Whether to show VAT ID in footer     |
| INVOICE_NEXTCLOUD_URL    | nextcloud.url           | Nextcloud server URL                 |
| INVOICE_NEXTCLOUD_SHARE  | nextcloud.share         | Nextcloud share ID                   |

## 3. TEMPLATE HANDLING

### Implement HTML Template System

Move templates to separate files and implement a template cache:

```go
// TemplateRenderer manages HTML templates
type TemplateRenderer struct {
    templates map[string]*template.Template
    baseDir   string
    cache     bool
}

func NewTemplateRenderer(baseDir string, cache bool) *TemplateRenderer {
    return &TemplateRenderer{
        templates: make(map[string]*template.Template),
        baseDir:   baseDir,
        cache:     cache,
    }
}

func (tr *TemplateRenderer) Render(w http.ResponseWriter, name string, data interface{}) error {
    // Check cache first
    if tr.cache {
        if tmpl, ok := tr.templates[name]; ok {
            return tmpl.ExecuteTemplate(w, "base", data)
        }
    }
    
    // Load template
    tmpl, err := template.New("base").ParseFiles(
        filepath.Join(tr.baseDir, "base.html"),
        filepath.Join(tr.baseDir, name),
    )
    if err != nil {
        return err
    }
    
    // Cache template
    if tr.cache {
        tr.templates[name] = tmpl
    }
    
    return tmpl.ExecuteTemplate(w, "base", data)
}
```

### Separate Templates

Create the following templates:

- `base.html` - Base layout template
- `index.html` - Main form template
- `result.html` - Invoice result template
- `settings.html` - Configuration settings template

### Customizable Invoice Templates

Implement a template system for PDF invoices:

```go
// InvoiceTemplate defines a template for invoice generation
type InvoiceTemplate struct {
    Name        string
    Description string
    Stylesheet  string
    Layout      string
    Header      string
    Items       string
    Footer      string
}

// TemplateRegistry manages available invoice templates
type TemplateRegistry struct {
    templates map[string]*InvoiceTemplate
}

func (tr *TemplateRegistry) Register(template *InvoiceTemplate) {
    tr.templates[template.Name] = template
}

func (tr *TemplateRegistry) Get(name string) (*InvoiceTemplate, error) {
    if template, ok := tr.templates[name]; ok {
        return template, nil
    }
    return nil, fmt.Errorf("template not found: %s", name)
}
```

## 4. RESTFUL API DESIGN

### API Endpoints

Implement a RESTful API with proper versioning:

```
/api/v1/invoices          GET    - List invoices
/api/v1/invoices          POST   - Create invoice
/api/v1/invoices/:id      GET    - Get invoice details
/api/v1/invoices/:id      PUT    - Update invoice
/api/v1/invoices/:id      DELETE - Delete invoice
/api/v1/invoices/:id/pdf  GET    - Get invoice PDF
/api/v1/config            GET    - Get configuration
/api/v1/config            PUT    - Update configuration
```

### Request/Response Models

Define proper request and response models:

```go
// InvoiceCreateRequest defines the request body for creating an invoice
type InvoiceCreateRequest struct {
    From           string   `json:"from" binding:"required"`
    To             string   `json:"to" binding:"required"`
    Items          []string `json:"items" binding:"required,min=1"`
    Quantities     []int    `json:"quantities" binding:"required,min=1,len=Items"`
    Rates          []float64 `json:"rates" binding:"required,min=1,len=Items"`
    Tax            float64  `json:"tax"`
    TaxExempt      bool     `json:"taxExempt"`
    Discount       float64  `json:"discount"`
    Currency       string   `json:"currency" binding:"required"`
    Note           string   `json:"note"`
    Id             string   `json:"id"`
    IdSuffix       string   `json:"idSuffix"`
    Footer         *FooterSettings `json:"footer"`
}

// InvoiceResponse defines the response for invoice operations
type InvoiceResponse struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
    Filename  string    `json:"filename"`
    URL       string    `json:"url,omitempty"`
    // Include other invoice fields
}

// ErrorResponse defines the error response format
type ErrorResponse struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}
```

### Validation

Implement proper request validation:

```go
func validateInvoiceRequest(req *InvoiceCreateRequest) error {
    if req.From == "" {
        return errors.New("from field is required")
    }
    if req.To == "" {
        return errors.New("to field is required")
    }
    if len(req.Items) == 0 {
        return errors.New("at least one item is required")
    }
    if len(req.Items) != len(req.Quantities) || len(req.Items) != len(req.Rates) {
        return errors.New("items, quantities, and rates must have the same length")
    }
    
    for i, rate := range req.Rates {
        if rate < 0 {
            return fmt.Errorf("rate for item %d cannot be negative", i+1)
        }
    }
    
    for i, qty := range req.Quantities {
        if qty <= 0 {
            return fmt.Errorf("quantity for item %d must be greater than zero", i+1)
        }
    }
    
    return nil
}
```

## 5. TESTING STRATEGY

### Unit Tests

Implement unit tests for core business logic:

```go
func TestCalculateInvoiceTotal(t *testing.T) {
    tests := []struct {
        name     string
        items    []float64
        tax      float64
        taxExempt bool
        discount float64
        expected float64
    }{
        {
            name:     "Regular invoice with tax",
            items:    []float64{100, 200, 300},
            tax:      0.19,
            taxExempt: false,
            discount: 0,
            expected: 714.00, // 600 + (600 * 0.19)
        },
        {
            name:     "Tax exempt invoice",
            items:    []float64{100, 200, 300},
            tax:      0.19,
            taxExempt: true,
            discount: 0,
            expected: 600.00, // No tax applied
        },
        {
            name:     "Invoice with discount",
            items:    []float64{100, 200, 300},
            tax:      0.19,
            taxExempt: false,
            discount: 0.1,
            expected: 642.60, // (600 - 60) + ((600 - 60) * 0.19)
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            invoice := &Invoice{
                Rates:      make([]float64, len(tt.items)),
                Quantities: make([]int, len(tt.items)),
                Tax:        tt.tax,
                TaxExempt:  tt.taxExempt,
                Discount:   tt.discount,
            }
            
            for i, item := range tt.items {
                invoice.Rates[i] = item
                invoice.Quantities[i] = 1
            }
            
            total := calculateTotal(invoice)
            if math.Abs(total-tt.expected) > 0.01 {
                t.Errorf("Expected total %f, got %f", tt.expected, total)
            }
        })
    }
}
```

### Integration Tests

Implement integration tests for API endpoints:

```go
func TestGenerateInvoiceAPI(t *testing.T) {
    // Setup test server
    router := setupRouter()
    server := httptest.NewServer(router)
    defer server.Close()
    
    // Test data
    reqBody := &InvoiceCreateRequest{
        From:       "Test Company",
        To:         "Test Client",
        Items:      []string{"Item 1", "Item 2"},
        Quantities: []int{1, 2},
        Rates:      []float64{100, 50},
        Tax:        0.19,
        Currency:   "EUR",
    }
    
    // Convert to JSON
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        t.Fatalf("Failed to marshal request: %v", err)
    }
    
    // Send request
    resp, err := http.Post(server.URL+"/api/v1/invoices", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        t.Fatalf("Failed to send request: %v", err)
    }
    defer resp.Body.Close()
    
    // Check response
    if resp.StatusCode != http.StatusCreated {
        t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
    }
    
    // Parse response
    var result InvoiceResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    
    // Validate response
    if result.ID == "" {
        t.Error("Expected invoice ID, got empty string")
    }
    if result.Filename == "" {
        t.Error("Expected filename, got empty string")
    }
}
```

### End-to-End Tests

Implement end-to-end tests for the invoice generation flow:

```go
func TestInvoiceGenerationE2E(t *testing.T) {
    // Skip in short mode
    if testing.Short() {
        t.Skip("Skipping end-to-end test in short mode")
    }
    
    // Setup temp dir for output
    tempDir, err := os.MkdirTemp("", "invoice-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tempDir)
    
    // Prepare test invoice
    invoice := &Invoice{
        From:       "Test Company",
        To:         "Test Client",
        Items:      []string{"Test Item"},
        Quantities: []int{1},
        Rates:      []float64{100},
        Tax:        0.19,
        Currency:   "EUR",
    }
    
    // Generate invoice
    outputPath := filepath.Join(tempDir, "invoice.pdf")
    err = GenerateInvoicePDF(invoice, outputPath)
    if err != nil {
        t.Fatalf("Failed to generate invoice: %v", err)
    }
    
    // Verify file exists
    if _, err := os.Stat(outputPath); os.IsNotExist(err) {
        t.Errorf("Expected output file %s to exist", outputPath)
    }
    
    // Verify file content
    // This would require a PDF parsing library
}
```

## 6. PDF GENERATION IMPROVEMENTS

### Alternative PDF Libraries

Replace gopdf with a more feature-rich library:

```go
// Alternative libraries:
// 1. github.com/jung-kurt/gofpdf - More features and Unicode support
// 2. github.com/unidoc/unipdf - Commercial, comprehensive PDF library
// 3. github.com/pdfcpu/pdfcpu - Robust PDF processing library

// Using gofpdf example:
func generateInvoicePDF(invoice *Invoice, outputPath string) error {
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Add Unicode font
    pdf.AddUTF8Font("Inter", "", "Inter-Regular.ttf")
    pdf.AddUTF8Font("Inter", "B", "Inter-Bold.ttf")
    
    // Set font
    pdf.SetFont("Inter", "", 12)
    
    // Draw invoice content
    writeHeader(pdf, invoice)
    writeItems(pdf, invoice)
    writeFooter(pdf, invoice)
    
    return pdf.OutputFileAndClose(outputPath)
}
```

### Customizable Invoice Templates

Implement a template-based PDF generation system:

```go
// Template-based PDF generation
type PDFTemplate struct {
    Layout     string // JSON layout description
    Stylesheet string // CSS-like styling
    Assets     map[string]string // Images, logos, etc.
}

func NewPDFRenderer(template *PDFTemplate) *PDFRenderer {
    return &PDFRenderer{
        template: template,
    }
}

func (r *PDFRenderer) Render(invoice *Invoice, w io.Writer) error {
    // Parse template layout
    layout := parseLayout(r.template.Layout)
    
    // Create PDF
    pdf := gofpdf.New("P", "mm", "A4", "")
    pdf.AddPage()
    
    // Apply stylesheet
    applyStylesheet(pdf, r.template.Stylesheet)
    
    // Render layout components
    for _, component := range layout.Components {
        switch component.Type {
        case "header":
            renderHeader(pdf, component, invoice)
        case "items":
            renderItems(pdf, component, invoice)
        case "totals":
            renderTotals(pdf, component, invoice)
        case "footer":
            renderFooter(pdf, component, invoice)
        }
    }
    
    return pdf.Output(w)
}
```

### Additional Output Formats

Support multiple output formats:

```go
// OutputFormat defines the format for invoice output
type OutputFormat string

const (
    FormatPDF  OutputFormat = "pdf"
    FormatHTML OutputFormat = "html"
    FormatJSON OutputFormat = "json"
    FormatCSV  OutputFormat = "csv"
)

// Renderer interface for multiple output formats
type Renderer interface {
    Render(invoice *Invoice, w io.Writer) error
}

// Factory to create renderers
func NewRenderer(format OutputFormat) (Renderer, error) {
    switch format {
    case FormatPDF:
        return &PDFRenderer{}, nil
    case FormatHTML:
        return &HTMLRenderer{}, nil
    case FormatJSON:
        return &JSONRenderer{}, nil
    case FormatCSV:
        return &CSVRenderer{}, nil
    default:
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
}
```

## 7. UI/UX IMPROVEMENTS

### Responsive Web Interface

Update the web interface to be fully responsive:

```html
<!-- Responsive meta tag -->
<meta name="viewport" content="width=device-width, initial-scale=1.0">

<!-- Mobile-first grid system -->
<div class="container">
    <div class="row">
        <div class="col-md-6 col-sm-12">
            <!-- Left column content -->
        </div>
        <div class="col-md-6 col-sm-12">
            <!-- Right column content -->
        </div>
    </div>
</div>

<!-- Mobile navigation -->
<nav class="navbar navbar-expand-lg navbar-light bg-light">
    <button class="navbar-toggler" type="button" data-toggle="collapse" data-target="#navbarNav">
        <span class="navbar-toggler-icon"></span>
    </button>
    <div class="collapse navbar-collapse" id="navbarNav">
        <ul class="navbar-nav">
            <li class="nav-item active">
                <a class="nav-link" href="/">Home</a>
            </li>
            <li class="nav-item">
                <a class="nav-link" href="/history">Invoice History</a>
            </li>
            <li class="nav-item">
                <a class="nav-link" href="/settings">Settings</a>
            </li>
        </ul>
    </div>
</nav>
```

### Form Validation

Implement client-side validation:

```javascript
// Form validation
const form = document.getElementById('invoice-form');
form.addEventListener('submit', function(event) {
    let isValid = true;
    
    // Reset previous validation messages
    form.querySelectorAll('.is-invalid').forEach(el => el.classList.remove('is-invalid'));
    form.querySelectorAll('.invalid-feedback').forEach(el => el.remove());
    
    // Validate required fields
    ['from', 'to'].forEach(field => {
        const input = document.getElementById(field);
        if (!input.value.trim()) {
            input.classList.add('is-invalid');
            const feedback = document.createElement('div');
            feedback.className = 'invalid-feedback';
            feedback.textContent = 'This field is required';
            input.parentNode.appendChild(feedback);
            isValid = false;
        }
    });
    
    // Validate items
    const items = document.querySelectorAll('.item-name');
    if (items.length === 0) {
        isValid = false;
        alert('At least one item is required');
    }
    
    // Validate rates (must be positive numbers)
    document.querySelectorAll('.item-rate').forEach(rate => {
        const value = parseFloat(rate.value);
        if (isNaN(value) || value <= 0) {
            rate.classList.add('is-invalid');
            const feedback = document.createElement('div');
            feedback.className = 'invalid-feedback';
            feedback.textContent = 'Rate must be a positive number';
            rate.parentNode.appendChild(feedback);
            isValid = false;
        }
    });
    
    if (!isValid) {
        event.preventDefault();
    }
});

// Real-time validation
document.querySelectorAll('input, textarea, select').forEach(input => {
    input.addEventListener('input', function() {
        if (this.classList.contains('is-invalid')) {
            this.classList.remove('is-invalid');
            const feedback = this.parentNode.querySelector('.invalid-feedback');
            if (feedback) {
                feedback.remove();
            }
        }
    });
});
```

### Preview and Autosave

Implement invoice preview and autosave functionality:

```javascript
// Preview functionality
function updatePreview() {
    const previewContainer = document.getElementById('invoice-preview');
    const data = collectFormData();
    
    // Show loading indicator
    previewContainer.innerHTML = '<div class="loading">Generating preview...</div>';
    
    // Send data to preview endpoint
    fetch('/api/preview', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })
    .then(response => response.json())
    .then(data => {
        if (data.success) {
            // Update preview iframe
            const iframe = document.createElement('iframe');
            iframe.src = data.previewUrl;
            iframe.width = '100%';
            iframe.height = '500px';
            iframe.frameBorder = '0';
            
            previewContainer.innerHTML = '';
            previewContainer.appendChild(iframe);
        } else {
            previewContainer.innerHTML = 
                '<div class="alert alert-danger">Preview failed: ' + data.message + '</div>';
        }
    })
    .catch(error => {
        previewContainer.innerHTML = 
            '<div class="alert alert-danger">Error: ' + error.message + '</div>';
    });
}

// Auto-save functionality
let autoSaveTimeout;
let lastSavedData = '';

function setupAutoSave() {
    const form = document.getElementById('invoice-form');
    const autoSaveStatus = document.getElementById('autosave-status');
    
    form.addEventListener('input', function() {
        // Clear previous timeout
        if (autoSaveTimeout) {
            clearTimeout(autoSaveTimeout);
        }
        
        // Set status to pending
        autoSaveStatus.textContent = 'Changes not saved';
        autoSaveStatus.className = 'text-warning';
        
        // Schedule autosave
        autoSaveTimeout = setTimeout(function() {
            const formData = collectFormData();
            const jsonData = JSON.stringify(formData);
            
            // Only save if data has changed
            if (jsonData !== lastSavedData) {
                // Save to localStorage
                localStorage.setItem('invoice_draft', jsonData);
                lastSavedData = jsonData;
                
                // Update status
                autoSaveStatus.textContent = 'Changes saved';
                autoSaveStatus.className = 'text-success';
                
                // Save to server if user is logged in
                if (isLoggedIn) {
                    saveToServer(formData);
                }
            }
        }, 2000); // 2 second delay
    });
    
    // Load draft on page load
    const savedDraft = localStorage.getItem('invoice_draft');
    if (savedDraft) {
        try {
            const formData = JSON.parse(savedDraft);
            fillFormWithData(formData);
            lastSavedData = savedDraft;
            autoSaveStatus.textContent = 'Draft loaded';
            autoSaveStatus.className = 'text-success';
        } catch (error) {
            console.error('Error loading draft', error);
        }
    }
}
```

## 8. DATA PERSISTENCE

### Database Schema

Implement a database schema for storing invoices:

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Companies table
CREATE TABLE companies (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    registration_info TEXT,
    vat_id VARCHAR(50),
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    zip VARCHAR(20) NOT NULL,
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(255),
    bank_name VARCHAR(100),
    bank_iban VARCHAR(50),
    bank_bic VARCHAR(20),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Invoices table
CREATE TABLE invoices (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    invoice_number VARCHAR(50) NOT NULL,
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    from_company_id INTEGER REFERENCES companies(id),
    to_company_id INTEGER REFERENCES companies(id),
    tax_rate DECIMAL(5,2) NOT NULL,
    tax_exempt BOOLEAN NOT NULL DEFAULT FALSE,
    discount_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    note TEXT,
    subtotal DECIMAL(12,2) NOT NULL,
    tax_amount DECIMAL(12,2) NOT NULL,
    discount_amount DECIMAL(12,2) NOT NULL,
    total DECIMAL(12,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    pdf_path VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Invoice items table
CREATE TABLE invoice_items (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    rate DECIMAL(12,2) NOT NULL,
    amount DECIMAL(12,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Invoice history table for tracking changes
CREATE TABLE invoice_history (
    id SERIAL PRIMARY KEY,
    invoice_id INTEGER REFERENCES invoices(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    details TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

### Storage Implementation

Implement a data persistence layer:

```go
// InvoiceRepository defines the interface for invoice storage
type InvoiceRepository interface {
    Save(ctx context.Context, invoice *Invoice) (string, error)
    Get(ctx context.Context, id string) (*Invoice, error)
    List(ctx context.Context, filter InvoiceFilter) ([]*Invoice, error)
    Update(ctx context.Context, invoice *Invoice) error
    Delete(ctx context.Context, id string) error
}

// PostgresInvoiceRepository implements InvoiceRepository for PostgreSQL
type PostgresInvoiceRepository struct {
    db *sqlx.DB
}

func NewPostgresInvoiceRepository(db *sqlx.DB) *PostgresInvoiceRepository {
    return &PostgresInvoiceRepository{
        db: db,
    }
}

func (r *PostgresInvoiceRepository) Save(ctx context.Context, invoice *Invoice) (string, error) {
    tx, err := r.db.BeginTxx(ctx, nil)
    if err != nil {
        return "", err
    }
    defer tx.Rollback()
    
    // Insert invoice record
    var id int64
    err = tx.QueryRowxContext(
        ctx,
        `INSERT INTO invoices (
            user_id, invoice_number, invoice_date, due_date,
            from_company_id, to_company_id, tax_rate, tax_exempt,
            discount_rate, currency, note, subtotal, tax_amount,
            discount_amount, total, status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
        RETURNING id`,
        invoice.UserID, invoice.InvoiceNumber, invoice.InvoiceDate, invoice.DueDate,
        invoice.FromCompanyID, invoice.ToCompanyID, invoice.TaxRate, invoice.TaxExempt,
        invoice.DiscountRate, invoice.Currency, invoice.Note, invoice.Subtotal,
        invoice.TaxAmount, invoice.DiscountAmount, invoice.Total, invoice.Status,
    ).Scan(&id)
    if err != nil {
        return "", err
    }
    
    // Insert invoice items
    for _, item := range invoice.Items {
        _, err = tx.ExecContext(
            ctx,
            `INSERT INTO invoice_items (
                invoice_id, description, quantity, rate, amount
            ) VALUES ($1, $2, $3, $4, $5)`,
            id, item.Description, item.Quantity, item.Rate, item.Amount,
        )
        if err != nil {
            return "", err
        }
    }
    
    // Log history
    _, err = tx.ExecContext(
        ctx,
        `INSERT INTO invoice_history (
            invoice_id, user_id, action, details
        ) VALUES ($1, $2, $3, $4)`,
        id, invoice.UserID, "created", "Invoice created",
    )
    if err != nil {
        return "", err
    }
    
    err = tx.Commit()
    if err != nil {
        return "", err
    }
    
    return strconv.FormatInt(id, 10), nil
}
```

### User Authentication

Implement user authentication for invoice management:

```go
// User represents a user in the system
type User struct {
    ID           string
    Email        string
    PasswordHash string
    Name         string
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// UserService handles user-related operations
type UserService interface {
    Register(ctx context.Context, email, password, name string) (string, error)
    Login(ctx context.Context, email, password string) (string, error)
    GetUser(ctx context.Context, id string) (*User, error)
}

// JWTService handles JSON Web Tokens
type JWTService struct {
    secretKey []byte
    expiry    time.Duration
}

func NewJWTService(secretKey string, expiry time.Duration) *JWTService {
    return &JWTService{
        secretKey: []byte(secretKey),
        expiry:    expiry,
    }
}

func (s *JWTService) GenerateToken(userID, email string) (string, error) {
    now := time.Now()
    claims := jwt.MapClaims{
        "sub": userID,
        "email": email,
        "iat": now.Unix(),
        "exp": now.Add(s.expiry).Unix(),
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secretKey)
}

func (s *JWTService) ValidateToken(tokenString string) (*jwt.Token, error) {
    return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.secretKey, nil
    })
}
```

### Search and History

Implement invoice search and history tracking:

```go
// InvoiceFilter defines search criteria for invoices
type InvoiceFilter struct {
    UserID       string
    FromDate     *time.Time
    ToDate       *time.Time
    MinAmount    *float64
    MaxAmount    *float64
    Status       []string
    Search       string
    SortBy       string
    SortDesc     bool
    Page         int
    PageSize     int
}

// SearchInvoices returns invoices matching the filter criteria
func (r *PostgresInvoiceRepository) List(ctx context.Context, filter InvoiceFilter) ([]*Invoice, error) {
    query := `
        SELECT i.*, c1.name as from_company_name, c2.name as to_company_name
        FROM invoices i
        LEFT JOIN companies c1 ON i.from_company_id = c1.id
        LEFT JOIN companies c2 ON i.to_company_id = c2.id
        WHERE i.user_id = $1
    `
    args := []interface{}{filter.UserID}
    argCount := 1
    
    // Apply date filters
    if filter.FromDate != nil {
        argCount++
        query += fmt.Sprintf(" AND i.invoice_date >= $%d", argCount)
        args = append(args, filter.FromDate)
    }
    if filter.ToDate != nil {
        argCount++
        query += fmt.Sprintf(" AND i.invoice_date <= $%d", argCount)
        args = append(args, filter.ToDate)
    }
    
    // Apply amount filters
    if filter.MinAmount != nil {
        argCount++
        query += fmt.Sprintf(" AND i.total >= $%d", argCount)
        args = append(args, filter.MinAmount)
    }
    if filter.MaxAmount != nil {
        argCount++
        query += fmt.Sprintf(" AND i.total <= $%d", argCount)
        args = append(args, filter.MaxAmount)
    }
    
    // Apply status filter
    if len(filter.Status) > 0 {
        placeholders := make([]string, len(filter.Status))
        for i := range filter.Status {
            argCount++
            placeholders[i] = fmt.Sprintf("$%d", argCount)
            args = append(args, filter.Status[i])
        }
        query += fmt.Sprintf(" AND i.status IN (%s)", strings.Join(placeholders, ","))
    }
    
    // Apply text search
    if filter.Search != "" {
        argCount++
        query += fmt.Sprintf(` AND (
            i.invoice_number ILIKE $%d OR
            c1.name ILIKE $%d OR
            c2.name ILIKE $%d
        )`, argCount, argCount, argCount)
        searchTerm := "%" + filter.Search + "%"
        args = append(args, searchTerm)
    }
    
    // Apply sorting
    if filter.SortBy != "" {
        query += fmt.Sprintf(" ORDER BY i.%s", filter.SortBy)
        if filter.SortDesc {
            query += " DESC"
        } else {
            query += " ASC"
        }
    } else {
        query += " ORDER BY i.invoice_date DESC"
    }
    
    // Apply pagination
    if filter.Page > 0 && filter.PageSize > 0 {
        offset := (filter.Page - 1) * filter.PageSize
        query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PageSize, offset)
    }
    
    rows, err := r.db.QueryxContext(ctx, query, args...)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var invoices []*Invoice
    for rows.Next() {
        var invoice Invoice
        if err := rows.StructScan(&invoice); err != nil {
            return nil, err
        }
        invoices = append(invoices, &invoice)
    }
    
    return invoices, nil
}
```

## Implementation Roadmap

For implementing these improvements, we recommend the following phased approach:

### Phase 1: Core Functionality (1-2 weeks)
- Reorganize code into packages
- Extract business logic from handlers
- Implement the tax exemption feature
- Implement optional footer fields
- Create unit tests for core business logic

### Phase 2: Configuration and API (2-3 weeks)
- Implement unified configuration system
- Design and implement RESTful API
- Implement proper validation
- Move templates to separate files
- Add integration tests

### Phase 3: UI and Templates (2-3 weeks)
- Make the web interface responsive
- Implement form validation
- Add preview functionality
- Implement template caching
- Create customizable invoice templates

### Phase 4: Data Persistence (3-4 weeks)
- Design and implement database schema
- Add user authentication
- Implement invoice history and search
- Add end-to-end tests
- Implement multiple output formats

### Phase 5: Advanced Features (3-4 weeks)
- Integrate with payment providers
- Add recurring invoice support
- Implement email notifications
- Add report generation
- Improve internationalization support