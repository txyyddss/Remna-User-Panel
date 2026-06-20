package httpapi

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"remna-user-panel/internal/config"
)

const (
	backupMaxFiles               = 2000
	backupMaxExpandedBytes int64 = 1 << 30
)

// safePgCommand builds an exec.Cmd for pg_dump/pg_restore using separate
// connection parameters instead of passing the full DatabaseURL as a
// command-line argument, preventing potential shell injection.
func safePgCommand(ctx context.Context, databaseURL string, prog string, extraArgs ...string) *exec.Cmd {
	u, err := url.Parse(databaseURL)
	if err != nil {
		// Fallback: if parsing fails, use the URL directly (backwards compatibility).
		args := append([]string{"-d", databaseURL}, extraArgs...)
		return exec.CommandContext(ctx, prog, args...)
	}
	args := []string{}
	if host := u.Hostname(); host != "" {
		args = append(args, "-h", host)
	}
	if port := u.Port(); port != "" {
		args = append(args, "-p", port)
	}
	if u.User != nil {
		if username := u.User.Username(); username != "" {
			args = append(args, "-U", username)
		}
	}
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName != "" {
		args = append(args, "-d", dbName)
	}
	args = append(args, extraArgs...)
	cmd := exec.CommandContext(ctx, prog, args...)
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			cmd.Env = append(os.Environ(), "PGPASSWORD="+password)
		}
	}
	return cmd
}

type backupArchiveInfo struct {
	Name        string   `json:"name"`
	SizeBytes   int64    `json:"size_bytes"`
	CreatedAt   string   `json:"created_at"`
	HasDatabase bool     `json:"has_database"`
	HasCompose  bool     `json:"has_compose"`
	Warnings    []string `json:"warnings"`
}

type backupManifest struct {
	Format         int    `json:"format"`
	CreatedAt      string `json:"created_at"`
	Version        string `json:"version"`
	DatabaseSHA256 string `json:"database_sha256,omitempty"`
	HasDatabase    bool   `json:"has_database"`
	HasCompose     bool   `json:"has_compose"`
}

func inspectBackupArchive(path string) backupArchiveInfo {
	stat, err := os.Stat(path)
	info := backupArchiveInfo{Name: filepath.Base(path), Warnings: []string{}}
	if err != nil {
		info.Warnings = append(info.Warnings, "stat_failed")
		return info
	}
	info.SizeBytes = stat.Size()
	info.CreatedAt = stat.ModTime().UTC().Format(time.RFC3339)
	if strings.EqualFold(filepath.Ext(path), ".zip") {
		reader, err := zip.OpenReader(path)
		if err != nil {
			info.Warnings = append(info.Warnings, "invalid_zip")
			return info
		}
		defer func() { _ = reader.Close() }()
		for _, file := range reader.File {
			if file.Name == "database/database.dump" {
				info.HasDatabase = true
			}
			if strings.HasPrefix(file.Name, "compose/") && !file.FileInfo().IsDir() {
				info.HasCompose = true
			}
		}
	} else if strings.HasSuffix(strings.ToLower(path), ".dump") {
		info.HasDatabase = true
		info.Warnings = append(info.Warnings, "legacy_database_dump")
	}
	return info
}

