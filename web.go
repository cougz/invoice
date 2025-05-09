package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// WebConfig holds the configuration for the web server
type WebConfig struct {
	Port           int    `json:"port"`
	NextcloudURL   string `json:"nextcloudUrl"`
	NextcloudShare string `json:"nextcloudShare"`
	UploadScript   string `json:"uploadScript"`
}

// InvoiceRequest represents the form data from the web UI
type InvoiceRequest struct {
	From            string  `json:"from"`
	To              string  `json:"to"`
	Items           string  `json:"items"`
	Quantities      string  `json:"quantities"`
	Rates           string  `json:"rates"`
	Tax             float64 `json:"tax"`
	TaxExempt       bool    `json:"taxExempt"`
	Discount        float64 `json:"discount"`
	Currency        string  `json:"currency"`
	Note            string  `json:"note"`
	Id              string  `json:"id"`
	IdSuffix        string  `json:"idSuffix"`
	ConfigFile      string  `json:"configFile"`
	UseConfig       bool    `json:"useConfig"`
	ShowRegistration bool   `json:"showRegistration"`
	ShowVatId       bool    `json:"showVatId"`
	CompanyName     string  `json:"companyName"` // Added to use in footer
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
	Message string `json:"message"`
}

// HTMLTemplates contains the HTML templates for the web UI
var HTMLTemplates = map[string]string{
	"index": `<!DOCTYPE html>
<html lang="en" data-theme="light">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Invoice Generator</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-9ndCyUaIbzAi2FUVXJi0CjmCapSmO7SnpJef0486qhLnuZ2cdeRhO02iuK6FUUVM" crossorigin="anonymous">
    <link href="/static/css/style.css" rel="stylesheet">
</head>
<body>
    <div class="container">
        <h1 class="text-center mb-4">Invoice Generator</h1>
        
	<div class="theme-switch">
	    <label for="theme-toggle">Toggle Dark Mode</label>
	    <label class="switch">
	        <input type="checkbox" id="theme-toggle">
	        <span class="slider">
	            <div class="star star_1"></div>
	            <div class="star star_2"></div>
	            <div class="star star_3"></div>
	            <svg class="cloud" viewBox="0 0 100 100">
	                <path d="M82.3,78.2H33.7c-10.6,0-19.3-8.6-19.3-19.3c0-9.3,6.6-17.1,15.4-19c0-0.5-0.1-1-0.1-1.5c0-15.3,12.4-27.7,27.7-27.7c12.2,0,22.8,8,26.4,19.5c8.9,0.8,15.8,8.3,15.8,17.4C99.6,67.8,92,78.2,82.3,78.2z"/>
	            </svg>
	        </span>
	    </label>
	</div>        
        <div class="card mb-4">
            <div class="card-header">
                <h5 class="mb-0">Invoice Details</h5>
            </div>
            <div class="card-body">
                <form id="invoice-form">
                    <div class="config-selection">
                        <div class="mb-3">
                            <label for="configFile" class="form-label">Pre-fill from config file</label>
                            <select class="form-select" id="configFile" name="configFile">
                                <option value="">None selected</option>
                                <!-- Config files will be populated via JavaScript -->
                            </select>
                        </div>
                    </div>
                            
                    <div class="row">
                        <div class="col-md-6">
                            <div class="mb-3">
                                <label for="id" class="form-label">Invoice ID</label>
                                <input type="text" class="form-control" id="id" name="id" placeholder="Auto-generated if empty">
                            </div>
                            <div class="mb-3">
                                <label for="idSuffix" class="form-label">ID Suffix (optional)</label>
                                <input type="text" class="form-control" id="idSuffix" name="idSuffix" placeholder="e.g., -R1">
                            </div>
                            <div class="mb-3">
                                <label for="from" class="form-label">From (Company)</label>
                                <textarea class="form-control" id="from" name="from" rows="3" placeholder="Your Company Name&#10;Address&#10;Contact Information" required></textarea>
                            </div>
                            <div class="mb-3">
                                <label for="to" class="form-label">To (Client)</label>
                                <textarea class="form-control" id="to" name="to" rows="3" placeholder="Client Company Name&#10;Address&#10;Contact Information" required></textarea>
                            </div>
                        </div>
                        <div class="col-md-6">
                            <div class="mb-3">
                                <label for="tax" class="form-label">Tax Rate</label>
                                <input type="number" class="form-control" id="tax" name="tax" step="0.01" value="0.19" required>
                                <small class="text-muted">Default: 19%</small>
                            </div>
                            <div class="mb-3 form-check">
                                <input type="checkbox" class="form-check-input" id="taxExempt" name="taxExempt">
                                <label class="form-check-label" for="taxExempt">Tax exemption (Kleinunternehmer-Regelung)</label>
                                <small class="form-text text-muted d-block">Check this if you are exempt from charging VAT</small>
                            </div>
                            <div class="mb-3">
                                <label for="discount" class="form-label">Discount Rate</label>
                                <input type="number" class="form-control" id="discount" name="discount" step="0.01" value="0">
                                <small class="text-muted">Optional, e.g. 0.1 for 10%</small>
                            </div>
                            <div class="mb-3">
                                <label for="currency" class="form-label">Currency</label>
                                <select class="form-control" id="currency" name="currency" required>
                                    <option value="EUR">EUR (€)</option>
                                    <option value="USD">USD ($)</option>
                                    <option value="GBP">GBP (£)</option>
                                    <option value="CHF">CHF</option>
                                    <option value="JPY">JPY (¥)</option>
                                    <option value="CAD">CAD (C$)</option>
                                    <option value="AUD">AUD (A$)</option>
                                </select>
                            </div>
                            <!-- Footer field visibility options -->
                            <div class="mb-3 form-check">
                                <input type="checkbox" class="form-check-input" id="showRegistration" name="showRegistration" checked>
                                <label class="form-check-label" for="showRegistration">Show Registration Info in Footer</label>
                            </div>
                            <div class="mb-3 form-check">
                                <input type="checkbox" class="form-check-input" id="showVatId" name="showVatId" checked>
                                <label class="form-check-label" for="showVatId">Show VAT ID in Footer</label>
                            </div>
                            <div class="mb-3 form-check tax-exempt-note" style="display: none;">
                                <div class="alert alert-info">
                                    <small>When tax exemption is enabled, the invoice will include a note about §19 UStG (Kleinunternehmer-Regelung)</small>
                                </div>
                            </div>
                            <div class="mb-3">
                                <label for="note" class="form-label">Note</label>
                                <textarea class="form-control" id="note" name="note" rows="3" placeholder="Payment terms, additional information, etc."></textarea>
                            </div>
                        </div>
                    </div>
                    
                    <h5 class="mt-4 mb-3">Invoice Items</h5>
                    <div id="items-container" class="items-container">
                        <div class="item-row">
                            <div class="flex-grow-1">
                                <label for="item-0" class="form-label">Item</label>
                                <input type="text" class="form-control item-name" id="item-0" placeholder="Description" required>
                            </div>
                            <div style="width: 100px;">
                                <label for="quantity-0" class="form-label">Quantity</label>
                                <input type="number" class="form-control item-quantity" id="quantity-0" value="1" min="1" required>
                            </div>
                            <div style="width: 120px;">
                                <label for="rate-0" class="form-label">Rate</label>
                                <input type="number" class="form-control item-rate" id="rate-0" step="0.01" required>
                            </div>
                            <div style="width: 40px;">
                                <button type="button" class="btn btn-danger btn-sm remove-item" disabled>x</button>
                            </div>
                        </div>
                    </div>
                    
                    <button type="button" id="add-item" class="btn btn-secondary btn-sm mt-2">+ Add Item</button>
                    
                    <div class="d-grid gap-2 d-md-flex justify-content-md-end mt-4">
                        <button type="submit" class="btn btn-primary">Generate Invoice</button>
                    </div>
                </form>
            </div>
        </div>
        
        <div id="result-section" class="card">
            <div class="card-header">
                <h5 class="card-title mb-0">Generated Invoice</h5>
            </div>
            <div class="card-body">
                <div class="row">
                    <div class="col-md-8">
                        <div class="ratio ratio-4x3 mb-3">
                            <iframe id="pdf-preview" src="" frameborder="0"></iframe>
                        </div>
                    </div>
                    <div class="col-md-4">
                        <div class="d-grid gap-2">
                            <p><strong>Filename:</strong> <span id="filename"></span></p>
                            <a id="download-link" href="#" class="btn btn-primary mb-2">Download PDF</a>
                            <button id="upload-btn" class="btn btn-success mb-2">Upload to Nextcloud</button>
                            <div id="upload-result" class="mt-2">
                                <div class="alert alert-success" id="upload-success" style="display:none;">
                                    <p>Upload successful!</p>
                                    <p>Share URL: <a id="share-url" href="#" target="_blank"></a></p>
                                </div>
                                <div class="alert alert-danger" id="upload-error" style="display:none;">
                                    <p>Upload failed:</p>
                                    <p id="error-message"></p>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
    <script>
        // Dark mode toggle functionality
	document.addEventListener('DOMContentLoaded', function() {
	    // Find the toggle switch
	    const toggleSwitch = document.getElementById('theme-toggle');
	    if (!toggleSwitch) {
	        console.error('Theme toggle switch not found!');
	        return;
	    }
            
	    // Initialize tax exemption note visibility based on checkbox state
	    const taxExemptCheckbox = document.getElementById('taxExempt');
	    const taxExemptNote = document.querySelector('.tax-exempt-note');
	    if (taxExemptCheckbox && taxExemptNote) {
	        taxExemptNote.style.display = taxExemptCheckbox.checked ? 'block' : 'none';
	    }
	    
	    // Function to set theme
	    function setTheme(themeName) {
	        document.documentElement.setAttribute('data-theme', themeName);
	        localStorage.setItem('theme', themeName);
	        console.log('Theme set to:', themeName);
	    }
	    
	    // Check for saved theme preference or use default
	    const savedTheme = localStorage.getItem('theme') || 'light';
	    setTheme(savedTheme);
	    
	    // Set the toggle switch position based on the current theme
	    toggleSwitch.checked = savedTheme === 'dark';
	    
	    // Add event listener to the toggle switch
	    toggleSwitch.addEventListener('change', function(event) {
	        if (event.target.checked) {
	            setTheme('dark');
	        } else {
	            setTheme('light');
	        }
	    });
                
            // Toggle tax field based on tax exemption status
            document.getElementById('taxExempt').addEventListener('change', function(event) {
                const taxField = document.getElementById('tax');
                if (this.checked) {
                    taxField.setAttribute('disabled', 'disabled');
                    taxField.value = '0';
                } else {
                    taxField.removeAttribute('disabled');
                    taxField.value = '0.19'; // Reset to default German VAT
                }
                
                // Don't clear the config selection for this checkbox
                // This fixes the double-click issue with tax exemption checkbox
                const configSelect = document.getElementById('configFile');
                if (configSelect.value !== "") {
                    // Prevent this event from triggering the auto-deselection
                    event.stopImmediatePropagation();
                }
                
                // Update the form to reflect the tax exemption status
                updateFormForTaxExemption(this.checked);
            });
            
            // Function to update form based on tax exemption status
            function updateFormForTaxExemption(isExempt) {
                // This function can be extended to show/hide additional UI elements
                // related to tax exemption status if needed
                console.log("Tax exemption status changed to:", isExempt);
            }
            
            // Add event listener for config file selection
            document.getElementById('configFile').addEventListener('change', function() {
                if (this.value) {
                    loadConfigData(this.value);
                }
            });
            
            // Add event listeners to automatically deselect config file when user changes form values
            function addChangeListenerToFormElements() {
                // Add listeners to all form input fields
                const formInputs = document.querySelectorAll('#invoice-form input, #invoice-form textarea, #invoice-form select:not(#configFile)');
                formInputs.forEach(function(input) {
                    // Listen for both change and input events to catch all modifications
                    ['change', 'input'].forEach(function(eventType) {
                        input.addEventListener(eventType, function() {
                            // Only deselect if a config file is currently selected
                            const configSelect = document.getElementById('configFile');
                            if (configSelect.value !== "") {
                                console.log('Form field changed, deselecting config file');
                                configSelect.value = "";
                            }
                        });
                    });
                });
                
                // Add specific listener for item add/remove buttons
                document.getElementById('add-item').addEventListener('click', function() {
                    const configSelect = document.getElementById('configFile');
                    if (configSelect.value !== "") {
                        configSelect.value = "";
                    }
                    
                    // Add listeners to the new row's inputs
                    setTimeout(function() {
                        addChangeListenerToFormElements();
                    }, 100);
                });
                
                // Special handling for footer checkboxes to prevent them from auto-clearing config selection
                const footerCheckboxes = ['showRegistration', 'showVatId', 'taxExempt'];
                footerCheckboxes.forEach(function(id) {
                    const checkbox = document.getElementById(id);
                    checkbox.addEventListener('change', function(event) {
                        // Capture the current state of checkboxes when using config files
                        const configSelect = document.getElementById('configFile');
                        if (configSelect.value !== "") {
                            // For these specific checkboxes, don't clear the config selection
                            // This prevents the double-click issue with the tax exemption checkbox
                            // and makes the footer checkboxes work properly with config files
                            event.stopPropagation();
                        }
                        
                        // Show/hide tax exemption note when taxExempt checkbox changes
                        if (id === 'taxExempt') {
                            const taxExemptNote = document.querySelector('.tax-exempt-note');
                            if (taxExemptNote) {
                                taxExemptNote.style.display = this.checked ? 'block' : 'none';
                            }
                        }
                    }, true); // Use capturing phase to intercept before other handlers
                });
                
                // Listen for remove item events through event delegation
                document.getElementById('items-container').addEventListener('click', function(e) {
                    if (e.target.classList.contains('remove-item')) {
                        const configSelect = document.getElementById('configFile');
                        if (configSelect.value !== "") {
                            configSelect.value = "";
                        }
                    }
                });
            }
            
            // Initial setup of change listeners
            addChangeListenerToFormElements();
            
            // Load available config files when page loads
            loadConfigFiles();
        });
        
        // Function to load available config files for the dropdown
        function loadConfigFiles() {
            fetch('/api/config-files')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        const select = document.getElementById('configFile');
                        // Keep the first "None selected" option
                        const defaultOption = select.options[0];
                        // Clear existing options
                        select.innerHTML = '';
                        // Add back the default option
                        select.appendChild(defaultOption);
                        
                        // Add each config file as an option
                        data.files.forEach(file => {
                            const fileName = file.split('/').pop();
                            const option = document.createElement('option');
                            option.value = fileName;
                            option.textContent = fileName;
                            select.appendChild(option);
                        });
                    } else {
                        console.error('Error loading config files:', data.message);
                    }
                })
                .catch(error => {
                    console.error('Error fetching config files:', error);
                });
        }
        
        // Function to load config data and pre-fill form
        function loadConfigData(filename) {
            fetch('/api/config-data/' + filename)
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        prefillForm(data.data);
                    } else {
                        alert('Error loading config data: ' + data.message);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('Failed to load config data.');
                });
        }
        
        // Function to pre-fill form with config data
        function prefillForm(data) {
            // Basic fields
            if (data.from) document.getElementById('from').value = data.from;
            if (data.to) document.getElementById('to').value = data.to;
            
            // Tax handling - handle tax exemption first, then tax value
            if (data.taxExempt !== undefined) {
                const taxExemptBox = document.getElementById('taxExempt');
                taxExemptBox.checked = data.taxExempt;
                
                // Ensure tax field is properly enabled/disabled based on exemption
                const taxField = document.getElementById('tax');
                if (data.taxExempt) {
                    taxField.setAttribute('disabled', 'disabled');
                    taxField.value = '0';
                } else {
                    taxField.removeAttribute('disabled');
                    // Only set default if tax is not defined
                    if (data.tax === undefined) {
                        taxField.value = '0.19';
                    }
                }
            }
            
            // Set tax value after handling exemption status
            if (data.tax !== undefined && !data.taxExempt) {
                document.getElementById('tax').value = data.tax;
            }
            
            // Footer visibility options - extract these from footer object if present
            if (data.footer) {
                if (data.footer.showRegistration !== undefined) {
                    document.getElementById('showRegistration').checked = data.footer.showRegistration;
                }
                if (data.footer.showVatId !== undefined) {
                    document.getElementById('showVatId').checked = data.footer.showVatId;
                }
            } else {
                // Backward compatibility for configs without footer object
                if (data.showRegistration !== undefined) document.getElementById('showRegistration').checked = data.showRegistration;
                if (data.showVatId !== undefined) document.getElementById('showVatId').checked = data.showVatId;
            }
            if (data.discount !== undefined) document.getElementById('discount').value = data.discount;
            if (data.currency) {
                const currencySelect = document.getElementById('currency');
                for (let i = 0; i < currencySelect.options.length; i++) {
                    if (currencySelect.options[i].value === data.currency) {
                        currencySelect.selectedIndex = i;
                        break;
                    }
                }
            }
            if (data.note) document.getElementById('note').value = data.note;
            
            // Items (array data)
            if (data.items && Array.isArray(data.items) && data.items.length > 0) {
                const container = document.getElementById('items-container');
                // Clear existing items except the first one
                while (container.children.length > 1) {
                    container.removeChild(container.lastChild);
                }
                
                // Fill the first item
                container.querySelector('.item-name').value = data.items[0] || '';
                
                if (data.quantities && data.quantities.length > 0) {
                    container.querySelector('.item-quantity').value = data.quantities[0] || 1;
                }
                
                if (data.rates && data.rates.length > 0) {
                    container.querySelector('.item-rate').value = data.rates[0] || '';
                }
                
                // Add additional items if needed
                for (let i = 1; i < data.items.length; i++) {
                    const newRow = document.createElement('div');
                    newRow.className = 'item-row';
                    newRow.innerHTML = '<div class="flex-grow-1"><label for="item-' + i + '" class="form-label">Item</label><input type="text" class="form-control item-name" id="item-' + i + '" placeholder="Description" required></div><div style="width: 100px;"><label for="quantity-' + i + '" class="form-label">Quantity</label><input type="number" class="form-control item-quantity" id="quantity-' + i + '" value="1" min="1" required></div><div style="width: 120px;"><label for="rate-' + i + '" class="form-label">Rate</label><input type="number" class="form-control item-rate" id="rate-' + i + '" step="0.01" required></div><div style="width: 40px;"><button type="button" class="btn btn-danger btn-sm remove-item">x</button></div>';
                    container.appendChild(newRow);
                    
                    // Fill in the data
                    newRow.querySelector('.item-name').value = data.items[i] || '';
                    
                    if (data.quantities && data.quantities.length > i) {
                        newRow.querySelector('.item-quantity').value = data.quantities[i] || 1;
                    }
                    
                    if (data.rates && data.rates.length > i) {
                        newRow.querySelector('.item-rate').value = data.rates[i] || '';
                    }
                }
                
                // Enable/disable remove buttons
                if (container.querySelectorAll('.item-row').length > 1) {
                    container.querySelectorAll('.remove-item').forEach(btn => {
                        btn.disabled = false;
                    });
                }
            }
        }

        // Item management
        let itemCount = 1;
        
        document.getElementById('add-item').addEventListener('click', function() {
            const container = document.getElementById('items-container');
            const newRow = document.createElement('div');
            newRow.className = 'item-row';
            newRow.innerHTML = '<div class="flex-grow-1"><label for="item-' + itemCount + '" class="form-label">Item</label><input type="text" class="form-control item-name" id="item-' + itemCount + '" placeholder="Description" required></div><div style="width: 100px;"><label for="quantity-' + itemCount + '" class="form-label">Quantity</label><input type="number" class="form-control item-quantity" id="quantity-' + itemCount + '" value="1" min="1" required></div><div style="width: 120px;"><label for="rate-' + itemCount + '" class="form-label">Rate</label><input type="number" class="form-control item-rate" id="rate-' + itemCount + '" step="0.01" required></div><div style="width: 40px;"><button type="button" class="btn btn-danger btn-sm remove-item">x</button></div>';
            container.appendChild(newRow);
            itemCount++;
            
            // Enable all remove buttons if more than one item exists
            if (container.querySelectorAll('.item-row').length > 1) {
                container.querySelectorAll('.remove-item').forEach(btn => {
                    btn.disabled = false;
                });
            }
        });
        
        // Event delegation for remove buttons
        document.getElementById('items-container').addEventListener('click', function(e) {
            if (e.target.classList.contains('remove-item')) {
                e.target.closest('.item-row').remove();
                
                // Disable remove button if only one item remains
                const container = document.getElementById('items-container');
                if (container.querySelectorAll('.item-row').length <= 1) {
                    container.querySelector('.remove-item').disabled = true;
                }
            }
        });

        // Invoice form submission
        document.getElementById('invoice-form').addEventListener('submit', function(e) {
            e.preventDefault();
            
            // Collect items, quantities, and rates
            const items = [];
            const quantities = [];
            const rates = [];
            
            document.querySelectorAll('.item-row').forEach(row => {
                items.push(row.querySelector('.item-name').value);
                quantities.push(row.querySelector('.item-quantity').value);
                rates.push(row.querySelector('.item-rate').value);
            });
            
            // Get config file value
            const configFileValue = document.getElementById('configFile').value;
            
            // Create form data
            const formData = {
                from: document.getElementById('from').value,
                to: document.getElementById('to').value,
                items: items.join('||'),
                quantities: quantities.join('||'),
                rates: rates.join('||'),
                tax: parseFloat(document.getElementById('tax').value),
                taxExempt: document.getElementById('taxExempt').checked,
                discount: parseFloat(document.getElementById('discount').value),
                currency: document.getElementById('currency').value,
                // Footer visibility options
                showRegistration: document.getElementById('showRegistration').checked,
                showVatId: document.getElementById('showVatId').checked,
                // Extract company name from the 'from' field (first line)
                companyName: document.getElementById('from').value.split('\n')[0],
                note: document.getElementById('note').value,
                id: document.getElementById('id').value,
                idSuffix: document.getElementById('idSuffix').value,
                // Only use config if a config file is selected in the dropdown
                useConfig: configFileValue !== "",
                configFile: configFileValue
            };
            
            generateInvoice(formData);
        });

        // Generate invoice function
        function generateInvoice(formData) {
            // Ensure tax exemption is properly handled
            if (formData.taxExempt) {
                formData.tax = 0; // Force tax to 0 when tax exempt
            }
            
            fetch('/api/generate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    // Show result section
                    document.getElementById('result-section').style.display = 'block';
                    
                    // Update preview
                    const previewFrame = document.getElementById('pdf-preview');
                    previewFrame.src = '/api/view/' + data.filename;
                    
                    // Update download link
                    const downloadLink = document.getElementById('download-link');
                    downloadLink.href = '/api/download/' + data.filename;
                    downloadLink.download = data.filename;
                    
                    // Update filename display
                    document.getElementById('filename').textContent = data.filename;
                    
                    // Reset upload result display
                    document.getElementById('upload-success').style.display = 'none';
                    document.getElementById('upload-error').style.display = 'none';
                    
                    // Scroll to results
                    document.getElementById('result-section').scrollIntoView({ behavior: 'smooth' });
                } else {
                    alert('Error generating invoice: ' + data.message);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('An error occurred. Please try again.');
            });
        }

        // Upload to Nextcloud
        document.getElementById('upload-btn').addEventListener('click', function() {
            const filename = document.getElementById('filename').textContent;
            
            fetch('/api/upload/' + filename, {
                method: 'POST'
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    document.getElementById('upload-success').style.display = 'block';
                    document.getElementById('upload-error').style.display = 'none';
                    document.getElementById('share-url').href = data.url;
                    document.getElementById('share-url').textContent = data.url;
                } else {
                    document.getElementById('upload-success').style.display = 'none';
                    document.getElementById('upload-error').style.display = 'block';
                    document.getElementById('error-message').textContent = data.message;
                }
            })
            .catch(error => {
                console.error('Error:', error);
                document.getElementById('upload-success').style.display = 'none';
                document.getElementById('upload-error').style.display = 'block';
                document.getElementById('error-message').textContent = 'Network error. Please try again.';
            });
        });
    </script>
</body>
</html>`,
}

