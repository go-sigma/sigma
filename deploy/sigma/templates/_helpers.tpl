{{- define "sigma.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "sigma.fullname" -}}
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

{{- define "sigma.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "sigma.labels" -}}
helm.sh/chart: {{ include "sigma.chart" . }}
{{ include "sigma.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "sigma.server.labels" -}}
app.kubernetes.io/name: {{ template "sigma.server" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "sigma.worker.labels" -}}
app.kubernetes.io/name: {{ template "sigma.worker" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "sigma.distribution.labels" -}}
app.kubernetes.io/name: {{ template "sigma.distribution" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "sigma.web.labels" -}}
app.kubernetes.io/name: {{ template "sigma.web" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "sigma.selectorLabels" -}}
app.kubernetes.io/name: {{ include "sigma.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "sigma.server" -}}
  {{- printf "%s-server" (include "common.names.fullname" .) -}}
{{- end -}}

{{- define "sigma.worker" -}}
  {{- printf "%s-worker" (include "common.names.fullname" .) -}}
{{- end -}}

{{- define "sigma.distribution" -}}
  {{- printf "%s-distribution" (include "common.names.fullname" .) -}}
{{- end -}}

{{- define "sigma.web" -}}
  {{- printf "%s-web" (include "common.names.fullname" .) -}}
{{- end -}}
