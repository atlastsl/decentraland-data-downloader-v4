package csv

import (
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

func MapToCSV(filePath string, data []map[string]any, headers []string) (err error) {
	csvFile, csvErr := os.Create(filePath)
	if csvErr != nil {
		err = fmt.Errorf("failed to create CSV file: %w", csvErr)
		return
	}
	defer func(csvFile *os.File) {
		_ = csvFile.Close()
	}(csvFile)

	writer := csv.NewWriter(csvFile)
	if writeErr := writer.Write(headers); writeErr != nil {
		err = fmt.Errorf("failed to write CSV headers: %w", writeErr)
		return
	}

	headerIndex := make(map[string]int, len(headers))
	for j, h := range headers {
		headerIndex[h] = j
	}

	record := make([]string, len(headers))
	for _, row := range data {
		for j := range record {
			record[j] = ""
		}
		for h, v := range row {
			j, ok := headerIndex[h]
			if !ok || v == nil {
				continue
			}
			switch val := v.(type) {
			case string:
				record[j] = val
			case int:
				record[j] = strconv.Itoa(val)
			case float64:
				record[j] = strconv.FormatFloat(val, 'f', -1, 64)
			default:
				record[j] = fmt.Sprintf("%v", val)
			}
		}
		if writeErr := writer.Write(record); writeErr != nil {
			err = fmt.Errorf("failed to write CSV row: %w", writeErr)
			return
		}
	}
	writer.Flush()
	if writeErr := writer.Error(); writeErr != nil {
		err = fmt.Errorf("csv flush error: %w", writeErr)
		return
	}

	return
}

func ReadToMap(filePath string) (data []map[string]string, err error) {
	// Open the file
	var file *os.File
	file, err = os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Read the first row as headers
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	// Initialize a slice to hold all rows
	data = make([]map[string]string, 0)

	// Read each row and map it to the headers
	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		entry := make(map[string]string)
		for i, value := range row {
			if i < len(headers) {
				entry[headers[i]] = value
			}
		}

		data = append(data, entry)
	}

	return data, nil
}

