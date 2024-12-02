{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "ggbridge.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Expand the name of the chart.
*/}}
{{- define "ggbridge.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "ggbridge.fullname" -}}
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
Create a default fully qualified client name.
{{ include "ggbridge.client.name" . }}
*/}}
{{- define "ggbridge.client.fullname" -}}
{{- printf "%s-client" (include "ggbridge.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Create a default fully qualified server name.
{{ include "ggbridge.server.name" }}
*/}}
{{- define "ggbridge.server.fullname" -}}
{{- printf "%s-server" (include "ggbridge.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Create a default fully qualified proxy name.
{{ include "ggbridge.proxy.fullname" }}
*/}}
{{- define "ggbridge.proxy.fullname" -}}
{{- printf "%s-proxy" (include "ggbridge.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "ggbridge.labels" -}}
helm.sh/chart: {{ include "ggbridge.chart" . }}
{{ include "ggbridge.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
tenant: {{ .Values.subdomain }}
{{- with .Values.commonLabels }}
{{ tpl (toYaml .) $ }}
{{- end -}}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "ggbridge.selectorLabels" -}}
app.kubernetes.io/name: {{ include "ggbridge.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Client labels
{{ include "ggbridge.client.labels" . }}
*/}}
{{- define "ggbridge.client.labels" -}}
{{ include "ggbridge.client.selectorLabels" . }}
{{- with .Values.labels }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end }}

{{/*
Client selector labels
{{ include "ggbridge.client.selectorLabels" . }}
*/}}
{{- define "ggbridge.client.selectorLabels" -}}
app.kubernetes.io/component: client
{{- end }}

{{/*
Server labels
{{ include "ggbridge.server.labels" . }}
*/}}
{{- define "ggbridge.server.labels" -}}
{{ include "ggbridge.server.selectorLabels" . }}
{{- with .Values.labels }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end }}

{{/*
Server selector labels
{{ include "ggbridge.server.selectorLabels" . }}
*/}}
{{- define "ggbridge.server.selectorLabels" -}}
app.kubernetes.io/component: server
{{- end }}

{{/*
Proxy labels
{{ include "ggbridge.proxy.labels" . }}
*/}}
{{- define "ggbridge.proxy.labels" -}}
{{ include "ggbridge.proxy.selectorLabels" . }}
{{- with .Values.proxy.labels }}
{{ tpl (toYaml .) $ }}
{{- end }}
{{- end }}

{{/*
Proxy selector labels
{{ include "ggbridge.proxy.selectorLabels" . }}
*/}}
{{- define "ggbridge.proxy.selectorLabels" -}}
app.kubernetes.io/component: proxy
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "ggbridge.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "ggbridge.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Return the proper image name.
If image tag and digest are not defined, termination fallbacks to chart appVersion.
{{ include "ggbridge.image" }}
*/}}
{{- define "ggbridge.image" -}}
{{- $registryName := .Values.image.registry -}}
{{- $repositoryName := .Values.image.repository -}}
{{- $separator := ":" -}}
{{- $termination := .Values.image.tag | toString -}}

{{- if not .Values.image.tag }}
  {{- if .Chart }}
    {{- $termination = .Chart.AppVersion | toString -}}
  {{- end -}}
{{- end -}}
{{- if .Values.image.digest }}
    {{- $separator = "@" -}}
    {{- $termination = .Values.image.digest | toString -}}
{{- end -}}
{{- if $registryName }}
    {{- printf "%s/%s%s%s" $registryName $repositoryName $separator $termination -}}
{{- else -}}
    {{- printf "%s%s%s"  $repositoryName $separator $termination -}}
{{- end -}}
{{- end -}}

{{/*
Return the proper image name.
If image tag and digest are not defined, termination fallbacks to chart appVersion.
{{ include "ggbridge.proxy.image" }}
*/}}
{{- define "ggbridge.proxy.image" -}}
{{- $registryName := .Values.proxy.image.registry -}}
{{- $repositoryName := .Values.proxy.image.repository -}}
{{- $separator := ":" -}}
{{- $termination := .Values.proxy.image.tag | toString -}}

{{- if not .Values.proxy.image.tag }}
  {{- if .Chart }}
    {{- $termination = .Chart.AppVersion | toString -}}
  {{- end -}}
{{- end -}}
{{- if .Values.proxy.image.digest }}
    {{- $separator = "@" -}}
    {{- $termination = .Values.proxy.image.digest | toString -}}
{{- end -}}
{{- if $registryName }}
    {{- printf "%s/%s%s%s" $registryName $repositoryName $separator $termination -}}
{{- else -}}
    {{- printf "%s%s%s"  $repositoryName $separator $termination -}}
{{- end -}}
{{- end -}}

{{/*
Returns hostname
{{ include "ggbridge.hostname" $ }}
*/}}
{{- define "ggbridge.hostname" -}}
{{ printf "%s.%s" .Values.subdomain .Values.domain }}
{{- end }}

{{/*
Returns proxy service name
{{ include "ggbridge.proxy.serviceName" $ }}
*/}}
{{- define "ggbridge.proxy.serviceName" -}}
{{- $name := .Values.proxy.service.name -}}
{{- if not $name -}}
{{ ternary .Values.subdomain (include "ggbridge.proxy.fullname" .) (eq .Values.mode "server") }}
{{- end -}}
{{- end -}}

{{/*
Returns true when proxy is enabled
{{ include "ggbridge.proxy.enabled" $ }}
*/}}
{{- define "ggbridge.proxy.enabled" -}}
{{- $result := "false" -}}
{{- $ports := ternary .Values.server.tunnels .Values.client.tunnels (eq .Values.mode "server") -}}
{{- if $ports -}}
  {{- $result = "true" -}}
{{- end -}}
{{ $result }}
{{- end -}}

{{/*
Returns pod affinity.
{{ include "ggbridge.affinity" (dict "index" 0 "context" $) }}
*/}}
{{- define "ggbridge.affinity" -}}
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - {{ .context.Values.mode }}
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
        topologyKey: "topology.kubernetes.io/zone"
    - weight: 10
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - {{ .context.Values.mode }}
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
        topologyKey: "kubernetes.io/hostname"
podAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - {{ .context.Values.mode }}
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
            - key: index
              operator: In
              values:
                - {{ .index | quote }}
        topologyKey: "topology.kubernetes.io/zone"
{{- end -}}

{{/*
Returns proxy pod affinity.
{{ include "ggbridge.proxy.affinity" (dict "index" 0 "context" $) }}
*/}}
{{- define "ggbridge.proxy.affinity" -}}
podAntiAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - proxy
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
        topologyKey: "topology.kubernetes.io/zone"
    - weight: 10
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - proxy
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
        topologyKey: "kubernetes.io/hostname"
podAffinity:
  preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - {{ .context.Values.mode }}
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
            - key: index
              operator: In
              values:
                - {{ .index | quote }}
        topologyKey: "topology.kubernetes.io/zone"
    - weight: 10
      podAffinityTerm:
        labelSelector:
          matchExpressions:
            - key: app.kubernetes.io/component
              operator: In
              values:
                - {{ .context.Values.mode }}
            - key: tenant
              operator: In
              values:
                - {{ .context.Values.subdomain }}
            - key: index
              operator: In
              values:
                - {{ .index | quote }}
        topologyKey: "kubernetes.io/hostname"
{{- end -}}

{{/*
Returns server service annotations
{{ include "ggbridge.server.service.annotations" $ }}
*/}}
{{- define "ggbridge.server.service.annotations" -}}
{{- $annotations := dict -}}
{{- if or .Values.commonAnnotations .Values.server.service.annotations }}
{{- $annotations = include "ggbridge.tplvalues.merge" ( dict "values" ( list .Values.server.service.annotations .Values.commonAnnotations ) "context" . ) }}
{{- end -}}
{{- if eq .Values.server.ingress.controller "traefik" -}}
  {{- if and .Values.tls.enabled (eq (lower .Values.tls.mode) "passthrough") -}}
    {{- $_ := set $annotations "traefik.ingress.kubernetes.io/service.serversscheme" "https" -}}
  {{- end -}}
{{- end -}}
{{ include "ggbridge.tplvalues.render" ( dict "value" $annotations "context" .) }}
{{- end -}}

{{/*
Returns server ingress annotations
{{ include "ggbridge.server.ingress.annotations" $ }}
*/}}
{{- define "ggbridge.server.ingress.annotations" -}}
{{- $annotations := dict -}}
{{- $fullname := include "ggbridge.fullname" . }}
{{- $serverFullname := include "ggbridge.server.fullname" . }}
{{- if eq .Values.server.ingress.controller "traefik" -}}
  {{- if .Values.tls.enabled -}}
    {{- $_ := set $annotations "traefik.ingress.kubernetes.io/router.entrypoints" "websecure" -}}
    {{- $_ := set $annotations "traefik.ingress.kubernetes.io/router.tls.options" (printf "%s-%s@kubernetescrd" .Release.Namespace $serverFullname ) -}}
  {{- else -}}
    {{- $_ := set $annotations "traefik.ingress.kubernetes.io/router.entrypoints" "web" -}}
  {{- end -}}
{{- else if eq .Values.server.ingress.controller "nginx" -}}
  {{- if .Values.tls.enabled -}}
    {{- $_ := set $annotations "nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream" "false" -}}
    {{- if eq (lower .Values.tls.mode) "passthrough" -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/backend-protocol" "HTTPS" -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/ssl-passthrough" "true" -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/ssl-redirect" "true" -}}
    {{- else -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/backend-protocol" "HTTP" -}}
    {{- end -}}
    {{- if eq (lower .Values.tls.mode) "mutual" -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/auth-tls-secret" (printf "%s/%s-crt" .Release.Namespace $serverFullname) -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/auth-tls-verify-client" "on" -}}
      {{- $_ := set $annotations "nginx.ingress.kubernetes.io/auth-tls-verify-depth" "1" -}}
    {{- end -}}
  {{- end -}}
{{- end -}}
{{- $annotations = include "ggbridge.tplvalues.merge" ( dict "values" ( list .Values.server.ingress.annotations $annotations .Values.commonAnnotations ) "context" . ) | fromYaml -}}
{{ include "ggbridge.tplvalues.render" ( dict "value" $annotations "context" .) }}
{{- end -}}

{{/*
Returns proxy ingress annotations
{{ include "ggbridge.proxy.ingress.annotations" $ }}
*/}}
{{- define "ggbridge.proxy.ingress.annotations" -}}
{{- $annotations := dict -}}
{{- if eq .Values.proxy.ingress.controller "nginx" -}}
  {{- $_ := set $annotations "nginx.ingress.kubernetes.io/backend-protocol" "HTTPS" -}}
  {{- $_ := set $annotations "nginx.ingress.kubernetes.io/ssl-passthrough" "true" -}}
  {{- $_ := set $annotations "nginx.ingress.kubernetes.io/ssl-redirect" "true" -}}
{{- else if .Values.proxy.ingress.controller "aws" -}}
  {{- $_ := set $annotations "service.beta.kubernetes.io/aws-load-balancer-type" "nlb" -}}
  {{- $_ := set $annotations "service.beta.kubernetes.io/aws-load-balancer-scheme" "internal" -}}
  {{- $_ := set $annotations "service.beta.kubernetes.io/aws-load-balancer-backend-protocol" "tcp" -}}
  {{- $_ := set $annotations "service.beta.kubernetes.io/aws-load-balancer-ssl-ports" "443" -}}
  {{- $_ := set $annotations "service.beta.kubernetes.io/aws-load-balancer-proxy-protocol" "*" -}}
{{- end -}}
{{- $annotations = include "ggbridge.tplvalues.merge" ( dict "values" ( list .Values.proxy.ingress.annotations $annotations .Values.commonAnnotations ) "context" . ) | fromYaml -}}
{{ include "ggbridge.tplvalues.render" ( dict "value" $annotations "context" .) }}
{{- end -}}

{{/*
Returns gateway tls mode
{{ include "ggbridge.server.gateway.tlsMode" $ }}
*/}}
{{- define "ggbridge.server.gateway.tlsMode" -}}
{{- $tlsMode := "Terminate" -}}
{{- if eq (lower .Values.tls.mode) "passthrough" -}}
{{- $tlsMode = "Passthrough" -}}
{{- end -}}
{{ $tlsMode }}
{{- end -}}

{{/*
Renders a value that contains template perhaps with scope if the scope is present.
Usage:
{{ include "ggbridge.tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $ ) }}
{{ include "ggbridge.tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $ "scope" $app ) }}
*/}}
{{- define "ggbridge.tplvalues.render" -}}
{{- $value := typeIs "string" .value | ternary .value (.value | toYaml) }}
{{- if contains "{{" (toJson .value) }}
  {{- if .scope }}
      {{- tpl (cat "{{- with $.RelativeScope -}}" $value "{{- end }}") (merge (dict "RelativeScope" .scope) .context) }}
  {{- else }}
    {{- tpl $value .context }}
  {{- end }}
{{- else }}
    {{- $value }}
{{- end }}
{{- end -}}

{{/*
Merge a list of values that contains template after rendering them.
Merge precedence is consistent with http://masterminds.github.io/sprig/dicts.html#merge-mustmerge
Usage:
{{ include "ggbridge.tplvalues.merge" ( dict "values" (list .Values.path.to.the.Value1 .Values.path.to.the.Value2) "context" $ ) }}
*/}}
{{- define "ggbridge.tplvalues.merge" -}}
{{- $dst := dict -}}
{{- range .values -}}
{{- $dst = include "ggbridge.tplvalues.render" (dict "value" . "context" $.context "scope" $.scope) | fromYaml | merge $dst -}}
{{- end -}}
{{ $dst | toYaml }}
{{- end -}}

{{/*
Merge a list of values that contains template after rendering them.
Merge precedence is consistent with https://masterminds.github.io/sprig/dicts.html#mergeoverwrite-mustmergeoverwrite
Usage:
{{ include "ggbridge.tplvalues.merge-overwrite" ( dict "values" (list .Values.path.to.the.Value1 .Values.path.to.the.Value2) "context" $ ) }}
*/}}
{{- define "ggbridge.tplvalues.merge-overwrite" -}}
{{- $dst := dict -}}
{{- range .values -}}
{{- $dst = include "ggbridge.tplvalues.render" (dict "value" . "context" $.context "scope" $.scope) | fromYaml | mergeOverwrite $dst -}}
{{- end -}}
{{ $dst | toYaml }}
{{- end -}}
