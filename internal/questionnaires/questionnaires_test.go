package questionnaires

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateFormURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "Google Forms", url: "https://docs.google.com/forms/d/e/example/viewform"},
		{name: "Google short", url: "https://forms.gle/abcdefgh"},
		{name: "Microsoft", url: "https://forms.office.com/r/example"},
		{name: "wrong Google product", url: "https://docs.google.com/document/d/example", wantErr: true},
		{name: "HTTP", url: "http://forms.office.com/r/example", wantErr: true},
		{name: "lookalike", url: "https://forms.gle.example.com/form", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateFormURL(test.url)
			if test.wantErr != errors.Is(err, ErrInvalidInput) {
				t.Fatalf("ValidateFormURL(%q) = %v, want invalid=%t", test.url, err, test.wantErr)
			}
		})
	}
}

func TestParseCSVAndExtractCodes(t *testing.T) {
	t.Parallel()

	raw := "\ufeffname;validation code\nAlice; txq-aaa \nBob;TXQ-BBB\nCopy;TXQ-AAA\nBad\n"
	document, err := ParseCSV(strings.NewReader(raw), 5<<20, 50_000, 100)
	if err != nil {
		t.Fatalf("ParseCSV(): %v", err)
	}
	if document.DelimiterName != "semicolon" || document.DataRowCount != 3 || document.MalformedRowCount != 1 {
		t.Fatalf("ParseCSV() = delimiter %q rows %d malformed %d", document.DelimiterName, document.DataRowCount, document.MalformedRowCount)
	}
	parsed, err := ExtractCodes(document.Raw, document.Delimiter, document.Headers, "validation code", 50_000)
	if err != nil {
		t.Fatalf("ExtractCodes(): %v", err)
	}
	if got, want := strings.Join(parsed.Codes, ","), "TXQ-AAA,TXQ-BBB"; got != want {
		t.Fatalf("codes = %q, want %q", got, want)
	}
	if parsed.DuplicateCount != 1 || parsed.MalformedCount != 1 {
		t.Fatalf("ExtractCodes() duplicate=%d malformed=%d, want 1/1", parsed.DuplicateCount, parsed.MalformedCount)
	}
}

func TestParseCSVLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		maxBytes   int64
		maxRows    int
		maxColumns int
	}{
		{name: "bytes", input: "code\n1234\n", maxBytes: 4, maxRows: 10, maxColumns: 10},
		{name: "rows", input: "code\n1234\n5678\n", maxBytes: 100, maxRows: 1, maxColumns: 10},
		{name: "columns", input: "one,two\na,b\n", maxBytes: 100, maxRows: 10, maxColumns: 1},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ParseCSV(strings.NewReader(test.input), test.maxBytes, test.maxRows, test.maxColumns)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("ParseCSV() error = %v, want ErrInvalidInput", err)
			}
		})
	}
}

func TestParseCSVRejectsMalformedPhysicalHeader(t *testing.T) {
	t.Parallel()

	// A quote error in the first record must not cause the next data row to be
	// presented as headers and then fail only after the administrator confirms.
	_, err := ParseCSV(strings.NewReader("\"unterminated\ncode,name\nTXQ-ONE,Alice\n"), 1<<20, 100, 10)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("ParseCSV() error = %v, want ErrInvalidInput", err)
	}
}