// DefaultWebConfig returns the default web configuration
func DefaultWebConfig() WebConfig {
	return WebConfig{
		Port:           8080,
		NextcloudURL:   "https://cloud.example.com",
		NextcloudShare: "/s/share-id",
		UploadScript:   "/var/scripts/cloudsend.sh",
	}
}

// loadWebConfig loads the web server configuration from a JSON file
func loadWebConfig(configPath string) (WebConfig, error) {
	config := DefaultWebConfig()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, fmt.Errorf("unable to read web config: %v", err)
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, fmt.Errorf("invalid JSON in web config: %v", err)
	}

	return config, nil
}

// runWebServer starts the web server
func runWebServer(webConfig WebConfig) error {
	router := gin.Default()

	// Serve static files
	router.Static("/static", "./web/static")

	// API routes
	api := router.Group("/api")
	{
		// Generate invoice
		api.POST("/generate", func(c *gin.Context) {
			var request InvoiceRequest
			if err := c.ShouldBindJSON(&request); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid request data"})
				return
			}

			// Process the request and generate the invoice
			filename, err := generateInvoiceFromRequest(request)
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
		})

		// List available configuration files
		api.GET("/config-files", func(c *gin.Context) {
			files, err := findConfigFiles()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "files": files})
		})
		
		// Get config file data for pre-filling form
		api.GET("/config-data/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			configData, err := getConfigData(filename)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": configData})
		})

		// View generated PDF
		api.GET("/view/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			c.File(filename)
		})

		// Download generated PDF
		api.GET("/download/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
			c.File(filename)
		})

		// Upload to Nextcloud
		api.POST("/upload/:filename", func(c *gin.Context) {
			filename := c.Param("filename")
			result, err := uploadToNextcloud(filename, webConfig.UploadScript, webConfig.NextcloudURL, webConfig.NextcloudShare)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "Upload failed: " + err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, result)
		})
	}

	// Handle index route - serve the HTML template directly
	router.GET("/", func(c *gin.Context) {
		// Debug output to verify our changes
		fmt.Println("\n--- Checking HTML template ---")
		fmt.Println("loadConfigFiles function call present:", strings.Contains(HTMLTemplates["index"], "loadConfigFiles()"))
		fmt.Println("loadConfigFiles function definition present:", strings.Contains(HTMLTemplates["index"], "function loadConfigFiles()"))
		
		c.Header("Content-Type", "text/html")
		c.String(http.StatusOK, HTMLTemplates["index"])
	})

	// Start the server
	return router.Run(fmt.Sprintf(":%d", webConfig.Port))
}

