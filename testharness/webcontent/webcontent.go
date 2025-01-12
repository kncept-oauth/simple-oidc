package webcontent

import "embed"

//go:embed *.html
var Views embed.FS

//go:embed *.css
var Static embed.FS
