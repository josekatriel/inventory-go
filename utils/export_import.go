package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// SupportedExportFormat defines the valid export formats
type SupportedExportFormat string

const (
	CSV   SupportedExportFormat = "csv"
	Excel SupportedExportFormat = "excel"
)

// ExportOptions contains configuration for data export
type ExportOptions struct {
	Format    SupportedExportFormat
	Headers   []string
	Data      [][]string
	SheetName string // For Excel format
	FileName  string
}

// ExportToResponse exports data to HTTP response in the requested format
func ExportToResponse(w http.ResponseWriter, options ExportOptions) error {
	// Generate filename if not provided
	if options.FileName == "" {
		timestamp := time.Now().Format("20060102-150405")
		options.FileName = fmt.Sprintf("export-%s", timestamp)
	}

	// Export based on the requested format
	switch options.Format {
	case CSV:
		return exportToCSV(w, options)
	case Excel:
		return exportToExcel(w, options)
	default:
		return fmt.Errorf("unsupported export format: %s", options.Format)
	}
}

// exportToCSV exports data as CSV file to the HTTP response
func exportToCSV(w http.ResponseWriter, options ExportOptions) error {
	// Set headers for CSV file download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", options.FileName))

	// Create CSV writer
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write headers if provided
	if len(options.Headers) > 0 {
		if err := writer.Write(options.Headers); err != nil {
			return fmt.Errorf("error writing CSV headers: %w", err)
		}
	}

	// Write data rows
	for _, row := range options.Data {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing CSV row: %w", err)
		}
	}

	return nil
}

// exportToExcel exports data as Excel file to the HTTP response
func exportToExcel(w http.ResponseWriter, options ExportOptions) error {
	// Create a new Excel file
	f := excelize.NewFile()

	// Use provided sheet name or default to "Sheet1"
	sheetName := options.SheetName
	if sheetName == "" {
		sheetName = "Sheet1"
	}

	// Set the default sheet name
	index, err := f.GetSheetIndex("Sheet1")
	if err == nil {
		f.SetSheetName(f.GetSheetName(index), sheetName)
	}

	// Write headers if provided
	if len(options.Headers) > 0 {
		for i, header := range options.Headers {
			col := string(rune('A' + i))
			cell := fmt.Sprintf("%s1", col)
			f.SetCellValue(sheetName, cell, header)
		}
	}

	// Write data rows
	for rowIndex, row := range options.Data {
		// Excel rows are 1-indexed, add 2 to account for header row
		excelRowIndex := rowIndex + 2
		
		for colIndex, cellValue := range row {
			col := string(rune('A' + colIndex))
			cell := fmt.Sprintf("%s%d", col, excelRowIndex)
			f.SetCellValue(sheetName, cell, cellValue)
		}
	}

	// Set headers for Excel file download
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.xlsx", options.FileName))

	// Write the Excel file to the response
	return f.Write(w)
}

// ImportFromRequest imports data from an uploaded file
func ImportFromRequest(r *http.Request, formField string) ([][]string, error) {
	// Get the file from the request
	file, header, err := r.FormFile(formField)
	if err != nil {
		return nil, fmt.Errorf("failed to get file from request: %w", err)
	}
	defer file.Close()

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	
	// Process based on file format
	switch ext {
	case ".csv":
		return importFromCSV(file)
	case ".xlsx":
		return importFromExcel(file)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", ext)
	}
}

// importFromCSV imports data from a CSV file
func importFromCSV(file multipart.File) ([][]string, error) {
	// Create CSV reader
	reader := csv.NewReader(file)
	
	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %w", err)
	}
	
	return records, nil
}

// importFromExcel imports data from an Excel file
func importFromExcel(file multipart.File) ([][]string, error) {
	// Create a temporary file to store the uploaded Excel file
	tempFile, err := os.CreateTemp("", "upload-*.xlsx")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()
	
	// Copy the uploaded file to the temporary file
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}
	
	// Open the Excel file
	f, err := excelize.OpenFile(tempFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()
	
	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}
	
	// Get all rows from the first sheet
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read rows from Excel: %w", err)
	}
	
	return rows, nil
}

