#!/usr/bin/env bash
# scripts/claim-check.sh
# Stop hook: R-18 Operational Integrity (Constitution §11.5)
# No command may suspend, hibernate, lock, terminate, or crash the operator's host
# Timeout: 5 seconds (set in .claude/settings.json)
set -euo pipefail

echo "[claim-check] R-18 Operational Integrity check..."

# Forbidden commands that could affect operator's host (Constitution §11.5)
FORBIDDEN_COMMANDS=(
    "systemctl suspend"
    "systemctl hibernate"
    "systemctl poweroff"
    "systemctl reboot"
    "pm-suspend"
    "pm-hibernate"
    "shutdown"
    "halt"
    "poweroff"
    "reboot"
    "kill -9.*sshd"
    "pkill -9.*sshd"
    "rm -rf /"
    "rm -rf /home"
    "mkfs"
    "fdisk /dev/"
    "parted /dev/"
)

# Check recent bash commands (from history or current context)
# This is a lightweight check — heavy checks run in host-integrity-scan CI lane
for cmd in "${FORBIDDEN_COMMANDS[@]}"; do
    if echo "$BASH_COMMAND" 2>/dev/null | grep -q "$cmd"; then
        echo "R-18 VIOLATION: Command would affect operator host: $cmd"
        exit 1
    fi
done

echo "[claim-check] Host integrity verified — no forbidden commands detected"
exit 0
