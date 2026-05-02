#!/bin/bash
# Host Integrity Scan — Constitution §11.5 / R-18 Operational Integrity
# Checks for forbidden commands, container hazards, and privilege escalation vectors

set -euo pipefail

EXIT_CODE=0

echo "=== Host Integrity Scan ==="
echo "Checking for forbidden commands in build scripts and container definitions..."

# Forbidden commands that must not appear in any script, Dockerfile, or entrypoint
FORBIDDEN=(
    "reboot"
    "shutdown"
    "poweroff"
    "halt"
    "init 0"
    "systemctl poweroff"
    "systemctl reboot"
    "kill -9 1"
    "echo o > /proc/sysrq-trigger"
    "docker rm -f"
    "rm -rf /"
    ":(){ :|:& };:"
    "mkfs."
    "dd if=/dev/zero of=/dev/"
)

# Scan all shell scripts, Dockerfiles, and Go files
for pattern in '*.sh' 'Dockerfile*' '*.go'; do
    while IFS= read -r -d '' file; do
        for cmd in "${FORBIDDEN[@]}"; do
            if grep -qF "$cmd" "$file" 2>/dev/null; then
                echo "  VIOLATION: $file contains forbidden command: $cmd"
                EXIT_CODE=1
            fi
        done
    done < <(find . -type f -name "$pattern" -print0 2>/dev/null || true)
done

# Check for setuid/setgid in container images if Docker is available
if command -v docker >/dev/null 2>&1; then
    echo "Checking running containers for privileged mode..."
    docker ps -q 2>/dev/null | while read -r cid; do
        priv=$(docker inspect --format='{{.HostConfig.Privileged}}' "$cid" 2>/dev/null || echo "false")
        if [ "$priv" = "true" ]; then
            echo "  WARNING: Container $cid is running in privileged mode"
        fi
    done || true
fi

# Check for world-writable directories in PATH
IFS=':' read -ra PATHDIRS <<< "$PATH"
for dir in "${PATHDIRS[@]}"; do
    if [ -d "$dir" ] && [ "$(stat -c %a "$dir" 2>/dev/null || echo 0)" = "777" ]; then
        echo "  WARNING: World-writable directory in PATH: $dir"
    fi
done

if [ $EXIT_CODE -eq 0 ]; then
    echo "=== Host Integrity Scan: PASSED ==="
else
    echo "=== Host Integrity Scan: FAILED ==="
fi

exit $EXIT_CODE
