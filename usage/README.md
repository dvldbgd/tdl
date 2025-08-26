# TDL Usage Guide

This document explains how to use the `tdl` command-line tool to scan your codebase and extract tagged comments.

---

## Command

```bash
tdl [options]
```

`tdl` scans files recursively in a specified directory and extracts comments containing tags such as `TODO`, `FIXME`, `NOTE`, `HACK`, `BUG`, `OPTIMIZE`, and `DEPRECATE`.

---

## Available Flags

| Flag         | Type   | Default             | Description                                                                           |
| ------------ | ------ | ------------------- | ------------------------------------------------------------------------------------- |
| `-dirpath`   | string | `.`                 | Directory to recursively scan. Accepts relative or absolute paths.                    |
| `-tag`       | string | All supported       | Comma-separated list of tags to filter (e.g., `TODO,FIXME`).                          |
| `-color`     | bool   | `true`              | Enable colorized output for easier readability.                                       |
| `-ignore`    | bool   | `true`              | Skip unsupported file extensions silently without printing errors.                    |
| `-workers`   | int    | Number of CPU cores | Number of concurrent worker goroutines for faster scanning.                           |
| `-output`    | string | Empty               | Holds the file format in which the output file containing the comments are written in |
| `-outputdir` | string | Empty               | Holds the path to directory to which the output file must be written                  |

---

## Examples

### Scan the current directory for all supported tags with color

```bash
tdl
```

### Scan a specific directory for TODO and FIXME comments only

```bash
tdl -dirpath ./myproject -tag TODO,FIXME
```

### Disable colored output

```bash
tdl -color false
```

### Use 4 concurrent workers

```bash
tdl -workers 4
```

### Ignore unsupported file errors (default behavior)

```bash
tdl -ignore true
```

---

## Notes

- **Supported file types** include Go, Python, JavaScript, C, C++, Java, Lua, Bash, YAML, and more. The full mapping is defined in `codes/core.go` under `singleLineCommentMap`.
- `.gitignore` rules are respected only if files are **untracked**. Already tracked files will still be scanned.
- For large projects, increasing the number of workers may improve performance, but too many may overload the system.

---

## Example Output

```
File: ./main.go
    12  // TODO: Implement error handling
    45  // FIXME: Optimize this loop
```

Colored output highlights different tags for quick scanning:

- `TODO` → Yellow
- `FIXME` → Red
- `NOTE` → Cyan
- `HACK` → Magenta
- `BUG` → Bright Red
- `OPTIMIZE` → Green
- `DEPRECATE` → Gray

