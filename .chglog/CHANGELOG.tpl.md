{{ range .Versions }}
<a name="{{ .Tag.Name }}"></a>
## {{ if .Tag.Previous }}[{{ .Tag.Name }}]({{ $.Info.RepositoryURL }}/compare/{{ .Tag.Previous.Name }}...{{ .Tag.Name }}){{ else }}{{ .Tag.Name }}{{ end }}

> {{ datetime "2006-01-02" .Tag.Date }}

{{ range .CommitGroups -}}
{{- if not (or (eq .Title "Version") (eq .Title "Ci") (eq .Title "Pipeline") (eq .Title "Chglog")) -}}
### {{ .Title }}

{{ range .Commits -}}
* {{ .Subject }}
{{ end }}
{{ end -}}
{{ end -}}

{{- if .RevertCommits -}}
### Reverts

{{ range .RevertCommits -}}
* {{ .Revert.Header }}
{{ end }}
{{ end -}}

{{ end -}}
