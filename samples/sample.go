package samples

import "fmt"

// [TODO] Add proper error handling
func loadConfig() {
	// loading config...
}

// [FIXME] Panics if config file is missing
func parseConfig(path string) {
	// ...
}

// [HACK] Hardcoding default path for now
var defaultPath = "/etc/app/config.json"

// [BUG] Off-by-one error in loop
func count() {
	for i := 0; i <= 10; i++ {
		fmt.Println(i)
	}
}

// [NOTE] This is just a sample function
func greet(name string) {
	fmt.Println("Hello,", name)
}

// [OPTIMIZE] Use a map instead of slice lookup
var users = []string{"alice", "bob", "charlie"}

// [DEPRECATED] Use NewUser() instead
func OldUser() {
	// old implementation
}
