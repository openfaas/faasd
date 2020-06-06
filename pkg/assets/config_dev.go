// +build dev

package assets

import "net/http"

// Config contains the faasd configuration assets
var Config http.FileSystem = http.Dir("config")
