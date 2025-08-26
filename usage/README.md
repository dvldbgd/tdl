# TDL Usage Guide

This document explains how to use the `tdl` command-line tool to manage your project and scan your codebase for tagged comments.

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
- Prompts for confirmation before deletion.
- Example confirmation:

```
Are you sure you want to destroy '.tdl'? (y/N):
```

---

### Scan your codebase for tagged comments

```bash
tdl scan [options]
```

- Recursively scans files in a specified directory.
- Extracts comments containing tags: `TODO`, `FIXME`, `NOTE`, `HACK`, `BUG`, `OPTIMIZE`, `DEPRECATE`.

---

## Scan Options (Flags)

| Flag         | Type   | Default             | Description                                                        |
| ------------ | ------ | ------------------- | ------------------------------------------------------------------ |
| `-dirpath`   | string | `.`                 | Directory to recursively scan. Accepts relative or absolute paths. |
| `-tag`       | string | All supported       | Comma-separated list of tags to filter (e.g., `TODO,FIXME`).       |
| `-color`     | bool   | `true`              | Enable colorized output for easier readability.                    |
| `-ignore`    | bool   | `true`              | Skip unsupported file extensions silently without printing errors. |
| `-workers`   | int    | Number of CPU cores | Number of concurrent worker goroutines for faster scanning.        |
| `-output`    | string | Empty               | File format for saving results (`json`, `yaml`, `text`).           |
| `-outputdir` | string | `.`                 | Directory path to save the output file.                            |
| `-summarize` | bool   | `false`             | Prints the frequency of each tag in the scanned codebase.          |

---

## Examples

### Initialize the `.tdl` directory

```bash
tdl init
```

### Destroy the `.tdl` directory (with confirmation)

```bash
tdl destroy
```

### Scan the current directory for all supported tags

```bash
tdl scan
```

### Scan a specific directory for TODO and FIXME comments only

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

### Save results to a JSON file

```bash
tdl scan -output=json -outputdir=reports
```

### Show a summary of all tags in the codebase

```bash
tdl scan -summarize
```

---

## Notes

- **Supported file types** include Go, Python, JavaScript, C, C++, Java, Lua, Bash, YAML, and more. See `core.go` for the full mapping in `singleLineCommentMap`.
- `.gitignore` rules are respected only for **untracked** files; tracked files will still be scanned.
- For large projects, increasing the number of workers can improve performance, but too many may overload the system.

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

