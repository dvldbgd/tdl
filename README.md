# TDL

A lightweight command-line tool to scan you codebase and extract `TODO`s, `FIXME`s, `NOTE`s and other tagged comments.
It helps quickly find actionable notes and reminders scattered across the codebase.

---

# Features

- Recursively scan directories for supported file types.
- Extract comments with tags like `TODO`, `FIXME`, `NOTE`, `HACK`, `BUG`, `OPTIMIZE`, `DEPRECATE`.
- Supports multiple languages: Go, Python, JavaScript, C, C++, Java, Lua, and more.
- Concurrent processing for faster scans.
- Optional colorized output for readability.

---

# Installation

Clone this repo and build the CLI:

```bash
git clone git@github.com:dvldbgd/tdl.git
cd tdl
go build -o tdl main.go

mv tdl /usr/local/bin/
```

For detailed usage, see [USAGE](usage.README.md)

