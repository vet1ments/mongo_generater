package internal

import _ "embed"

//go:embed templates/embed_model.go.tmpl
var EmbedTemplateBinary string

//go:embed templates/model.go.tmpl
var ModelTemplateBinary string
