# TDL Versioning Guide

This document explains how version numbers are managed in TDL and how contributors should decide when to bump versions.

## Semantic Versioning (SemVer)

TDL follows **Semantic Versioning**: `MAJOR.MINOR.PATCH`

### 1. MAJOR version

- Increment for **breaking changes**.
- Examples:
    - Removing or renaming a flag or subcommand.
    - Changing default behavior that breaks existing scripts.
- Example: `1.1.0 → 2.0.0`

### 2. MINOR version

- Increment for **new features in a backward-compatible way**.
- Examples:
    - Adding new subcommands (`init`, `destroy`).
    - Adding new optional flags.
    - Enhancements that don’t break existing usage.
- Example: `1.0.0 → 1.1.0`

### 3. PATCH version

- Increment for **bug fixes and minor improvements**.
- Examples:
    - Fixing crashes or typos.
    - Performance improvements.
    - Small output formatting changes that don’t add features.
- Example: `1.1.0 → 1.1.1`

### Notes for contributors

- Always check if your change is **backward-compatible**.
- For breaking changes, bump the **MAJOR** version.
- For new features that don’t break anything, bump the **MINOR** version.
- For small fixes or improvements, bump the **PATCH** version.
- Use Git tags for official releases:
    ```bash
    git tag -a v1.1.0 -m "Description of release"
    git push origin v1.1.0
    ```

