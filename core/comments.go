package core

// Comment represents one tagged comment (TODO/FIXME/etc.) found in a source file.
// It keeps the tag, the comment content, its location, and Git blame metadata.
type Comment struct {
	Tag           string `json:"tag" yaml:"tag"`         // The tag (TODO, FIXME, etc.)
	Content       string `json:"content" yaml:"content"` // The full comment text
	FilePath      string `json:"file" yaml:"file"`       // Path to the file containing this comment
	LineNumber    int    `json:"line" yaml:"line"`       // Line number in the file
	CreationStamp string `json:"stamp" yaml:"stamp"`     // RFC3339 timestamp from Git blame
	Author        string `json:"author" yaml:"author"`   // Author of the commit that introduced this line
	Commit        string `json:"commit" yaml:"commit"`   // Commit hash from Git blame
}

var (
	// Maps single-line comment delimiters to file extensions that use them.
	// This lets the scanner know how to detect comments in different languages.
	singleLineCommentMap = map[string][]string{
		"//": {
			".go", ".java", ".c", ".cpp", ".h", ".hpp", ".cs", ".swift", ".kt", ".rs", ".scala",
			".ts", ".js", ".jsx", ".tsx",
		},
		"#": {
			".py", ".rb", ".sh", ".bash", ".zsh", ".yml", ".yaml", ".toml", ".pl", ".pm", ".mk",
			"makefile", "dockerfile", ".ini",
		},
		";":  {".lisp", ".clj", ".scm", ".s", ".asm"},
		"--": {".lua", ".hs", ".sql", ".adb"},
		"'":  {".vb", ".vbs"},
		"..": {".rst"},
	}

	// List of supported tags to detect
	SupportedTags       = []string{"TODO", "FIXME", "NOTE", "HACK", "BUG", "OPTIMIZE", "DEPRECATE"}
	supportedTagsLookup = make(map[string]struct{}) // fast lookup map of tags
	extensionToChar     = make(map[string]string)   // maps file extension -> comment delimiter

	// Scanner buffer limit (1MB)
	maxScanCapacity = 1024 * 1024
)

