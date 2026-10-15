{{- define "cart-api.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "cart-api.fullname" -}}
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

{{- define "cart-api.dbName" -}}
{{- printf "%s-db" (include "cart-api.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "cart-api.dbHeadless" -}}
{{- printf "%s-headless" (include "cart-api.dbName" .) }}
{{- end }}

{{- define "cart-api.dbHost" -}}
{{- printf "%s-0.%s.%s.svc.cluster.local" (include "cart-api.dbName" .) (include "cart-api.dbHeadless" .) .Release.Namespace }}
{{- end }}

{{- define "cart-api.secretName" -}}
{{- if .Values.secret.existingSecret }}
{{- .Values.secret.existingSecret }}
{{- else }}
{{- printf "%s-secret" (include "cart-api.fullname" .) }}
{{- end }}
{{- end }}

{{- define "cart-api.configName" -}}
{{- printf "%s-config" (include "cart-api.fullname" .) }}
{{- end }}

{{- define "cart-api.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "cart-api.labels" -}}
helm.sh/chart: {{ include "cart-api.chart" . }}
app.kubernetes.io/name: {{ include "cart-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "cart-api.selectorLabels" -}}
app: {{ include "cart-api.fullname" . }}
app.kubernetes.io/name: {{ include "cart-api.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "cart-api.dbSelectorLabels" -}}
app: {{ include "cart-api.dbName" . }}
app.kubernetes.io/name: {{ include "cart-api.name" . }}-db
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "cart-api.dbStorageClass" -}}
{{- if .Values.db.persistence.storageClass }}
{{- .Values.db.persistence.storageClass }}
{{- else if .Values.db.persistence.staticPv.enabled }}
{{- "manual" }}
{{- end }}
{{- end }}
