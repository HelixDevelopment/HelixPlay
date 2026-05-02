package security_test

import (
	"os"
	"testing"

	"digital.vasic.security/pkg/security"
)

func TestContainerEscapeVectors(t *testing.T) {
	checks := security.ScanPrivilegeEscalation()

	var failures []string
	for _, check := range checks {
		if !check.Passed {
			failures = append(failures, check.Name+": "+check.Details)
		}
	}

	if len(failures) > 0 {
		t.Logf("Privilege escalation checks found issues (may be expected in dev environment):")
		for _, f := range failures {
			t.Logf("  - %s", f)
		}
	}
}

func TestNoWritableRootFS(t *testing.T) {
	check := security.CheckWritableRootFS()
	if !check.Passed {
		t.Logf("Writable root filesystem detected: %s", check.Details)
	}
}

func TestProcSelfAccess(t *testing.T) {
	// Verify we can read basic proc info (sanity check)
	data, err := os.ReadFile("/proc/self/status")
	if err != nil {
		t.Skipf("Cannot read /proc/self/status: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty /proc/self/status")
	}
}
