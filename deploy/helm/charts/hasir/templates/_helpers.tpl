{{/*
Expand the name of the chart.
*/}}
{{- define "hasir.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "hasir.fullname" -}}
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
{{- define "hasir.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "hasir.labels" -}}
helm.sh/chart: {{ include "hasir.chart" . }}
{{ include "hasir.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "hasir.selectorLabels" -}}
app.kubernetes.io/name: {{ include "hasir.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the API service.
*/}}
{{- define "hasir.api.fullname" -}}
{{- printf "%s-api" (include "hasir.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Create the name of the Dashboard service.
*/}}
{{- define "hasir.dashboard.fullname" -}}
{{- printf "%s-dashboard" (include "hasir.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Create the name of the PostgreSQL subchart.
*/}}
{{- define "hasir.postgresql.fullname" -}}
{{- printf "%s-postgresql" .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end }}
