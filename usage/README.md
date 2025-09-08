# TDL Usage Guide

`tdl` is a command-line tool to manage your project’s `.tdl` directory and scan your codebase for tagged comments (TODOs, FIXMEs, etc.).

---

## Commands

### Initialize a `.tdl` directory

```bash
tdl init
```

- Creates a `.tdl` directory in the current working directory.
- If it already exists, TDL will notify you and do nothing.

---

### Destroy the `.tdl` directory

```bash
tdl destroy
```

- Deletes the `.tdl` directory and all its contents.
- Prompts for confirmation before deletion:

```
Are you sure you want to destroy '.tdl'? (y/N):
```

- Any response other than `y`, `Y`, `yes`, or `YES` aborts the deletion.

---

### Scan your codebase for tagged comments

```bash
tdl scan [flags]
```

- Recursively scans a directory for comments containing tags:
  `TODO`, `FIXME`, `NOTE`, `HACK`, `BUG`, `OPTIMIZE`, `DEPRECATE`.
- Saves results in `.tdl/comments.json` by default.
- Optionally pretty-prints results with colors and displays stats.

---

## Scan Flags

| Flag       | Type   | Default             | Description                                                 |
| ---------- | ------ | ------------------- | ----------------------------------------------------------- |
| `-dirpath` | string | `.`                 | Directory to recursively scan.                              |
| `-tag`     | string | All supported       | Comma-separated tags to filter by (e.g., `TODO,FIXME`).     |
| `-color`   | bool   | `true`              | Enable colorized output.                                    |
| `-ignore`  | bool   | `true`              | Skip unsupported or binary files silently.                  |
| `-workers` | int    | Number of CPU cores | Number of concurrent worker goroutines for faster scanning. |
| `-print`   | bool   | `false`             | Pretty-print results after scanning.                        |

> Notes: Output is always saved to `.tdl/comments.json. YAML or text output is not currently supported in CLI flags. For custom formats, see `core.PrepareOutputFile\` usage in code.

---

## Examples

### Initialize `.tdl`

```bash
tdl init
```

### Destroy `.tdl` (with confirmation)

```bash
tdl destroy
```

### Scan current directory for all tags

```bash
tdl scan
```

### Scan a specific directory for TODO and FIXME

```bash
tdl scan -dirpath ./myproject -tag TODO,FIXME
```

### Disable colored output

```bash
tdl scan -color=false
```

### Use 4 concurrent workers

```bash
tdl scan -workers=4
```

### Pretty-print results immediately after scanning

```bash
tdl scan -print
```

---

## Notes

- **Supported file types** include Go, Python, JavaScript, C, C++, Java, Lua, Bash, YAML, and more. See `core.go` `singleLineCommentMap` for the full mapping.
- Git blame metadata (author, commit, timestamp) is automatically attached to each comment.
- Large projects benefit from increasing worker count, but spawning too many may overload the system.
- Comments are grouped and sorted by file and line number for easy reading.

---

## Example Output

```
File: ./main.go
    12  // TODO: Implement error handling
    45  // FIXME: Optimize this loop
```

- Colored output highlights tags for quick scanning:

| Tag       | Color      |
| --------- | ---------- |
| TODO      | Yellow     |
| FIXME     | Red        |
| NOTE      | Cyan       |
| HACK      | Magenta    |
| BUG       | Bright Red |
| OPTIMIZE  | Green      |
| DEPRECATE | Gray       |

