package main

import (
        "encoding/json"
        "fmt"
        "log"
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

        log.Printf("DEBUG: Read file %s with %d bytes", path, len(fileText))

        // Remove UTF-8 BOM if present
        if len(fileText) >= 3 && fileText[0] == 0xEF && fileText[1] == 0xBB && fileText[2] == 0xBF {
                fileText = fileText[3:]
                log.Printf("DEBUG: Removed UTF-8 BOM from file")
        }

        // Create temporary structure to ensure footer gets populated
        tempStructure := DefaultInvoice()

        // Import based on file extension
        if strings.HasSuffix(path, ".json") {
                log.Printf("DEBUG: Processing as JSON file")

                // First parse JSON into a map to validate it
                var jsonMap map[string]interface{}
                err := json.Unmarshal(fileText, &jsonMap)
                if err != nil {
                        return fmt.Errorf("invalid JSON: %v", err)
                }

                // Now parse into our temp structure
                err = json.Unmarshal(fileText, &tempStructure)
                if err != nil {
                        return fmt.Errorf("JSON structure mapping error: %v", err)
                }

                // Debug what was parsed
                log.Printf("DEBUG: JSON parsed company name: %s", tempStructure.Footer.CompanyName)

                // Copy the temp structure to the actual one
                *structure = tempStructure

        } else if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
                log.Printf("DEBUG: Processing as YAML file")
                err = yaml.Unmarshal(fileText, &tempStructure)
                if err != nil {
                        return fmt.Errorf("yaml parsing error: %v", err)
                }

                // Copy the temp structure to the actual one
                *structure = tempStructure
        } else {
                return fmt.Errorf("unsupported file type")
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
                log.Printf("DEBUG: Flag override: %s", string(b))
        })

        // Apply flag overrides without touching the footer
        footerBackup := structure.Footer
        for _, bytes := range byteBuffer {
                err = json.Unmarshal(bytes, structure)
                if err != nil {
                        log.Printf("WARNING: Error applying flag override: %v", err)
                }
        }

        // Restore footer if it was overwritten by flags
        if structure.Footer.CompanyName == "" && footerBackup.CompanyName != "" {
                structure.Footer = footerBackup
        }

        log.Printf("DEBUG: Final footer company name: %s", structure.Footer.CompanyName)

        return nil
}

func importJson(text []byte, structure *Invoice) error {
        err := json.Unmarshal(text, structure)
        if err != nil {
                return fmt.Errorf("json parsing error: %v", err)
        }

        return nil
}

func importYaml(text []byte, structure *Invoice) error {
        err := yaml.Unmarshal(text, structure)
        if err != nil {
                return fmt.Errorf("yaml parsing error: %v", err)
        }

        return nil
}
