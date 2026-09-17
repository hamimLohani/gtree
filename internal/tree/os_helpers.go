package tree

import "os"

// osUserHomeDir is a variable so tests can override it.
var osUserHomeDir = func() string {
	home, _ := os.UserHomeDir()
	return home
}