func createBackupArchive(ctx context.Context, settings config.Settings, target string) (backupArchiveInfo, error) {
	tmpDir, err := os.MkdirTemp(filepath.Dir(target), ".backup-work-")
	if err != nil {
		return backupArchiveInfo{}, err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()
	dumpPath := filepath.Join(tmpDir, "database.dump")
	if strings.TrimSpace(settings.DatabaseURL) == "" {
		return backupArchiveInfo{}, fmt.Errorf("database_url_not_configured")
	}
	output, err := safePgCommand(ctx, settings.DatabaseURL, "pg_dump", "--format=custom", "--no-owner", "--no-privileges", "-f", dumpPath).CombinedOutput()
	if err != nil {
		return backupArchiveInfo{}, fmt.Errorf("pg_dump_failed: %s", strings.TrimSpace(string(output)))
	}
	dumpBody, err := os.ReadFile(dumpPath)
	if err != nil {
		return backupArchiveInfo{}, err
	}
	digest := sha256.Sum256(dumpBody)
	tmpZip := target + ".tmp"
	file, err := os.OpenFile(tmpZip, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return backupArchiveInfo{}, err
	}
	zw := zip.NewWriter(file)
	writeErr := writeZipFile(zw, "database/database.dump", dumpBody, 0o600)
	hasCompose := false
	if writeErr == nil {
		composeRoot := filepath.Join("/app", "compose-source")
		if stat, statErr := os.Stat(composeRoot); statErr == nil && stat.IsDir() {
			writeErr = filepath.WalkDir(composeRoot, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if path == composeRoot || entry.IsDir() {
					return nil
				}
				rel, err := filepath.Rel(composeRoot, path)
				if err != nil {
					return err
				}
				if !isDeploymentConfig(rel) {
					return nil
				}
				body, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if int64(len(body)) > 64<<20 {
					return nil
				}
				hasCompose = true
				return writeZipFile(zw, filepath.ToSlash(filepath.Join("compose", rel)), body, 0o600)
			})
		}
	}
	manifest := backupManifest{Format: 1, CreatedAt: time.Now().UTC().Format(time.RFC3339), Version: readSmallBuildFile(".build-version", "dev"), DatabaseSHA256: hex.EncodeToString(digest[:]), HasDatabase: true, HasCompose: hasCompose}
	if writeErr == nil {
		body, _ := json.MarshalIndent(manifest, "", "  ")
		writeErr = writeZipFile(zw, "manifest.json", append(body, '\n'), 0o600)
	}
	if closeErr := zw.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if closeErr := file.Close(); writeErr == nil {
		writeErr = closeErr
	}
	if writeErr != nil {
		_ = os.Remove(tmpZip)
		return backupArchiveInfo{}, writeErr
	}
	if err := os.Rename(tmpZip, target); err != nil {
		_ = os.Remove(tmpZip)
		return backupArchiveInfo{}, err
	}
	return inspectBackupArchive(target), nil
}

func writeZipFile(writer *zip.Writer, name string, body []byte, mode os.FileMode) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(mode)
	header.Modified = time.Now()
	destination, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = destination.Write(body)
	return err
}

func isDeploymentConfig(rel string) bool {
	base := strings.ToLower(filepath.Base(rel))
	if base == ".env" || strings.HasPrefix(base, ".env.") || base == "caddyfile" {
		return true
	}
	return (strings.HasPrefix(base, "docker-compose") || strings.HasPrefix(base, "compose")) && (strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml"))
}

func extractBackupArchive(path, tempDir string, restoreCompose bool) (string, bool, error) {
	reader, err := zip.OpenReader(path)
	if err != nil {
		return "", false, err
	}
	defer func() { _ = reader.Close() }()
	if len(reader.File) > backupMaxFiles {
		return "", false, fmt.Errorf("archive_too_many_files")
	}
	var expanded int64
	dumpPath := ""
	hasCompose := false
	for _, file := range reader.File {
		clean := filepath.Clean(filepath.FromSlash(file.Name))
		if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			return "", false, fmt.Errorf("archive_path_invalid")
		}
		expanded += int64(file.UncompressedSize64)
		if expanded > backupMaxExpandedBytes {
			return "", false, fmt.Errorf("archive_too_large")
		}
		if file.FileInfo().IsDir() {
			continue
		}
		var target string
		switch {
		case clean == filepath.Join("database", "database.dump"):
			target = filepath.Join(tempDir, "database.dump")
			dumpPath = target
		case restoreCompose && strings.HasPrefix(clean, "compose"+string(filepath.Separator)):
			target = filepath.Join("/app", "compose-source", strings.TrimPrefix(clean, "compose"+string(filepath.Separator)))
			hasCompose = true
		default:
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return "", false, err
		}
		source, err := file.Open()
		if err != nil {
			return "", false, err
		}
		destination, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			_ = source.Close()
			return "", false, err
		}
		_, copyErr := io.Copy(destination, io.LimitReader(source, backupMaxExpandedBytes+1))
		_ = source.Close()
		_ = destination.Close()
		if copyErr != nil {
			return "", false, copyErr
		}
	}
	return dumpPath, hasCompose, nil
}

func snapshotComposeBeforeRestore(backupDir string) (string, error) {
	root := filepath.Join("/app", "compose-source")
	stat, err := os.Stat(root)
	if err != nil || !stat.IsDir() {
		return "", nil
	}
	name := "pre-restore-compose-" + time.Now().Format("20060102-150405") + ".zip"
	target := filepath.Join(backupDir, name)
	file, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", err
	}
	writer := zip.NewWriter(file)
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root || entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if !isDeploymentConfig(rel) {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return writeZipFile(writer, filepath.ToSlash(filepath.Join("compose", rel)), body, 0o600)
	})
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(target)
		return "", err
	}
	return name, nil
}
