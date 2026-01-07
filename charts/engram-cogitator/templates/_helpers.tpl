{{/* charts/engram-cogitator/templates/_helpers.tpl */}}

{{/*
Expand the name of the chart.
*/}}
{{- define "engram-cogitator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "engram-cogitator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "engram-cogitator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "engram-cogitator.labels" -}}
helm.sh/chart: {{ include "engram-cogitator.chart" . }}
{{ include "engram-cogitator.selectorLabels" . }}
app.kubernetes.io/version: {{ .Values.image.tag | default .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "engram-cogitator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "engram-cogitator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "engram-cogitator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "engram-cogitator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
API component name
*/}}
{{- define "engram-cogitator.api.name" -}}
{{- printf "%s-api" (include "engram-cogitator.fullname" .) }}
{{- end }}

{{/*
Ollama component name
*/}}
{{- define "engram-cogitator.ollama.name" -}}
{{- printf "%s-ollama" (include "engram-cogitator.fullname" .) }}
{{- end }}

{{/*
Postgres service name - uses Bitnami naming when internal
*/}}
{{- define "engram-cogitator.postgres.name" -}}
{{- if .Values.storage.postgres.internal }}
{{- printf "%s-postgresql" .Release.Name }}
{{- else }}
{{- .Values.storage.postgres.host }}
{{- end }}
{{- end }}

{{/*
Postgres DSN
*/}}
{{- define "engram-cogitator.postgres.dsn" -}}
{{- if .Values.storage.postgres.internal }}
{{- printf "postgres://%s:$(POSTGRES_PASSWORD)@%s:5432/%s?sslmode=disable" .Values.postgresql.auth.username (include "engram-cogitator.postgres.name" .) .Values.postgresql.auth.database }}
{{- else }}
{{- printf "postgres://%s:$(POSTGRES_PASSWORD)@%s:%d/%s?sslmode=disable" .Values.storage.postgres.username .Values.storage.postgres.host (.Values.storage.postgres.port | int) .Values.storage.postgres.database }}
{{- end }}
{{- end }}

{{/*
Postgres secret name
*/}}
{{- define "engram-cogitator.postgres.secretName" -}}
{{- if .Values.storage.postgres.internal }}
{{- printf "%s-postgresql" .Release.Name }}
{{- else }}
{{- required "storage.postgres.existingSecret is required when using external postgres" .Values.storage.postgres.existingSecret }}
{{- end }}
{{- end }}

{{/*
Ollama URL - internal service or external URL
*/}}
{{- define "engram-cogitator.ollama.url" -}}
{{- if .Values.ollama.enabled }}
{{- printf "http://%s:11434" (include "engram-cogitator.ollama.name" .) }}
{{- else }}
{{- required "ollama.externalUrl is required when ollama.enabled is false" .Values.ollama.externalUrl }}
{{- end }}
{{- end }}