func StructToCSV(filePath string, data any) (err error) {
	csvFile, csvErr := os.Create(filePath)
	if csvErr != nil {
		err = fmt.Errorf("failed to create CSV file: %w", err)
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(csvFile)

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	dataSlice := reflect.ValueOf(data)
	if dataSlice.Kind() != reflect.Slice || dataSlice.Len() == 0 {
		return fmt.Errorf("data must be a non-empty slice of structs")
	}

	elemType := dataSlice.Index(0).Type()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("data must be a slice of structs")
	}

	headers := make([]string, 0)
	for i := 0; i < elemType.NumField(); i++ {
		h, ok := elemType.Field(i).Tag.Lookup("bson")
		if !ok {
			h, _ = elemType.Field(i).Tag.Lookup("json")
		}
		h = strings.Split(h, ",")[0]
		if h != "metadata" && h != "" {
			headers = append(headers, h)
		}
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	for k := 0; k < dataSlice.Len(); k++ {
		datum := dataSlice.Index(k)
		if datum.Type().Kind() == reflect.Ptr {
			datum = datum.Elem()
		}
		if datum.Type() != elemType {
			return fmt.Errorf("provided data contains element of incorrect type")
		}

		values := make([]string, len(headers))
		i := 0
		for j := 0; j < elemType.NumField(); j++ {
			h, ok := elemType.Field(j).Tag.Lookup("bson")
			if !ok {
				h, _ = elemType.Field(j).Tag.Lookup("json")
			}
			h = strings.Split(h, ",")[0]
			if !slices.Contains(headers, h) {
				continue
			}

			field := datum.Field(j)
			if field.Kind() == reflect.Ptr {
				if !field.IsNil() {
					field = field.Elem()
				} else {
					values[i] = ""
					i++
					continue
				}
			}
			if field.IsValid() && field.CanInterface() {
				switch field.Kind() {
				case reflect.String:
					values[i] = field.String()
				case reflect.Int, reflect.Int64:
					values[i] = fmt.Sprintf("%d", field.Int())
				case reflect.Float64:
					values[i] = fmt.Sprintf("%f", field.Float())
				case reflect.Struct:
					if field.Type() == reflect.TypeOf(time.Time{}) {
						values[i] = field.Interface().(time.Time).Format(time.RFC3339)
					} else {
						values[i] = fmt.Sprintf("%v", field.Interface())
					}
				default:
					values[i] = fmt.Sprintf("%v", field.Interface())
				}
			} else {
				values[i] = ""
			}
			i++
		}
		if err := writer.Write(values); err != nil {
			return fmt.Errorf("failed to write transaction record: %w", err)
		}
	}
	writer.Flush()
	if writeErr := writer.Error(); writeErr != nil {
		err = fmt.Errorf("csv flush error: %w", writeErr)
		return err
	}

	return nil
}

func ReadToStruct(filePath string, result any) error {
	// result must be a pointer to a slice of structs
	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() != reflect.Ptr || resultValue.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result must be a pointer to a slice of structs")
	}

	sliceValue := resultValue.Elem()
	elemType := sliceValue.Type().Elem()

	isPtr := false
	structType := elemType
	if structType.Kind() == reflect.Ptr {
		isPtr = true
		structType = structType.Elem()
	}
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("result must be a pointer to a slice of structs")
	}

	// Build a map from tag name -> field index (mirroring StructToCSV tag logic)
	tagToFieldIndex := make(map[string]int)
	for i := 0; i < structType.NumField(); i++ {
		h, ok := structType.Field(i).Tag.Lookup("bson")
		if !ok {
			h, _ = structType.Field(i).Tag.Lookup("json")
		}
		h = strings.Split(h, ",")[0]
		if h != "metadata" && h != "" {
			tagToFieldIndex[h] = i
		}
	}

	// Open the CSV file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	reader := csv.NewReader(file)

	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Map each CSV column index to its corresponding struct field index
	colToFieldIndex := make([]int, len(headers))
	for i, h := range headers {
		if idx, ok := tagToFieldIndex[h]; ok {
			colToFieldIndex[i] = idx
		} else {
			colToFieldIndex[i] = -1
		}
	}

	// Read rows
	for {
		row, readErr := reader.Read()
		if readErr != nil {
			if readErr.Error() == "EOF" {
				break
			}
			return fmt.Errorf("failed to read CSV row: %w", readErr)
		}

		newElem := reflect.New(structType).Elem()

		for i, value := range row {
			if i >= len(colToFieldIndex) || colToFieldIndex[i] == -1 {
				continue
			}

			fieldIdx := colToFieldIndex[i]
			field := newElem.Field(fieldIdx)

			if err := setFieldValue(field, value); err != nil {
				return fmt.Errorf("failed to set field %q (column %q): %w",
					structType.Field(fieldIdx).Name, headers[i], err)
			}
		}

		if isPtr {
			ptr := reflect.New(structType)
			ptr.Elem().Set(newElem)
			sliceValue = reflect.Append(sliceValue, ptr)
		} else {
			sliceValue = reflect.Append(sliceValue, newElem)
		}
	}

	resultValue.Elem().Set(sliceValue)
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	if value == "" {
		return nil
	}

	// Handle pointer fields: allocate and dereference
	if field.Kind() == reflect.Ptr {
		ptr := reflect.New(field.Type().Elem())
		if err := setFieldValue(ptr.Elem(), value); err != nil {
			return err
		}
		field.Set(ptr)
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as int: %w", value, err)
		}
		field.SetInt(intVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %q as float: %w", value, err)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("cannot parse %q as bool: %w", value, err)
		}
		field.SetBool(boolVal)
	case reflect.Struct:
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return fmt.Errorf("cannot parse %q as time.Time: %w", value, err)
			}
			field.Set(reflect.ValueOf(t))
		} else {
			return fmt.Errorf("unsupported struct type: %s", field.Type())
		}
	default:
		return fmt.Errorf("unsupported field kind: %s", field.Kind())
	}
	return nil
}
