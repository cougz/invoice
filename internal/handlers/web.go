package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	
	"invoice/internal/models"
	"invoice/internal/services/invoice"
	
	"github.com/gin-gonic/gin"
)

// WebHandler handles web interface requests
type WebHandler struct {
	invoiceService   invoice.Service
	configLoader     models.ConfigLoader
	webConfig        models.WebConfig
	htmlTemplateText string
}

// NewWebHandler creates a new WebHandler instance
func NewWebHandler(
	invoiceService invoice.Service,
	configLoader models.ConfigLoader,
	webConfig models.WebConfig,
	htmlTemplateText string,
) *WebHandler {
	return &WebHandler{
		invoiceService:   invoiceService,
		configLoader:     configLoader,
		webConfig:        webConfig,
		htmlTemplateText: htmlTemplateText,
	}
}

// RegisterRoutes registers all web routes to the provided router
func (h *WebHandler) RegisterRoutes(router *gin.Engine) {
	// Serve static files
	router.Static("/static", "./web/static")
	
	// API routes
	api := router.Group("/api")
	{
		// Generate invoice
		api.POST("/generate", h.handleGenerateInvoice)
		
		// List available configuration files
		api.GET("/config-files", h.handleListConfigFiles)
		
		// Get config file data for pre-filling form
		api.GET("/config-data/:filename", h.handleGetConfigData)
		
		// View generated PDF
		api.GET("/view/:filename", h.handleViewPDF)
		
		// Download generated PDF
		api.GET("/download/:filename", h.handleDownloadPDF)
		
		// Upload to Nextcloud
		api.POST("/upload/:filename", h.handleUpload)
	}
	
	// Handle index route - serve the HTML template directly
	router.GET("/", h.handleIndex)
}

// handleIndex serves the index page
func (h *WebHandler) handleIndex(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, h.htmlTemplateText)
}

// handleGenerateInvoice generates an invoice from a web request
func (h *WebHandler) handleGenerateInvoice(c *gin.Context) {
	var request models.InvoiceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request data"})
		return
	}
	
	// Parse the request into generate options
	options, err := h.invoiceService.ParseRequest(&request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false, 
			"message": "Failed to parse request: " + err.Error(),
		})
		return
	}
	
	// Generate the invoice
	filename, err := h.invoiceService.Generate(options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false, 
			"message": "Failed to generate invoice: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"filename": filename,
	})
}

// handleListConfigFiles lists all available config files
func (h *WebHandler) handleListConfigFiles(c *gin.Context) {
	files, err := h.findConfigFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "files": files})
}

// handleGetConfigData returns the data from a config file
func (h *WebHandler) handleGetConfigData(c *gin.Context) {
	filename := c.Param("filename")
	configData, err := h.getConfigData(filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": configData})
}

// handleViewPDF serves a PDF file for viewing
func (h *WebHandler) handleViewPDF(c *gin.Context) {
	filename := c.Param("filename")
	c.File(filename)
}

// handleDownloadPDF serves a PDF file for download
func (h *WebHandler) handleDownloadPDF(c *gin.Context) {
	filename := c.Param("filename")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.File(filename)
}

// handleUpload handles file upload to Nextcloud
func (h *WebHandler) handleUpload(c *gin.Context) {
	filename := c.Param("filename")
	result, err := h.uploadToNextcloud(filename, h.webConfig.UploadScript, h.webConfig.NextcloudURL, h.webConfig.NextcloudShare)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Upload failed: " + err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, result)
}

// findConfigFiles returns a list of JSON and YAML config files
func (h *WebHandler) findConfigFiles() ([]string, error) {
	var files []string
	
	// Find JSON and YAML files in the config directory
	configDir := "config"
	jsonFiles, err := filepath.Glob(filepath.Join(configDir, "*.json"))
	if err != nil {
		return nil, err
	}
	ymlFiles, err := filepath.Glob(filepath.Join(configDir, "*.yml"))
	if err != nil {
		return nil, err
	}
	yamlFiles, err := filepath.Glob(filepath.Join(configDir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	
	// Merge all files
	files = append(files, jsonFiles...)
	files = append(files, ymlFiles...)
	files = append(files, yamlFiles...)
	
	// Filter out non-invoice config files
	var configFiles []string
	for _, file := range files {
		// Skip known non-invoice config files
		basename := filepath.Base(file)
		if basename == "currency.json" || basename == "web_config.json" {
			continue
		}
		configFiles = append(configFiles, file)
	}
	
	return configFiles, nil
}

// getConfigData gets the data from a config file
func (h *WebHandler) getConfigData(filename string) (map[string]interface{}, error) {
	// Ensure we're looking in the config directory
	if filepath.Dir(filename) == "." {
		filename = filepath.Join("config", filename)
	}
	
	// Read the file
	fileText, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}
	
	// Remove UTF-8 BOM if present
	if len(fileText) >= 3 && fileText[0] == 0xEF && fileText[1] == 0xBB && fileText[2] == 0xBF {
		fileText = fileText[3:]
	}
	
	// Create a map to hold the data
	var configData map[string]interface{}
	
	// Check file type and parse accordingly
	if strings.HasSuffix(filename, ".json") {
		err = json.Unmarshal(fileText, &configData)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON: %v", err)
		}
		
		// Ensure tax exemption is properly reflected in the UI
		// If taxExempt is true, ensure tax is set to 0
		if taxExempt, ok := configData["taxExempt"].(bool); ok && taxExempt {
			configData["tax"] = 0
		}
		
	} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return nil, fmt.Errorf("YAML files not supported for web interface preview")
	} else {
		return nil, fmt.Errorf("unsupported file type: only .json is supported for preview")
	}
	
	return configData, nil
}

// uploadToNextcloud uploads a file to Nextcloud using the provided script
func (h *WebHandler) uploadToNextcloud(filename, scriptPath, nextcloudURL, shareID string) (models.UploadResult, error) {
	result := models.UploadResult{
		Success: false,
	}
	
	// Check if the upload script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return result, fmt.Errorf("upload script not found: %s", scriptPath)
	}
	
	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return result, fmt.Errorf("file not found: %s", filename)
	}
	
	// Construct the share URL
	shareURL := nextcloudURL + shareID
	
	// Run the upload script
	cmd := exec.Command(scriptPath, filename, shareURL)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	if err != nil {
		return result, fmt.Errorf("upload failed: %v\nStderr: %s", err, stderr.String())
	}
	
	// Format the correct Nextcloud share URL
	// This creates a URL like: https://cloud.seiffert.me/index.php/s/CAr4Gfs9NFd9RqG?path=&files=filename.pdf
	formattedURL := fmt.Sprintf("%s?path=&files=%s", shareURL, filename)
	
	result.Success = true
	result.URL = formattedURL
	result.Message = "File uploaded successfully"
	
	return result, nil
}