// findConfigFiles returns a list of JSON and YAML config files
func findConfigFiles() ([]string, error) {
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

// generateInvoiceFromRequest processes a web request and generates an invoice
func generateInvoiceFromRequest(request InvoiceRequest) (string, error) {
	var args []string
	var err error

	// Process based on whether we're using a config file or form data
	if request.UseConfig && request.ConfigFile != "" {
		// Using a config file
		args = append(args, "generate", "--import", request.ConfigFile)
		
		// Add optional ID overrides
		if request.Id != "" {
			args = append(args, "--id", request.Id)
		}
		if request.IdSuffix != "" {
			args = append(args, "--id-suffix", request.IdSuffix)
		}
		
		// Other form fields can override config values if provided
		if request.From != "" {
			args = append(args, "--from", request.From)
		}
		if request.To != "" {
			args = append(args, "--to", request.To)
		}
		
		// Process items, quantities, and rates if provided
		if request.Items != "" {
			items := strings.Split(request.Items, "||")
			quantities := strings.Split(request.Quantities, "||")
			rates := strings.Split(request.Rates, "||")

			for i, item := range items {
				args = append(args, "--item", item)
				if i < len(quantities) {
					args = append(args, "--quantity", quantities[i])
				}
				if i < len(rates) {
					args = append(args, "--rate", rates[i])
				}
			}
		}
		
		// Handle tax exemption first
		if request.TaxExempt {
			args = append(args, "--tax-exempt")
			// When tax exempt, force tax to 0
			args = append(args, "--tax", "0")
		} else if request.Tax != 0 {
			// Only add tax if not exempt and value is provided
			args = append(args, "--tax", fmt.Sprintf("%f", request.Tax))
		}
		
		// Add additional fields if provided
		if request.Discount != 0 {
			args = append(args, "--discount", fmt.Sprintf("%f", request.Discount))
		}
		if request.Currency != "" {
			args = append(args, "--currency", request.Currency)
		}
		if request.Note != "" {
			args = append(args, "--note", request.Note)
		}
	} else {
		// Using form data directly
		args = append(args, "generate")
		
		// Add basic invoice info
		if request.From != "" {
			args = append(args, "--from", request.From)
		}
		if request.To != "" {
			args = append(args, "--to", request.To)
		}
		
		// Create a custom config file with footer visibility settings
		tempConfig, err := createTempConfigWithFooterSettings(request)
		if err == nil && tempConfig != "" {
			// Use the temp config
			args = append(args, "--import", tempConfig)
		}

		// Process items, quantities, and rates
		if request.Items != "" {
			items := strings.Split(request.Items, "||")
			quantities := strings.Split(request.Quantities, "||")
			rates := strings.Split(request.Rates, "||")

			for i, item := range items {
				args = append(args, "--item", item)
				if i < len(quantities) {
					args = append(args, "--quantity", quantities[i])
				}
				if i < len(rates) {
					args = append(args, "--rate", rates[i])
				}
			}
		}

		// Add additional fields
		if request.Tax != 0 {
			args = append(args, "--tax", fmt.Sprintf("%f", request.Tax))
		}
		if request.TaxExempt {
			args = append(args, "--tax-exempt")
		}
		if request.Discount != 0 {
			args = append(args, "--discount", fmt.Sprintf("%f", request.Discount))
		}
		if request.Currency != "" {
			args = append(args, "--currency", request.Currency)
		}
		if request.Note != "" {
			args = append(args, "--note", request.Note)
		}
		if request.Id != "" {
			args = append(args, "--id", request.Id)
		}
		if request.IdSuffix != "" {
			args = append(args, "--id-suffix", request.IdSuffix)
		}
	}

	// Create a temporary file to capture the output
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("./invoice", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command failed: %v\nStderr: %s", err, stderr.String())
	}

	// Parse the output to find the generated filename
	// The output should be like "Generated filename.pdf"
	output := stdout.String()
	if strings.Contains(output, "Generated") {
		parts := strings.Split(output, "Generated ")
		if len(parts) > 1 {
			filename := strings.TrimSpace(parts[1])
			return filename, nil
		}
	}

	return "", fmt.Errorf("failed to determine output filename from: %s", output)
}

// uploadToNextcloud uploads a file to Nextcloud using the provided script
func uploadToNextcloud(filename, scriptPath, nextcloudURL, shareID string) (UploadResult, error) {
        result := UploadResult{
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
// createTempConfigWithFooterSettings creates a temporary config file with footer visibility settings
func createTempConfigWithFooterSettings(request InvoiceRequest) (string, error) {
	// Create a minimal invoice with just the footer settings
	invoice := DefaultInvoice()
	
	// Set company name in footer - prefer explicit company name if provided
	if request.CompanyName != "" {
		invoice.Footer.CompanyName = request.CompanyName
	} else if request.From != "" {
		// Fall back to extracting from 'From' field (first line)
		fromLines := strings.Split(request.From, "\n")
		if len(fromLines) > 0 {
			invoice.Footer.CompanyName = fromLines[0]
		}
	}
	
	// Set footer visibility settings
	invoice.Footer.ShowRegistration = request.ShowRegistration
	invoice.Footer.ShowVatId = request.ShowVatId
	
	// If tax exemption is checked, ensure it's reflected in the config
	invoice.TaxExempt = request.TaxExempt
	if request.TaxExempt {
		// Force tax to 0 when tax exempt
		invoice.Tax = 0
	} else if request.Tax > 0 {
		// Only set tax if not exempt and a value is provided
		invoice.Tax = request.Tax
	}
	
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "invoice-config-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}
	defer tmpFile.Close()
	
	// Write the JSON data to the file
	data, err := json.MarshalIndent(invoice, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %v", err)
	}
	
	if _, err := tmpFile.Write(data); err != nil {
		return "", fmt.Errorf("failed to write to temp file: %v", err)
	}
	
	return tmpFile.Name(), nil
}

func getConfigData(filename string) (map[string]interface{}, error) {
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
	} else if strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
		return nil, fmt.Errorf("YAML files not supported for web interface preview")
	} else {
		return nil, fmt.Errorf("unsupported file type: only .json is supported for preview")
	}

	return configData, nil
}
