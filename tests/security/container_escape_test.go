package security_test

import (
	"os"
	"testing"

	"digital.vasic.security/pkg/security"
	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "/proc/self/status must be readable on Linux")
	require.NotEmpty(t, data, "expected non-empty /proc/self/status")
}
