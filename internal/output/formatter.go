package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"spacectl/internal/models"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

// Format represents the output format
type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatYAML  Format = "yaml"
	FormatCSV   Format = "csv"
)

// Formatter handles output formatting
type Formatter struct {
	format    Format
	noHeaders bool
	writer    io.Writer
}

// NewFormatter creates a new formatter
func NewFormatter(format Format, noHeaders bool, writer io.Writer) *Formatter {
	return &Formatter{
		format:    format,
		noHeaders: noHeaders,
		writer:    writer,
	}
}

// FormatData formats and outputs data
func (f *Formatter) FormatData(data interface{}) error {
	switch f.format {
	case FormatJSON:
		return f.formatJSON(data)
	case FormatYAML:
		return f.formatYAML(data)
	case FormatCSV:
		return f.formatCSV(data)
	case FormatTable:
		return f.formatTable(data)
	default:
		return fmt.Errorf("unsupported format: %s", f.format)
	}
}

func (f *Formatter) formatJSON(data interface{}) error {
	encoder := json.NewEncoder(f.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (f *Formatter) formatYAML(data interface{}) error {
	encoder := yaml.NewEncoder(f.writer)
	defer encoder.Close()
	return encoder.Encode(data)
}

func (f *Formatter) formatCSV(data interface{}) error {
	writer := csv.NewWriter(f.writer)
	defer writer.Flush()

	// Convert data to slice of maps
	records, err := f.convertToRecords(data)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	// Get headers from first record (deterministic order)
	var headers []string
	if !f.noHeaders {
		headers = getOrderedHeadersFromRecord(records[0])
		if err := writer.Write(headers); err != nil {
			return err
		}
	}

	// Write data rows
	for _, record := range records {
		var row []string
		if !f.noHeaders {
			for _, header := range headers {
				row = append(row, fmt.Sprintf("%v", record[header]))
			}
		} else {
			for _, value := range record {
				row = append(row, fmt.Sprintf("%v", value))
			}
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func (f *Formatter) formatTable(data interface{}) error {
	// Convert data to slice of maps
	records, err := f.convertToRecords(data)
	if err != nil {
		return err
	}

	if len(records) == 0 {
		fmt.Fprintln(f.writer, "No data found")
		return nil
	}

	// Create table
	table := tablewriter.NewWriter(f.writer)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)

	// Get headers from first record (deterministic order)
	var headers []string
	for _, key := range getOrderedHeadersFromRecord(records[0]) {
		headers = append(headers, strings.Title(key))
	}
	table.SetHeader(headers)

	// Add data rows
	for _, record := range records {
		var row []string
		for _, header := range headers {
			row = append(row, fmt.Sprintf("%v", record[strings.ToLower(header)]))
		}
		table.Append(row)
	}

	table.Render()
	return nil
}

// convertToRecords converts data to a slice of maps for table/CSV formatting
func (f *Formatter) convertToRecords(data interface{}) ([]map[string]interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		var records []map[string]interface{}
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			// Special-case pretty printing for organizations list
			switch m := item.(type) {
			case models.OrganizationMembershipResponse:
				records = append(records, map[string]interface{}{
					"organization": m.Organization.Name,
					"role":         m.Role,
					"is_default":   m.IsDefault,
				})
			case *models.OrganizationMembershipResponse:
				if m != nil {
					records = append(records, map[string]interface{}{
						"organization": m.Organization.Name,
						"role":         m.Role,
						"is_default":   m.IsDefault,
					})
				}
			case models.ProjectMembership:
				records = append(records, map[string]interface{}{
					"project": m.Project.Name,
					"role":    m.Role,
				})
			case *models.ProjectMembership:
				if m != nil {
					records = append(records, map[string]interface{}{
						"project": m.Project.Name,
						"role":    m.Role,
					})
				}
			case models.Project:
				records = append(records, map[string]interface{}{
					"id":              m.ID,
					"name":            m.Name,
					"organization_id": m.OrganizationID,
					"max_tenants":     m.MaxTenants,
					"max_compute":     m.MaxCompute,
					"max_memory_gb":   m.MaxMemoryGB,
				})
			case *models.Project:
				if m != nil {
					records = append(records, map[string]interface{}{
						"id":              m.ID,
						"name":            m.Name,
						"organization_id": m.OrganizationID,
						"max_tenants":     m.MaxTenants,
						"max_compute":     m.MaxCompute,
						"max_memory_gb":   m.MaxMemoryGB,
					})
				}
			case map[string]interface{}:
				records = append(records, item.(map[string]interface{}))
			default:
				record, err := f.structToMap(item)
				if err != nil {
					return nil, err
				}
				records = append(records, record)
			}
		}
		return records, nil
	case reflect.Struct:
		// Special-case pretty printing for single organization membership
		switch m := v.Interface().(type) {
		case models.OrganizationMembershipResponse:
			return []map[string]interface{}{map[string]interface{}{
				"organization": m.Organization.Name,
				"role":         m.Role,
				"is_default":   m.IsDefault,
			}}, nil
		case *models.OrganizationMembershipResponse:
			if m != nil {
				return []map[string]interface{}{map[string]interface{}{
					"organization": m.Organization.Name,
					"role":         m.Role,
					"is_default":   m.IsDefault,
				}}, nil
			}
			return nil, nil
		case models.ProjectMembership:
			return []map[string]interface{}{map[string]interface{}{
				"project": m.Project.Name,
				"role":    m.Role,
			}}, nil
		case *models.ProjectMembership:
			if m != nil {
				return []map[string]interface{}{map[string]interface{}{
					"project": m.Project.Name,
					"role":    m.Role,
				}}, nil
			}
			return nil, nil
		case models.Project:
			return []map[string]interface{}{map[string]interface{}{
				"id":              m.ID,
				"name":            m.Name,
				"organization_id": m.OrganizationID,
				"max_tenants":     m.MaxTenants,
				"max_compute":     m.MaxCompute,
				"max_memory_gb":   m.MaxMemoryGB,
			}}, nil
		case *models.Project:
			if m != nil {
				return []map[string]interface{}{map[string]interface{}{
					"id":              m.ID,
					"name":            m.Name,
					"organization_id": m.OrganizationID,
					"max_tenants":     m.MaxTenants,
					"max_compute":     m.MaxCompute,
					"max_memory_gb":   m.MaxMemoryGB,
				}}, nil
			}
			return nil, nil
		case map[string]interface{}:
			return []map[string]interface{}{data.(map[string]interface{})}, nil
		default:
			record, err := f.structToMap(data)
			if err != nil {
				return nil, err
			}
			return []map[string]interface{}{record}, nil
		}
	default:
		return nil, fmt.Errorf("unsupported data type for table/CSV formatting")
	}
}

// getOrderedHeadersFromRecord returns a deterministic header order for a record.
// If the record looks like an organization membership row, we enforce a
// human-friendly order. Otherwise, keys are sorted alphabetically.
func getOrderedHeadersFromRecord(record map[string]interface{}) []string {
	// Preferred order for organization membership list
	if hasKeys(record, "organization", "role", "is_default") {
		return []string{"organization", "role", "is_default"}
	}

	// Fallback: sort keys alphabetically for stability
	var keys []string
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func hasKeys(m map[string]interface{}, keys ...string) bool {
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

// structToMap converts a struct to a map[string]interface{}
func (f *Formatter) structToMap(data interface{}) (map[string]interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %s", v.Kind())
	}

	t := v.Type()
	result := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !fieldValue.CanInterface() {
			continue
		}

		// Get JSON tag name, fallback to field name
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty and other options
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "" {
			jsonName = field.Name
		}

		result[jsonName] = fieldValue.Interface()
	}

	return result, nil
}
