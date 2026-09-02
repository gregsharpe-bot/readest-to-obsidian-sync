{{- define "readest-obsidian-sync.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "readest-obsidian-sync.fullname" -}}
{{- if .Values.fullnameOverride }}{{ .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}{{ else }}{{ include "readest-obsidian-sync.name" . }}{{ end }}
{{- end }}

{{- define "readest-obsidian-sync.serviceAccountName" -}}
{{- default (include "readest-obsidian-sync.fullname" .) .Values.serviceAccount.name }}
{{- end }}
