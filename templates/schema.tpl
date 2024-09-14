{{ define "file" }}
package {{ .PackageName }}

import (
	"encoding/json"
	"time"
)

{{ range .Schemas }}
{{- if eq .Type "struct" }}
{{ template "struct" . }}
{{- else if eq .Type "enum" }}
{{ template "enum" . }}
{{- else }}
{{ template "type" . }}
{{- end }}
{{ end }}

{{ end }}

{{ define "type" }}
{{- if .Description }}
{{ template "description" .Description }}
{{- end }}
type {{ .Name }} {{ .Type }}
{{ end }}

{{ define "enum" }}
{{- if .Description }}
{{ template "description" .Description }}
{{- end }}
type {{ .Name }} string

const (
{{- range .Values }}
	{{ . }} {{ $.Name }} = "{{ . }}"
{{- end }}
)
{{ end }}

{{ define "struct" }}
{{- if .Description -}}
{{ template "description" .Description }}
{{- end -}}
type {{ .Name }} struct {
{{- range .Properties }}
{{- template "field" . -}}
{{ end -}}
}
{{ end }}

{{ define "field" }}
{{- if .Description -}}
	{{ template "description" .Description }}
{{- end -}}
	{{ .Name }} {{ .Type }} `json:"{{ .JSONName }}{{ if .OmitEmpty }},omitempty{{ end }}"`
{{ end }}

{{ define "description" }}
{{- $lines := splitLines . -}}
{{- range $index, $line := $lines }}
// {{ $line }}
{{- end }}
{{ end }}
