package codes

// Comment represents a tagged comment found in a file.
type Comment struct {
	Tag        string // The tag type (TODO, FIXME, etc.)
	Content    string // The full comment content including the tag
	FilePath   string // Absolute or relative path of the file
	LineNumber int    // Line number in the file where comment was found
}

