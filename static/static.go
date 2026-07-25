package static

import (
	"embed"
)

// Files embeds all static web assets (HTML, CSS, JS) into the Go binary
//
//go:embed *
var Files embed.FS
