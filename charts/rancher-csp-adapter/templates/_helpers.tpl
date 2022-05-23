{{- define "system_default_registry" -}}
{{- if .Values.global.cattle.systemDefaultRegistry -}}
{{- printf "%s/" .Values.global.cattle.systemDefaultRegistry -}}
{{- else -}}
{{- "" -}}
{{- end -}}
{{- end -}}

{{- define "csp-adapter.labels" -}}
app: rancher-csp-adapter
{{- end }}

{{- define "csp-adapter.outputConfigMap" -}}
csp-config
{{- end }}

{{- define "csp-adapter.outputNotification" -}}
csp-compliance
{{- end }}

{{- define "csp-adapter.cacheSecret" -}}
csp-adapter-cache
{{- end }}

{{- define "csp-adapter.hostnameSetting" -}}
server-url
{{- end }}

{{- define "csp-adapter.versionSetting" -}}
server-version
{{- end }}
