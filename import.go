package main

import (
        "encoding/json"
        "fmt"
        "os"
        "strings"

        "github.com/spf13/pflag"
        "gopkg.in/yaml.v3"
)

func importData(path string, structure *Invoice, flags *pflag.FlagSet) error {
        // Read the file
        fileText, err := os.ReadFile(path)
        if err != nil {
                return fmt.Errorf("unable to read file: %v", err)
        }

        // Remove UTF-8 BOM if present
        if len(fileText) >= 3 && fileText[0] == 0xEF && fileText[1] == 0xBB && fileText[2] == 0xBF {
                fileText = fileText[3:]
        }

        // Create temporary structure to ensure footer gets populated
        tempStructure := DefaultInvoice()

        // Check file type first
        var fileType string
        if strings.HasSuffix(path, ".json") {
                fileType = "json"
        } else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
                fileType = "yaml"
        } else {
                return fmt.Errorf("unsupported file type: only .json, .yaml, or .yml are supported")
        }

        // Now copy the structure after checking file type
        *structure = tempStructure

        // Import based on file extension
        if fileType == "json" {
                // First parse JSON into a map to validate it
                var jsonMap map[string]interface{}
                err := json.Unmarshal(fileText, &jsonMap)
                if err != nil {
                        return fmt.Errorf("invalid JSON: %v", err)
                }

                // Now parse into our structure
                err = json.Unmarshal(fileText, structure)
                if err != nil {
                        return fmt.Errorf("JSON structure mapping error: %v", err)
                }
        } else if fileType == "yaml" {
                err = yaml.Unmarshal(fileText, structure)
                if err != nil {
                        return fmt.Errorf("YAML parsing error: %v", err)
                }
        }

        // Process command line flags (these override file values)
        var byteBuffer [][]byte
        flags.Visit(func(f *pflag.Flag) {
                var b []byte
                if f.Value.Type() != "string" {
                        b = []byte(fmt.Sprintf(`{"%s":%s}`, f.Name, f.Value))
                } else {
                        b = []byte(fmt.Sprintf(`{"%s":"%s"}`, f.Name, f.Value))
                }
                byteBuffer = append(byteBuffer, b)
        })

        // Apply flag overrides without touching the footer
        footerBackup := structure.Footer
        for _, bytes := range byteBuffer {
                err = json.Unmarshal(bytes, structure)
                if err != nil {
                        fmt.Fprintf(os.Stderr, "Warning: Error applying flag override: %v\n", err)
                }
        }

        // Restore footer if it was overwritten by flags
        if structure.Footer.CompanyName == "" && footerBackup.CompanyName != "" {
                structure.Footer = footerBackup
        }

        return nil
}