// ExtractDataFromStructs converts a slice of structs to a slice of string slices for CSV/Excel export
func ExtractDataFromStructs(items interface{}, fields []string) ([][]string, error) {
	// Get the reflect.Value of the slice
	v := reflect.ValueOf(items)
	
	// Ensure it's a slice
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("expected slice, got %s", v.Kind())
	}
	
	// Get the number of items in the slice
	n := v.Len()
	result := make([][]string, n)
	
	// Iterate over each item in the slice
	for i := 0; i < n; i++ {
		// Get the i-th item
		item := v.Index(i)
		
		// If it's a pointer, get the element it points to
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		
		// Ensure it's a struct
		if item.Kind() != reflect.Struct {
			return nil, fmt.Errorf("expected struct, got %s", item.Kind())
		}
		
		// Create a row for this item
		row := make([]string, len(fields))
		
		// For each field, get the value
		for j, field := range fields {
			// Handle nested fields with dot notation (e.g., "Basic.Name")
			fieldParts := strings.Split(field, ".")
			fieldValue := item
			
			for _, part := range fieldParts {
				// Convert to title case to match Go struct field naming convention
				// Handle special cases for JSON tag names that might be lowercase
				fieldExists := false
				
				// Try with original field name
				if fieldValue.Kind() == reflect.Struct {
					if f := fieldValue.FieldByName(part); f.IsValid() {
						fieldValue = f
						fieldExists = true
					}
				}
				
				// If field not found, try with first letter capitalized
				if !fieldExists && len(part) > 0 {
					capitalized := strings.ToUpper(part[:1]) + part[1:]
					if fieldValue.Kind() == reflect.Struct {
						if f := fieldValue.FieldByName(capitalized); f.IsValid() {
							fieldValue = f
							fieldExists = true
						}
					}
				}
				
				if !fieldExists {
					fieldValue = reflect.Value{} // Invalid value
					break
				}
			}
			
			// Convert the field value to string
			if fieldValue.IsValid() {
				// Handle different types
				switch fieldValue.Kind() {
				case reflect.String:
					row[j] = fieldValue.String()
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					row[j] = fmt.Sprintf("%d", fieldValue.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					row[j] = fmt.Sprintf("%d", fieldValue.Uint())
				case reflect.Float32, reflect.Float64:
					row[j] = fmt.Sprintf("%f", fieldValue.Float())
				case reflect.Bool:
					row[j] = fmt.Sprintf("%t", fieldValue.Bool())
				case reflect.Struct:
					// Handle time.Time specially
					if t, ok := fieldValue.Interface().(time.Time); ok {
						row[j] = t.Format(time.RFC3339)
					} else {
						// For other structs, use JSON
						bytes, err := json.Marshal(fieldValue.Interface())
						if err != nil {
							row[j] = fmt.Sprintf("%v", fieldValue.Interface())
						} else {
							row[j] = string(bytes)
						}
					}
				case reflect.Map, reflect.Slice:
					// For maps and slices, use JSON
					bytes, err := json.Marshal(fieldValue.Interface())
					if err != nil {
						row[j] = fmt.Sprintf("%v", fieldValue.Interface())
					} else {
						row[j] = string(bytes)
					}
				default:
					row[j] = fmt.Sprintf("%v", fieldValue.Interface())
				}
			} else {
				row[j] = "" // Empty string for invalid fields
			}
		}
		
		result[i] = row
	}
	
	return result, nil
}

// GetStructFieldNames returns a list of field names for a struct type
func GetStructFieldNames(structType interface{}) []string {
	t := reflect.TypeOf(structType)
	
	// If it's a pointer, get the element it points to
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	
	// Ensure it's a struct
	if t.Kind() != reflect.Struct {
		return nil
	}
	
	fieldCount := t.NumField()
	fields := make([]string, 0, fieldCount)
	
	for i := 0; i < fieldCount; i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}
		
		// Get the field name from the json tag if available
		fieldName := field.Name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "-" {
				fieldName = parts[0]
			}
		}
		
		fields = append(fields, fieldName)
	}
	
	return fields
}
