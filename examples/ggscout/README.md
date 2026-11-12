# Connect ggscout with the GitGuardian Platform

[![Watch the video](https://img.youtube.com/vi/_q0A-LZM2FI/hqdefault.jpg)](https://youtu.be/_q0A-LZM2FI)

You want to deploy **[ggscout](https://docs.gitguardian.com/ggscout-docs/home)** in your Kubernetes cluster, but it doesn’t have access to the GitGuardian platform, preventing you from viewing the NHI inventory from the dashboard.

`ggbridge` can help you establish a secure and mTLS-authenticated gateway between your network and the GitGuardian platform, allowing ggscout to communicate with the GitGuardian public API, and thus, transmit the inventory.

The ggbridge client exposes an internal Service in your cluster. ggscout sends its API requests to that Service over HTTP, and the client forwards them through the bridge's tunnel to the GitGuardian API over HTTPS.

## Requirements

- A bridge created in the GitGuardian dashboard (**Settings > Security > GitGuardian Bridge**), with its certificate bundle and bridge URL. If your workspace has no access to GitGuardian Bridge yet, contact GitGuardian support to enable it.
- ggscout running in the **same Kubernetes cluster** as the ggbridge client.

> [!NOTE]
> ggscout running outside that cluster, for example on another network, can reach the GitGuardian API through the client's TLS routing instead: see [2-way tunneling](../2-way-tunneling).

## Configuration

### ggbridge Client configuration

To enable the web tunnel and expose the GitGuardian API inside your cluster, you need to configure the ggbridge client.

Follow these steps to create a `values.yaml` file with the appropriate Helm values.

#### 1. Set the ggbridge Hostname

Configure the ggbridge client to connect to your hostname:

```yaml
# The bridge URL shown in the dashboard:
#   US: <my-subdomain>.ggbridge.gitguardian.com
#   EU: <my-subdomain>.ggbridge.eu1.gitguardian.com
hostname: <my-subdomain>.ggbridge.gitguardian.com

tls:
  enabled: true
  existingSecret: ggbridge-client-crt
  existingSecretKeys:
    caCrt: ca.crt
    crt: tls.crt
    key: tls.key
```

- `hostname`: The bridge URL of your ggbridge instance, as shown in the dashboard.
- `tls.enabled`: Enables TLS for secure communication.
- `tls.existingSecret`: References the Kubernetes secret storing your client certificates.
- `tls.existingSecretKeys`: Specifies the keys within the secret.

#### 2. Enable the Web Tunnel

The web tunnel allows HTTP traffic from the client to the server:

```yaml
client:
  tunnels:
    web:
      enabled: true
```

#### 3. Expose the GitGuardian API in Your Cluster

Declare a Service listener whose backend is the GitGuardian API of your region. The bridge server only lets that host through: `api.gitguardian.com` for US bridges, `api.eu1.gitguardian.com` for EU bridges.

```yaml
proxy:
  tunnels:
    web:
      service:
        listeners:
          - name: <my-service-name>
            # EU bridges: api.eu1.gitguardian.com
            backend: api.gitguardian.com
```

The GitGuardian API will be accessible internally at:

```shell
http://<my-service-name>.<namespace>.svc.cluster.local
```

Use this full name: the client picks the backend from the request's `Host` header, and only this name is mapped to the GitGuardian API.

Helm values file example 👉 [values-service.yaml](./helm/values-service.yaml)

#### 4. Deploy The ggbridge Client

Install/Upgrade the GGbridge client using Helm:

```shell
helm -n ggbridge upgrade --install --create-namespace \
  ggbridge oci://ghcr.io/gitguardian/ggbridge/helm/ggbridge \
  -f values.yaml
```

---

### ggscout Configuration

#### 1. Configure the Helm values

To enable **ggscout** to communicate with the GitGuardian API via the configured ggbridge client, update the `values.yaml` file with the internal API endpoint:

```yaml
inventory:
  config:
    gitguardian:
      endpoint: "http://<my-service-name>.<namespace>.svc.cluster.local"
      api_token: <my-api-token>
```

- `endpoint`: The internal URL of the GitGuardian API, from step 3 above.
- `api_token`: The authentication token for API access.

Refer to the [ggscout GitHub project](https://github.com/GitGuardian/nhi-scout-helm-charts) for more documentation on how to configure the Helm chart.

#### 2. Deploy ggscout

Add the repository to Helm with:

```shell
helm repo add gg-nhi https://gitguardian.github.io/nhi-scout-helm-charts
```

Install/Upgrade the Scout using Helm:

```shell
helm -n scout upgrade --install --create-namespace \
  scout gg-nhi/nhi-scout \
  -f values.yaml
```
