package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestFormatDataJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatJSON, false, buf)

	type person struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	data := []person{{ID: 1, Name: "Alice"}}

	if err := formatter.FormatData(data); err != nil {
		t.Fatalf("FormatData(JSON) returned error: %v", err)
	}

	want := "[\n  {\n    \"id\": 1,\n    \"name\": \"Alice\"\n  }\n]\n"
	if got := buf.String(); got != want {
		t.Fatalf("unexpected JSON output:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFormatDataYAML(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatYAML, false, buf)

	data := map[string]interface{}{
		"id":   1,
		"name": "Alice",
	}

	if err := formatter.FormatData(data); err != nil {
		t.Fatalf("FormatData(YAML) returned error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "id: 1") || !strings.Contains(got, "name: Alice") {
		t.Fatalf("unexpected YAML output: %q", got)
	}
}

func TestFormatDataCSV(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatCSV, false, buf)

	data := []map[string]interface{}{
		{"b": 2, "a": 1},
	}

	if err := formatter.FormatData(data); err != nil {
		t.Fatalf("FormatData(CSV) returned error: %v", err)
	}

	want := "a,b\n1,2\n"
	if got := buf.String(); got != want {
		t.Fatalf("unexpected CSV output:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestFormatDataEmptyTable(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(FormatTable, false, buf)

	var data []map[string]interface{}
	if err := formatter.FormatData(data); err != nil {
		t.Fatalf("FormatData(Table) returned error: %v", err)
	}

	want := "No data found\n"
	if got := buf.String(); got != want {
		t.Fatalf("unexpected table output for empty data: want %q, got %q", want, got)
	}
}

func TestGetOrderedHeadersFromRecord(t *testing.T) {
	record := map[string]interface{}{
		"role":         "admin",
		"organization": "kubespaces",
		"is_default":   true,
	}

	got := getOrderedHeadersFromRecord(record)
	want := []string{"organization", "role", "is_default"}
	if len(got) != len(want) {
		t.Fatalf("unexpected header length: want %d, got %d", len(want), len(got))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("unexpected headers order: want %v, got %v", want, got)
		}
	}
}

func TestFormatDataUnsupportedFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	formatter := NewFormatter(Format("unsupported"), false, buf)

	err := formatter.FormatData(map[string]string{"key": "value"})
	if err == nil {
		t.Fatalf("expected unsupported format to return an error")
	}
}
