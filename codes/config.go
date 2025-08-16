package codes

import "strings"

// singleLineCommentMap defines the mapping between comment characters and supported file extensions.
var singleLineCommentMap = map[string][]string{
	"//": {".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala",
		".ts", ".js", ".jsx", ".tsx"},
	"#": {".py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk",
		"makefile", "dockerfile", ".ini"},
	";":   {".lisp", ".clj", ".scm", ".s", ".asm"},
	"--":  {".lua", ".hs", ".sql", ".adb"},
	"'":   {".vb", ".vbs"},
	".. ": {".rst"},
}

// SupportedTags defines the set of recognized comment tags.
var SupportedTags = map[string]struct{}{
	"TODO":      {},
	"FIXME":     {},
	"NOTE":      {},
	"HACK":      {},
	"BUG":       {},
	"OPTIMIZE":  {},
	"DEPRECATE": {},
}

// extensionToCommentChar maps a file extension to its comment character for quick lookup.
var extensionToCommentChar map[string]string

// init initializes the extensionToCommentChar mapping.
func init() {
	extensionToCommentChar = make(map[string]string)
	for commentChar, extensions := range singleLineCommentMap {
		for _, extension := range extensions {
			extensionToCommentChar[strings.ToLower(extension)] = commentChar
		}
	}
}

