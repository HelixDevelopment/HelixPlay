// containers/tests/bootstrap_test.go
package containers_test

import (
    "os"
    "path/filepath"
    "testing"
)

func TestContainerDefinitionsExist(t *testing.T) {
    containers := []string{
        "host-agent", "capture-service", "encoder-service", "discovery-beacon",
    }
    // The test runs from tests/ directory, so we need to go up one level
    wd, _ := os.Getwd()
    projectRoot := filepath.Dir(wd)
    
    for _, name := range containers {
        dockerfile := filepath.Join(projectRoot, "containers", name, "Dockerfile")
        if _, err := os.Stat(dockerfile); os.IsNotExist(err) {
            t.Errorf("Missing Dockerfile for %s: %s", name, dockerfile)
        }
    }
}
