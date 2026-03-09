package assets

import _ "embed"

//go:embed favicon.svg
var Favicon []byte

//go:embed index.html
var WebUI []byte
