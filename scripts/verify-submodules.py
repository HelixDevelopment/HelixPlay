#!/usr/bin/env python3
# verify-submodules.py
# Check .gitmodules has all required submodules with correct paths and urls

REQUIRED = {
    "vasic-digital/Auth", "vasic-digital/Cache", "vasic-digital/Challenges",
    "vasic-digital/Concurrency", "vasic-digital/Containers", "vasic-digital/Database",
    "vasic-digital/Discovery", "vasic-digital/EventBus", "vasic-digital/Formatters",
    "vasic-digital/HelixQA", "vasic-digital/Media", "vasic-digital/Memory",
    "vasic-digital/Messaging", "vasic-digital/Middleware", "vasic-digital/Observability",
    "vasic-digital/Plugins", "vasic-digital/RAG", "vasic-digital/RateLimiter",
    "vasic-digital/Recovery", "vasic-digital/Security", "vasic-digital/Storage",
    "vasic-digital/Streaming", "vasic-digital/VectorDB",
    "HelixDevelopment/Catalogizer", "HelixDevelopment/HelixQA"
}

# Parse .gitmodules
import re
with open('.gitmodules', 'r') as f:
    content = f.read()

# Extract repo paths from urls (e.g., git@github.com:vasic-digital/Auth.git -> vasic-digital/Auth)
found = set()
for match in re.finditer(r'url = git@github\.com:(.+)\.git', content):
    found.add(match.group(1).strip())

missing = REQUIRED - found

if missing:
    print(f"MISSING submodules: {missing}")
    print(f"Found: {len(found)} submodules")
    print(f"Missing: {len(missing)} submodules")
    exit(1)
else:
    print(f"All {len(REQUIRED)} submodules present")
    exit(0)
