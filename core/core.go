package core

import (
	"strings"
)

// init builds lookup maps for faster scanning at runtime.
func init() {
	// Build set of supported tags for O(1) lookups.
	for _, tag := range SupportedTags {
		supportedTagsLookup[tag] = struct{}{}
	}
	// Map each extension to its comment delimiter (e.g. ".go" -> "//")
	for ch, exts := range singleLineCommentMap {
		for _, e := range exts {
			extensionToChar[strings.ToLower(e)] = ch
		}
	}
}

