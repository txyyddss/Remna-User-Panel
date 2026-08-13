package questionnaires

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

var candidateDelimiters = []rune{',', ';', '\t'}

// ParsedCodes contains unique normalized codes and row-quality counts.

type ParsedCodes struct {
	Codes          []string
	DuplicateCount int
	MalformedCount int
}

// ParseCSV reads and parses a bounded UTF-8 CSV upload.

func ParseCSV(reader io.Reader, maxBytes int64, maxRows, maxColumns int) (CSVDocument, error) {
	if reader == nil || maxBytes <= 0 || maxRows <= 0 || maxColumns <= 0 {
		return CSVDocument{}, fmt.Errorf("%w: invalid CSV parser limits", ErrInvalidInput)
	}
	raw, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return CSVDocument{}, fmt.Errorf("read questionnaire CSV: %w", err)
	}
	if int64(len(raw)) > maxBytes {
		return CSVDocument{}, fmt.Errorf("%w: CSV exceeds %d bytes", ErrInvalidInput, maxBytes)
	}
	raw = bytes.TrimPrefix(raw, []byte{0xef, 0xbb, 0xbf})
	if len(raw) == 0 || !utf8.Valid(raw) {
		return CSVDocument{}, fmt.Errorf("%w: CSV must be non-empty UTF-8", ErrInvalidInput)
	}
	delimiter := detectDelimiter(raw)
	document, err := parseDocument(raw, delimiter, maxRows, maxColumns)
	if err != nil {
		return CSVDocument{}, err
	}
	document.Raw = append([]byte(nil), raw...)
	document.Delimiter = delimiter
	document.DelimiterName = delimiterLabel(delimiter)
	return document, nil
}

func parseDocument(raw []byte, delimiter rune, maxRows, maxColumns int) (CSVDocument, error) {
	reader := newCSVReader(raw, delimiter)
	var document CSVDocument
	rowNumber := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		rowNumber++
		if rowNumber > maxRows+1 {
			return CSVDocument{}, fmt.Errorf("%w: CSV exceeds %d data rows", ErrInvalidInput, maxRows)
		}
		if document.Headers == nil {
			// The selected headers must describe the physical first record because
			// ExtractCodes reparses the durable upload from that same boundary.
			// Silently skipping a malformed first row would produce a preview that
			// can never be analyzed reliably.
			if err != nil || len(record) == 0 || len(record) > maxColumns {
				return CSVDocument{}, fmt.Errorf("%w: CSV header row is malformed", ErrInvalidInput)
			}
			document.Headers = trimRecord(record)
			if hasEmptyHeader(document.Headers) {
				return CSVDocument{}, fmt.Errorf("%w: CSV headers must be non-empty", ErrInvalidInput)
			}
			continue
		}
		if err != nil {
			document.MalformedRowCount++
			continue
		}
		if len(record) == 0 || len(record) > maxColumns {
			document.MalformedRowCount++
			continue
		}
		if len(record) != len(document.Headers) {
			document.MalformedRowCount++
			continue
		}
		document.DataRowCount++
		if document.DataRowCount > maxRows {
			return CSVDocument{}, fmt.Errorf("%w: CSV exceeds %d data rows", ErrInvalidInput, maxRows)
		}
		if len(document.SampleRows) < 10 {
			document.SampleRows = append(document.SampleRows, trimRecord(record))
		}
	}
	if len(document.Headers) == 0 {
		return CSVDocument{}, fmt.Errorf("%w: CSV has no header row", ErrInvalidInput)
	}
	return document, nil
}

// ExtractCodes reparses a stored document using an administrator-selected column.

func ExtractCodes(raw []byte, delimiter rune, headers []string, codeColumn string, maxRows int) (ParsedCodes, error) {
	columnIndex := -1
	for index, header := range headers {
		if header == codeColumn {
			if columnIndex != -1 {
				return ParsedCodes{}, fmt.Errorf("%w: selected code header is ambiguous", ErrInvalidInput)
			}
			columnIndex = index
		}
	}
	if columnIndex == -1 {
		return ParsedCodes{}, fmt.Errorf("%w: selected code column does not exist", ErrInvalidInput)
	}
	reader := newCSVReader(raw, delimiter)
	if _, err := reader.Read(); err != nil {
		return ParsedCodes{}, fmt.Errorf("%w: cannot read stored CSV header", ErrInvalidInput)
	}
	seen := make(map[string]struct{})
	result := ParsedCodes{Codes: make([]string, 0)}
	rows := 0
	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		rows++
		if rows > maxRows {
			return ParsedCodes{}, fmt.Errorf("%w: stored CSV exceeds row limit", ErrInvalidInput)
		}
		if err != nil || len(record) != len(headers) || columnIndex >= len(record) {
			result.MalformedCount++
			continue
		}
		code := NormalizeValidationCode(record[columnIndex])
		if code == "" {
			result.MalformedCount++
			continue
		}
		if _, exists := seen[code]; exists {
			result.DuplicateCount++
			continue
		}
		seen[code] = struct{}{}
		result.Codes = append(result.Codes, code)
	}
	return result, nil
}

func detectDelimiter(raw []byte) rune {
	best := candidateDelimiters[0]
	bestScore := -1
	for _, delimiter := range candidateDelimiters {
		reader := newCSVReader(raw, delimiter)
		first, err := reader.Read()
		if err != nil || len(first) == 0 {
			continue
		}
		score := len(first) * 100
		for count := 0; count < 20; count++ {
			record, readErr := reader.Read()
			if errors.Is(readErr, io.EOF) {
				break
			}
			if readErr == nil && len(record) == len(first) {
				score++
			}
		}
		if score > bestScore {
			best, bestScore = delimiter, score
		}
	}
	return best
}

func newCSVReader(raw []byte, delimiter rune) *csv.Reader {
	reader := csv.NewReader(bytes.NewReader(raw))
	reader.Comma = delimiter
	reader.FieldsPerRecord = -1
	reader.ReuseRecord = false
	return reader
}
