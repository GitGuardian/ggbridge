# ggbridge

![Version: 0.0.0](https://img.shields.io/badge/Version-0.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: unstable](https://img.shields.io/badge/AppVersion-unstable-informational?style=flat-square)

A Helm chart for installing ggbridge

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity for pod assignment |
| annotations | object | `{}` | Set server annotations |
| caBundle.certs | string | `""` | Specify CA certificates to inject (PEM format) |
| caBundle.existingSecret | string | `""` | Specify the secret containing the CA certificate to inject |
| caBundle.existingSecretKey | string | `"ca.crt"` | Specify secret key under the CA certificated is stored |
| caBundle.image.digest | string | `""` | Image digest in the way sha256:aa.... |
| caBundle.image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| caBundle.image.pullSecrets | list | `[]` | Image pull secrets |
| caBundle.image.registry | string | `"ghcr.io"` | Image registry |
| caBundle.image.repository | string | `"gitguardian/ggbridge"` | Image repository |
| caBundle.image.tag | string | `""` | Image tag |
| client.connectionMinIdle | int | `0` | Pool of open connection to the server, in order to speed up the connection process |
| client.readinessProbe.enabled | bool | `true` | Enable Client Readiness Probe |
| client.readinessProbe.exec.command[0] | string | `"ggbridge"` |  |
| client.readinessProbe.exec.command[1] | string | `"healthcheck"` |  |
| client.readinessProbe.exec.command[2] | string | `"-grace-period=60"` |  |
| client.readinessProbe.exec.command[3] | string | `"http://127.0.0.1:9081/healthz"` |  |
| client.readinessProbe.failureThreshold | int | `1` |  |
| client.readinessProbe.initialDelaySeconds | int | `10` |  |
| client.readinessProbe.periodSeconds | int | `15` |  |
| client.readinessProbe.successThreshold | int | `1` |  |
| client.readinessProbe.timeoutSeconds | int | `5` |  |
| client.reverseTunnels.health.enabled | bool | `true` | Enable server to client health tunnel |
| client.reverseTunnels.socks.enabled | bool | `true` | Enable server to client socks tunnel |
| client.reverseTunnels.tls.enabled | bool | `false` | Enable server to client tls tunnel |
| client.reverseTunnels.web.enabled | bool | `false` | Enable server to client web tunnel (for HTTP/HTTPS traffic) |
| client.tunnels.health.enabled | bool | `true` | Enable client to server health tunnel |
| client.tunnels.socks.enabled | bool | `false` | Enable client to server socks tunnel |
| client.tunnels.tls.enabled | bool | `false` | Enable client to server tls tunnel |
| client.tunnels.web.enabled | bool | `false` | Enable client to server web tunnel (for HTTP/HTTPS traffic) |
| clusterDomain | string | `"cluster.local"` | Kubernetes cluster domain |
| commonAnnotations | object | `{}` | Add annotations to all the deployed resources |
| commonLabels | object | `{}` | Add labels to all the deployed resources |
| containerSecurityContext.enabled | bool | `true` | Enable Container security Context in deployments |
| deploymentCount | int | `3` | Number of deployments |
| dnsResolver | string | `""` | Dns resolver to use to lookup ips of domain name |
| domain | string | `"ggbridge.gitguardian.com"` | Domain |
| extraEnv | list | `[]` | Array with extra environment variables # e.g: # extraEnv: #   - name: FOO #     value: "bar" # |
| fullnameOverride | string | `""` | Override the default fully qualified app name |
| hostname | string | `""` | Hostname |
| image.digest | string | `""` | Image digest in the way sha256:aa.... |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.pullSecrets | list | `[]` | Image pull secrets |
| image.registry | string | `"ghcr.io"` | Image registry |
| image.repository | string | `"gitguardian/ggbridge"` | Image repository |
| image.tag | string | `""` | Image tag |
| labels | object | `{}` | Set server labels |
| logLevel | string | `"INFO"` | Set log level |
| mode | string | `"client"` | Deployment mode: "client" or "server" |
| nameOverride | string | `""` | Override the default chart name |
| networkPolicy.allowExternal | bool | `true` | When true, server will accept connections from any source |
| networkPolicy.enabled | bool | `true` | Specifies whether a NetworkPolicy should be created |
| networkPolicy.extraEgress | list | `[]` | Add egress rules to the NetworkPolicy. By default, allow all egress traffic. e.g: extraEgress:   - to:       - ipBlock:           cidr: 0.0.0.0/0 |
| networkPolicy.extraIngress | list | `[]` | Add extra ingress rules to the NetworkPolicy |
| networkPolicy.ingressNSMatchLabels | object | `{}` | Labels to match to allow traffic to the proxy server from other namespaces |
| networkPolicy.ingressNSPodMatchLabels | object | `{}` | Pod labels to match to allow traffic to the proxy server from other namespaces |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| pdb.create | bool | `true` | Enable/disable a Pod Disruption Budget creation |
| pdb.maxUnavailable | string | `""` | Max number of pods that can be unavailable after the eviction |
| pdb.minAvailable | int | `1` | Minimum number of pods that must still be available after the eviction |
| podAnnotations | object | `{}` | This is for setting Kubernetes Annotations to a Pod |
| podLabels | object | `{}` | This is for setting Kubernetes Labels to a Pod |
| podSecurityContext.enabled | bool | `true` | Enable Pod security Context in deployments |
| proxy.affinity | object | `{}` | Affinity for pod assignment |
| proxy.annotations | object | `{}` | Set proxy annotations |
| proxy.labels | object | `{}` | Set proxy labels |
| proxy.networkPolicy.allowExternal | bool | `true` | When true, server will accept connections from any source |
| proxy.networkPolicy.enabled | bool | `true` | Specifies whether a NetworkPolicy should be created |
| proxy.networkPolicy.extraEgress | list | `[]` | Add extra egress rules to the NetworkPolicy |
| proxy.networkPolicy.extraIngress | list | `[]` | Add extra ingress rules to the NetworkPolicy |
| proxy.networkPolicy.ingressNSMatchLabels | object | `{}` | Labels to match to allow traffic to the proxy server from other namespaces |
| proxy.networkPolicy.ingressNSPodMatchLabels | object | `{}` | Pod labels to match to allow traffic to the proxy server from other namespaces |
| proxy.nodeSelector | object | `{}` | Node labels for pod assignment |
| proxy.readinessProbe.enabled | bool | `true` | Whether to enable readiness probe for proxy |
| proxy.readinessProbe.exec.command[0] | string | `"ggbridge"` |  |
| proxy.readinessProbe.exec.command[1] | string | `"healthcheck"` |  |
| proxy.readinessProbe.exec.command[2] | string | `"-pid-file=/var/run/nginx.pid"` |  |
| proxy.readinessProbe.exec.command[3] | string | `"-grace-period=60"` |  |
| proxy.readinessProbe.exec.command[4] | string | `"http://127.0.0.1:9081/healthz"` |  |
| proxy.readinessProbe.failureThreshold | int | `1` |  |
| proxy.readinessProbe.initialDelaySeconds | int | `10` |  |
| proxy.readinessProbe.periodSeconds | int | `15` |  |
| proxy.readinessProbe.successThreshold | int | `1` |  |
| proxy.readinessProbe.timeoutSeconds | int | `5` |  |
| proxy.replicaCount | int | `1` | Number of pods for each deployment |
| proxy.resources.limits | object | `{}` | Set proxy container limits |
| proxy.resources.requests | object | `{"cpu":"50m","memory":"64Mi"}` | Set proxy container requests |
| proxy.service.annotations | object | `{"service.kubernetes.io/topology-mode":"Auto"}` | Set proxy service annotations |
| proxy.service.name | string | `""` | - Set the proxy service name |
| proxy.service.ports.health.containerPort | int | `9081` | Specify the health tunnel port |
| proxy.service.ports.health.exposed | bool | `false` | Defines whether the health port is exposed if service.type is LoadBalancer or NodePort |
| proxy.service.ports.health.port | int | `8081` | Specify the health service port number |
| proxy.service.ports.health.remoteContainerPort | int | `8081` | Specify the health tunnel remote port |
| proxy.service.ports.socks.containerPort | int | `9180` | Specify the socks tunnel port |
| proxy.service.ports.socks.exposed | bool | `true` | Defines whether the socks port is exposed if service.type is LoadBalancer or NodePort |
| proxy.service.ports.socks.port | int | `1080` | Specify the socks service port number |
| proxy.service.ports.tls.containerPort | int | `9443` | Specify the tls tunnel port |
| proxy.service.ports.tls.exposed | bool | `true` | Defines whether the tls port is exposed if service.type is LoadBalancer or NodePort |
| proxy.service.ports.tls.port | int | `443` | Specify the tls service port number |
| proxy.service.ports.tls.remoteContainerPort | int | `8443` | Specify the tls tunnel remote port |
| proxy.service.ports.web.containerPort | int | `9080` | Specify the web tunnel port (nginx) |
| proxy.service.ports.web.exposed | bool | `true` | Defines whether the web port is exposed if service.type is LoadBalancer or NodePort |
| proxy.service.ports.web.internalPort | int | `8080` | Specify the web tunnel internal port (wstunnel) |
| proxy.service.ports.web.port | int | `80` | Specify the web service port number |
| proxy.service.ports.web.remoteContainerPort | int | `8443` | Specify the web tunnel remote port |
| proxy.service.type | string | `"ClusterIP"` | Kubernetes Service type |
| proxy.tolerations | list | `[]` | Tolerations for pod assignment |
| proxy.topologySpreadConstraints | list | `[]` | Topology spread constraints for pod assignment |
| proxy.tunnels.socks.gateway.annotations | object | `{}` | Set gateway annotations |
| proxy.tunnels.socks.gateway.enabled | bool | `false` | Enable socks tunnel exposure using Kubernetes Gateway API |
| proxy.tunnels.socks.gateway.gateway.className | string | `""` | Set the gatewayClassName |
| proxy.tunnels.socks.gateway.gateway.create | bool | `true` | Specifies whether a Gateway resource should be created alongside the routing resource |
| proxy.tunnels.socks.gateway.gateway.ports | object | `{"socks":1080}` | Specify Gateway ports number |
| proxy.tunnels.socks.gateway.parentRefs | list | `[]` | Specify the existing gateway resources |
| proxy.tunnels.socks.service.annotations | object | `{"service.kubernetes.io/topology-mode":"Auto"}` | Specify socks service annotations |
| proxy.tunnels.socks.service.ports.health.exposed | bool | `false` | Defines whether the health port is exposed if service.type is LoadBalancer or NodePort |
| proxy.tunnels.socks.service.ports.health.port | int | `8081` | Specify the health port number |
| proxy.tunnels.socks.service.ports.socks.exposed | bool | `true` | Defines whether the socks port is exposed if service.type is LoadBalancer or NodePort |
| proxy.tunnels.socks.service.ports.socks.port | int | `1080` | Specify the socks port number |
| proxy.tunnels.socks.service.type | string | `"ClusterIP"` | Kubernetes Service type |
| proxy.tunnels.tls.gateway.annotations | object | `{}` | Set gateway annotations |
| proxy.tunnels.tls.gateway.enabled | bool | `false` | Enable tls tunnel exposure using Kubernetes Gateway API |
| proxy.tunnels.tls.gateway.gateway.className | string | `""` | Set the gatewayClassName |
| proxy.tunnels.tls.gateway.gateway.create | bool | `true` | Specifies whether a Gateway resource should be created alongside the routing resource |
| proxy.tunnels.tls.gateway.gateway.listeners | list | `[{"hostname":"api.gitguardian.com"},{"hostname":"hook.gitguardian.com"}]` | Specify tls tunnel listeners |
| proxy.tunnels.tls.gateway.gateway.ports | object | `{"tls":443}` | Specify Gateway ports number |
| proxy.tunnels.tls.gateway.parentRefs | list | `[]` | Specify the existing gateway resources |
| proxy.tunnels.tls.ingress.annotations | object | `{}` | Set ingress annotations |
| proxy.tunnels.tls.ingress.className | string | `""` | Set the ingress ClassName |
| proxy.tunnels.tls.ingress.controller | string | `""` | Specify the ingress controller |
| proxy.tunnels.tls.ingress.enabled | bool | `false` | Enable tls tunnel exposure using Kubernetes Ingress API |
| proxy.tunnels.tls.ingress.listeners | list | `[{"hostname":"api.gitguardian.com"},{"hostname":"hook.gitguardian.com"}]` | Specify tls tunnel listeners |
| proxy.tunnels.tls.service.annotations | object | `{"service.kubernetes.io/topology-mode":"Auto"}` | Specify tls service annotations |
| proxy.tunnels.tls.service.ports.health.exposed | bool | `false` | Defines whether the health port is exposed if service.type is LoadBalancer or NodePort |
| proxy.tunnels.tls.service.ports.health.port | int | `8081` | Specify the health port number |
| proxy.tunnels.tls.service.ports.tls.exposed | bool | `true` | Defines whether the tls port is exposed if service.type is LoadBalancer or NodePort |
| proxy.tunnels.tls.service.ports.tls.port | int | `443` | Specify the tls port number |
| proxy.tunnels.tls.service.type | string | `"ClusterIP"` | Kubernetes Service type |
| proxy.tunnels.web.gateway.annotations | object | `{}` | Set gateway annotations |
| proxy.tunnels.web.gateway.enabled | bool | `false` | Enable web tunnel exposure using Kubernetes Gateway API |
| proxy.tunnels.web.gateway.gateway.className | string | `""` | Set the gatewayClassName |
| proxy.tunnels.web.gateway.gateway.create | bool | `true` | Specifies whether a Gateway resource should be created alongside the routing resource |
| proxy.tunnels.web.gateway.gateway.listeners | list | `[]` | Specify web tunnel listeners # e.g: # listeners: #    - hostname: api.internal.com #      backend: api.gitguardian.com #      tls: #        secretName: "internal-crt" #    - hostname: hook.internal.com #      backend: hook.gitguardian.com #      tls: #        secretName: "internal-crt" |
| proxy.tunnels.web.gateway.gateway.ports | object | `{"http":80,"https":443}` | Specify Gateway ports number |
| proxy.tunnels.web.gateway.parentRefs | list | `[]` | Specify the existing gateway resources |
| proxy.tunnels.web.ingress.annotations | object | `{}` | Set ingress annotations |
| proxy.tunnels.web.ingress.className | string | `""` | Set the ingress ClassName |
| proxy.tunnels.web.ingress.controller | string | `""` | Specify the ingress controller |
| proxy.tunnels.web.ingress.enabled | bool | `false` | Enable web tunnel exposure using Kubernetes Ingress API |
| proxy.tunnels.web.ingress.listeners | list | `[]` | Specify web tunnel listeners # In this example, the following redirection will occur through the web tunnel: # - https://api.internal.com -> https://api.gitguardian.com # - https://hook.internal.com -> https://hook.gitguardian.com # e.g: # listeners: #    - hostname: api.internal.com #      backend: api.gitguardian.com #      tls: #        secretName: "internal-crt" #    - hostname: hook.internal.com #      backend: hook.gitguardian.com #      tls: #        secretName: "internal-crt" |
| proxy.tunnels.web.service.annotations | object | `{"service.kubernetes.io/topology-mode":"Auto"}` | Specify web service annotations |
| proxy.tunnels.web.service.listeners | list | `[]` | Specify web tunnel listeners # Each listener defines a service name that will be used to construct the full internal service DNS. # - `name`: Corresponds to the service name and will be suffixed by `<namespace>.svc.<clusterDomain>`, # making it resolvable within the cluster. # e.g., if `name: api-gitguardian-com` is defined in the `ggbridge` namespace with cluster domain `cluster.local`, # the full DNS would be `api-gitguardian-com.ggbridge.svc.cluster.local`. # # - `backend`: Specifies the external host where the request will be redirected. #   This is typically a public or internal endpoint that the service should forward traffic to. # In this example, the following redirection will occur through the web tunnel: # http://api-gitguardian-com.ggbridge.svc.cluster.local -> https://api.gitguardian.com # # listeners: #   - name: api-gitguardian-com #     backend: api.gitguardian.com |
| proxy.tunnels.web.service.ports.health.exposed | bool | `false` | Defines whether the health port is exposed if service.type is LoadBalancer or NodePort |
| proxy.tunnels.web.service.ports.health.port | int | `8081` | Specify the health port number |
| proxy.tunnels.web.service.ports.web.exposed | bool | `true` | Defines whether the web port is exposed if service.type is LoadBalancer or NodePort |
| proxy.tunnels.web.service.ports.web.port | int | `80` | Specify the web port number |
| proxy.tunnels.web.service.type | string | `"ClusterIP"` | Kubernetes Service type |
| proxy.updateStrategy.rollingUpdate.maxSurge | int | `1` |  |
| proxy.updateStrategy.rollingUpdate.maxUnavailable | int | `0` |  |
| proxy.updateStrategy.type | string | `"RollingUpdate"` | Customize updateStrategy |
| proxyProtocol.enabled | bool | `true` | When true, enables proxy protocol v2 for web/tls tunnels |
| replicaCount | int | `1` | Number of pods for each deployment |
| resources.limits | object | `{}` | Set container limits |
| resources.requests | object | `{"cpu":"100m","memory":"128Mi"}` | Set container requests |
| server.gateway.annotations | object | `{}` | Set gateway annotations |
| server.gateway.enabled | bool | `false` | Enable server exposure using Kubernetes Gateway API |
| server.gateway.gateway.className | string | `""` | Set the gatewayClassName |
| server.gateway.gateway.create | bool | `true` | Specifies whether a Gateway resource should be created alongside the routing resource (HTTPRoute) |
| server.gateway.gateway.ports | object | `{"http":80,"https":443}` | Specify Gateway ports number |
| server.gateway.parentRefs | list | `[]` | Specify the existing gateway resources |
| server.idleTimeout | int | `30` | Configure how much time a tunnel server is going to wait idle (without any new ws clients) before unbinding itself/stopping the server |
| server.ingress.annotations | object | `{}` | Set ingress annotations |
| server.ingress.className | string | `""` | Set the ingress ClassName |
| server.ingress.controller | string | `""` | Specify the ingress controller |
| server.ingress.enabled | bool | `false` | Enable exposure using Kubernetes Ingress API |
| server.istio.annotations | object | `{}` | Set Istio annotations |
| server.istio.enabled | bool | `false` | Enable server exposure using Istio ingress |
| server.istio.gateway.create | bool | `true` | Specifies whether an Istio Gateway resource should be created alongside the Virtual Service |
| server.istio.gateway.namespace | string | `""` | Specify the gateway namespace |
| server.istio.gateway.ports | object | `{"http":80,"https":443}` | Specify Istio Gateway ports number |
| server.istio.gateway.selector | object | `{"istio":"ingress"}` | Set Istio Gateway selector |
| server.istio.gateway.tls | object | `{"credentialName":"","minProtocolVersion":"TLSV1_2"}` | Specify Gateway TLS options |
| server.istio.gateway.tls.credentialName | string | `""` | Set the exising TLS secret |
| server.istio.gateways | list | `[]` | Specify the existing gateway resources for Virtual Service |
| server.service.annotations | object | `{}` | Specify server serivce annottions |
| server.service.ports.ws.containerPort | int | `9000` | Set the server websocket container port |
| server.service.ports.ws.port | int | `80` | Set the server websocket service port |
| server.service.ports.wss.containerPort | int | `9000` | Set the server websocket container port |
| server.service.ports.wss.port | int | `443` | Set the server secured websocket service port |
| server.service.type | string | `"ClusterIP"` | Kubernetes Service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the service account |
| serviceAccount.create | bool | `true` | Specifies whether a service account should be created |
| serviceAccount.name | string | `""` | The name of the service account to use |
| subdomain | string | `""` | Subdomain |
| terminationGracePeriodSeconds | int | `300` |  |
| tls.autoGenerated | bool | `false` | Generate automatically self-signed TLS certificates |
| tls.caCrt | string | `""` | CA certificate in PEM format |
| tls.certManager.enabled | bool | `false` | Manage certifcates with cert-manager |
| tls.certManager.issuer.kind | string | `"Issuer"` | Issuer kind ("Issuer" or "ClusterIssuer") |
| tls.certManager.issuer.name | string | `""` | Set Issuer to use |
| tls.certManager.issuer.spec | object | `{}` | Set Issuer spec (if specified, it will manage the issuer) |
| tls.certManager.namespace | string | `"cert-manager"` | cert-manager namespace |
| tls.certManager.rbac.create | bool | `true` |  |
| tls.certManager.rbac.subjects | list | `[]` | Grants cert-manager permissions to the sepcfied subjects (Deprecated) # e.g: # subjects: #   - kind: ServiceAccount #     name: gim #     namespace: gim |
| tls.certManager.serviceAccount | string | `"cert-manager"` | cert-manager service account name |
| tls.crt | string | `""` | Server certificate in PEM format |
| tls.enabled | bool | `false` | Enable TLS traffic support |
| tls.existingSecret | string | `""` | Specify TLS secret containing the certificate |
| tls.existingSecretKeys.caCrt | string | `""` | Existing secret key storing the Certificate Authority |
| tls.existingSecretKeys.crt | string | `""` | Existing secret key storing the server certificate |
| tls.existingSecretKeys.key | string | `""` | Existing secret key storing the server certificate key |
| tls.key | string | `""` | Server certificate key in PEM format |
| tls.mode | string | `"mutual"` | TLS mode (can be "mutual" or "simple") |
| tls.verifyCert | bool | `true` | Enable TLS certificate verification |
| tolerations | list | `[]` | Tolerations for pod assignment |
| topologySpreadConstraints | list | `[]` | Topology spread constraints for pod assignment |
| updateStrategy.rollingUpdate.maxSurge | int | `1` |  |
| updateStrategy.rollingUpdate.maxUnavailable | int | `0` |  |
| updateStrategy.type | string | `"RollingUpdate"` | Customize updateStrategy of Deployment |
| whitelist | object | `{"cidr":[],"hosts":[]}` | Specify hosts whitelist (Only available for web and tls tunnel) # e.g: # whitelist: #   hosts: #     - hook.gitguardian.com #   cidr: #     - 10.85.0.0/16 |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
