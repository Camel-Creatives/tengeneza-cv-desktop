package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/troubleman96/tengeneza-cv-desktop/internal/domain"
)

func TestLocalStorePersistsCVsAndAllowsUnlimitedExports(t *testing.T) {
	path := filepath.Join(t.TempDir(), "private", "workspace.json")
	local, err := OpenLocalStore(path)
	if err != nil {
		t.Fatal(err)
	}
	user, err := local.GetUser(DefaultUserID)
	if err != nil {
		t.Fatal(err)
	}
	if user.Plan != LocalDesktopPlan {
		t.Fatalf("plan = %q, want %q", user.Plan, LocalDesktopPlan)
	}
	if got := local.DailyExportsRemaining(DefaultUserID, user.CreatedAt); got != -1 {
		t.Fatalf("remaining exports = %d, want -1 for unlimited", got)
	}

	cv, err := local.CreateCV(domain.CV{
		UserID: DefaultUserID, Title: "Local CV", TemplateID: "ats_classic",
		DocumentStyle: "resume", TargetRegion: "eac", Language: "en",
		Content: map[string]any{"summary": "initial"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		content := map[string]any{"summary": fmt.Sprintf("revision %d", i)}
		if _, err := local.PatchCV(DefaultUserID, cv.ID, content, "test revision"); err != nil {
			t.Fatal(err)
		}
		if _, err := local.ConsumeExportCredit(DefaultUserID, cv.ID, "pdf"); err != nil {
			t.Fatalf("export %d was blocked: %v", i+1, err)
		}
	}

	reopened, err := OpenLocalStore(path)
	if err != nil {
		t.Fatal(err)
	}
	persisted, err := reopened.CV(DefaultUserID, cv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got := persisted.Content["summary"]; got != "revision 7" {
		t.Fatalf("persisted summary = %#v", got)
	}
	versions, err := reopened.Versions(DefaultUserID, cv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 9 {
		t.Fatalf("version count = %d, want 9", len(versions))
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0077 != 0 {
		t.Fatalf("workspace permissions = %o, want private to current user", info.Mode().Perm())
	}
}

func TestHostedMemoryStoreKeepsFiveExportLimit(t *testing.T) {
	memory := NewMemoryStore()
	cv, err := memory.CreateCV(domain.CV{
		UserID: DefaultUserID, Title: "Hosted CV", TemplateID: "ats_classic",
		Content: map[string]any{"summary": "initial"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < DailyExportLimit; i++ {
		if i > 0 {
			if _, err := memory.PatchCV(DefaultUserID, cv.ID, map[string]any{"summary": fmt.Sprintf("hosted %d", i)}, "test"); err != nil {
				t.Fatal(err)
			}
		}
		if _, err := memory.ConsumeExportCredit(DefaultUserID, cv.ID, "pdf"); err != nil {
			t.Fatalf("export %d unexpectedly failed: %v", i+1, err)
		}
	}
	if _, err := memory.PatchCV(DefaultUserID, cv.ID, map[string]any{"summary": "sixth"}, "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := memory.ConsumeExportCredit(DefaultUserID, cv.ID, "pdf"); err != ErrDailyExportLimit {
		t.Fatalf("sixth hosted export error = %v, want ErrDailyExportLimit", err)
	}
}
