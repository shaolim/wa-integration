package web

import _ "embed"

//go:embed login.html
var LoginHTML string

//go:embed files.html
var FilesHTML string

//go:embed file_detail.html
var FileDetailHTML string
