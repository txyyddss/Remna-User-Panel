package backup

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestReconcileUploadsCompletesPublishingCandidate(t *testing.T) {
	service, store, ctx, migrations := newUploadTestService(t)
	user, _, err := store.UpsertTelegramUser(ctx, telegramUploadProfile(44002), false)
	if err != nil {
		t.Fatal(err)
	}
	source, err := service.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(source.Path)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := service.ReceiveUpload(ctx, user.ID, "recover-publishing", "source.db", file, source.SizeBytes+1)
	_ = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.DB().ExecContext(ctx, `UPDATE backup_uploads SET status='publishing' WHERE id=?`, candidate.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.ReconcileUploads(ctx, migrations); err != nil {
		t.Fatalf("ReconcileUploads(): %v", err)
	}
	var uploadStatus, runStatus string
	if err := store.DB().QueryRowContext(ctx, `SELECT status FROM backup_uploads WHERE id=?`, candidate.ID).Scan(&uploadStatus); err != nil {
		t.Fatal(err)
	}
	if err := store.DB().QueryRowContext(ctx, `SELECT status FROM backup_runs WHERE id=?`, candidate.Backup.ID).Scan(&runStatus); err != nil {
		t.Fatal(err)
	}
	if uploadStatus != "complete" || runStatus != "complete" {
		t.Fatalf("reconciled statuses = upload %q run %q", uploadStatus, runStatus)
	}
}

func TestReconcileUploadsRejectsSymlinkedPublishedFile(t *testing.T) {
	service, store, ctx, migrations := newUploadTestService(t)
	user, _, err := store.UpsertTelegramUser(ctx, telegramUploadProfile(44003), false)
	if err != nil {
		t.Fatal(err)
	}
	source, err := service.Run(ctx)
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(source.Path)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := service.ReceiveUpload(ctx, user.ID, "symlink-publishing", "source.db", file, source.SizeBytes+1)
	_ = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	var temporary, final string
	if err := store.DB().QueryRowContext(ctx, `SELECT temporary_path,final_path FROM backup_uploads WHERE id=?`, candidate.ID).
		Scan(&temporary, &final); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(temporary, final); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if _, err := store.DB().ExecContext(ctx, `UPDATE backup_uploads SET status='publishing' WHERE id=?`, candidate.ID); err != nil {
		t.Fatal(err)
	}
	if err := service.ReconcileUploads(ctx, migrations); err != nil {
		t.Fatalf("ReconcileUploads(symlink): %v", err)
	}
	var status string
	if err := store.DB().QueryRowContext(ctx, `SELECT status FROM backup_uploads WHERE id=?`, candidate.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "failed" {
		t.Fatalf("symlinked upload status = %q, want failed", status)
	}
	if _, err := os.Lstat(final); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("symlinked final file still exists: %v", err)
	}
}

func TestReconcileUploadsFailsPrePublicationCandidate(t *testing.T) {
	service, store, ctx, migrations := newUploadTestService(t)
	user, _, err := store.UpsertTelegramUser(ctx, telegramUploadProfile(44004), false)
	if err != nil {
		t.Fatal(err)
	}
	candidate, err := service.ReceiveUpload(ctx, user.ID, "interrupted-validating", "bad.db", strings.NewReader("candidate"), 1024)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ReconcileUploads(ctx, migrations); err != nil {
		t.Fatal(err)
	}
	var status string
	if err := store.DB().QueryRowContext(ctx, `SELECT status FROM backup_uploads WHERE id=?`, candidate.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "failed" {
		t.Fatalf("interrupted upload status = %q, want failed", status)
	}
}
