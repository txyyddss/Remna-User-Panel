package backup

import (
	"encoding/json"
	"fmt"
	"io"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func truncateError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 500 {
		return value[:500]
	}
	return value
}

func markerPath(databasePath string) string { return databasePath + markerSuffix }
func resultPath(databasePath string) string { return databasePath + resultSuffix }

func validStagePath(databasePath, stagePath, jobID string) bool {
	expected := filepath.Join(filepath.Dir(databasePath), "."+filepath.Base(databasePath)+".restore-"+jobID+".stage")
	return samePath(expected, stagePath)
}

func samePath(left, right string) bool {
	leftAbsolute, leftErr := filepath.Abs(filepath.Clean(left))
	rightAbsolute, rightErr := filepath.Abs(filepath.Clean(right))
	if leftErr != nil || rightErr != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(leftAbsolute, rightAbsolute)
	}
	return leftAbsolute == rightAbsolute
}

func readRestoreMarker(path string) (restoreMarker, error) {
	data, err := readSmallFile(path)
	if err != nil {
		return restoreMarker{}, err
	}
	var marker restoreMarker
	if err := json.Unmarshal(data, &marker); err != nil {
		return restoreMarker{}, fmt.Errorf("decode restore marker: %w", err)
	}
	return marker, nil
}

func readStartupResult(path string) (StartupRestoreResult, error) {
	data, err := readSmallFile(path)
	if err != nil {
		return StartupRestoreResult{}, err
	}
	var result StartupRestoreResult
	if err := json.Unmarshal(data, &result); err != nil {
		return StartupRestoreResult{}, fmt.Errorf("decode restore result: %w", err)
	}
	return result, nil
}

func readSmallFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return io.ReadAll(io.LimitReader(file, 64<<10))
}

func writeJSONAtomic(path string, value any) (returnedErr error) {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode durable state: %w", err)
	}
	temporary := path + ".tmp"
	file, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			if closeErr := file.Close(); returnedErr == nil && closeErr != nil {
				returnedErr = closeErr
			}
		}
		if returnedErr != nil {
			_ = os.Remove(temporary)
		}
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	closed = true
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		directory, err := os.Open(filepath.Dir(path))
		if err != nil {
			return err
		}
		syncErr := directory.Sync()
		closeErr := directory.Close()
		if syncErr != nil {
			return syncErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
