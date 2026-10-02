{{/*
Common labels
*/}}
{{- define "tmplsvc.labels" -}}
app: tmplsvc
app.kubernetes.io/name: tmplsvc
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version | replace "+" "_" }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tmplsvc.selectorLabels" -}}
app: tmplsvc
app.kubernetes.io/name: tmplsvc
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Database environment variables, shared by the service and the init container.
*/}}
{{- define "tmplsvc.dbEnvVars" -}}
- name: TMPLSVC_DB_USER
  valueFrom:
    configMapKeyRef:
      name: tmplsvc-config
      key: db_user
- name: TMPLSVC_DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: tmplsvc-secret
      key: db_password
- name: TMPLSVC_DB_HOST
  valueFrom:
    configMapKeyRef:
      name: tmplsvc-config
      key: db_host
- name: TMPLSVC_DB_NAME
  valueFrom:
    configMapKeyRef:
      name: tmplsvc-config
      key: db_name
- name: TMPLSVC_DB_DISABLE_TLS
  valueFrom:
    configMapKeyRef:
      name: tmplsvc-config
      key: db_disabletls
{{- end -}}

{{/*
Kubernetes metadata environment variables, reported by /v1/liveness.
*/}}
{{- define "tmplsvc.k8sEnvVars" -}}
- name: KUBERNETES_NAMESPACE
  valueFrom:
    fieldRef:
      fieldPath: metadata.namespace
- name: KUBERNETES_NAME
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: KUBERNETES_POD_IP
  valueFrom:
    fieldRef:
      fieldPath: status.podIP
- name: KUBERNETES_NODE_NAME
  valueFrom:
    fieldRef:
      fieldPath: spec.nodeName
{{- end -}}
