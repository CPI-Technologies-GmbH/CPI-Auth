{{/*
Expand the name of the chart.
*/}}
{{- define "cpi-auth.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "cpi-auth.fullname" -}}
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
{{- define "cpi-auth.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cpi-auth.labels" -}}
helm.sh/chart: {{ include "cpi-auth.chart" . }}
{{ include "cpi-auth.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cpi-auth.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cpi-auth.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Service account name
*/}}
{{- define "cpi-auth.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "cpi-auth.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Database host - returns internal service or external host
*/}}
{{- define "cpi-auth.databaseHost" -}}
{{- if .Values.postgresql.external.enabled }}
{{- .Values.postgresql.external.host }}
{{- else }}
{{- printf "%s-postgresql" (include "cpi-auth.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Database port
*/}}
{{- define "cpi-auth.databasePort" -}}
{{- if .Values.postgresql.external.enabled }}
{{- .Values.postgresql.external.port | toString }}
{{- else }}
{{- "5432" }}
{{- end }}
{{- end }}

{{/*
Database name
*/}}
{{- define "cpi-auth.databaseName" -}}
{{- if .Values.postgresql.external.enabled }}
{{- .Values.postgresql.external.database }}
{{- else }}
{{- .Values.postgresql.auth.database }}
{{- end }}
{{- end }}

{{/*
Database user
*/}}
{{- define "cpi-auth.databaseUser" -}}
{{- if .Values.postgresql.external.enabled }}
{{- .Values.postgresql.external.username }}
{{- else }}
{{- .Values.postgresql.auth.username }}
{{- end }}
{{- end }}

{{/*
Redis host
*/}}
{{- define "cpi-auth.redisHost" -}}
{{- if .Values.redis.external.enabled }}
{{- printf "%s:%s" .Values.redis.external.host (.Values.redis.external.port | toString) }}
{{- else }}
{{- printf "%s-redis:6379" (include "cpi-auth.fullname" .) }}
{{- end }}
{{- end }}

{{/*
NATS URL
*/}}
{{- define "cpi-auth.natsURL" -}}
{{- if .Values.nats.external.enabled }}
{{- .Values.nats.external.url }}
{{- else }}
{{- printf "nats://%s-nats:4222" (include "cpi-auth.fullname" .) }}
{{- end }}
{{- end }}

{{/*
Image tag - defaults to appVersion
*/}}
{{- define "cpi-auth.imageTag" -}}
{{- .tag | default $.Chart.AppVersion }}
{{- end }}

{{/*
Database password - resolves from external or internal config
*/}}
{{- define "cpi-auth.databasePassword" -}}
{{- if .Values.postgresql.external.enabled }}
{{- required "postgresql.external.password is required when using external PostgreSQL" .Values.postgresql.external.password }}
{{- else }}
{{- default (randAlphaNum 24) .Values.postgresql.auth.password }}
{{- end }}
{{- end }}
