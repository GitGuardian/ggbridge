# Enabling 2-Way Tunneling

By default, the `ggbridge` client does not expose any **client → server** proxies that allow accessing remote resources through the established tunnel.

However, `ggbridge` supports two types of client-to-server proxies:

- **TLS Routing**: Exposes a TLS port to forward all incoming TLS traffic through a TCP tunnel. This requires DNS redirection, e.g., redirecting `hook.gitguardian.com` requests to the ggbridge client.
- **HTTPS Routing**: Exposes an HTTPS service and allows defining routing rules. For instance, you can expose a hostname like `hook.gitguardian.internal` and route traffic to `hook.gitguardian.com` through the ggbridge tunnel.

Both approaches require DNS configuration:
- For **TLS Routing**, redirect the hostname `hook.gitguardian.com` to the ggbridge client endpoint (for example, the LoadBalancer Service IP address).
- For **HTTPS Routing**, create a new DNS entry for the custom hostname that will be used to reach `hook.gitguardian.com`.

> [!NOTE]  
> Please enable either **TLS Routing** or **HTTPS Routing** but not both.

> [!CAUTION]  
> On the server side, only requests to `hook.gitguardian.com` and `api.gitguardian.com` are allowed for client → server traffic.

---

## HTTPS Routing

![https-routing](../../docs/images/ggbridge-https-routing.drawio.png)

To enable HTTPS Routing, update your Helm values:

```yaml
client:
  tunnels:
    web:
      enabled: true
```

> [!IMPORTANT]  
> The Ingress and Gateway approaches require generating a certificate for the custom hostname. We recommend using a certificate issued by your trusted CA rather than a self‑signed certificate. Once the certificate is ready, create a Kubernetes TLS `Secret` that stores the certificate (public key) and its corresponding private key. To generate the certificate, run:
> ```bash
>  kubectl create secret tls internal-crt --cert=tls.crt --key=tls.key -n ggbridge
> ```

This service can be exposed via Ingress or Gateway:

1. Ingress

> [!NOTE]  
> Ingress approach requires an Ingress controller installed in your cluster


Helm values file example 👉 [values-https-ingress.yaml](./helm/values-https-ingress.yaml)

```yaml
proxy:
  tunnels:
    web:
      ingress:
        enabled: true
        # -- Set the ingress ClassName (leave empty to use default)
        className: ""
        listeners:
          # This listener exposes an HTTPS service with the custom hostname
          # `hook.private.com` and routes traffic to `hook.gitguardian.com`
          - hostname: hook.private.com
            backend: hook.gitguardian.com
            tls:
              # Certificate for hook.private.com
              secretName: internal-crt
```

After deployment, verify that the DNS entry resolves to the `Ingress` IP address. All traffic destined for `hook.private.com` will be routed to your Ingress (ggbridge-proxy-web), forwarded through the tunnel, and ultimately reach `hook.gitguardian.com`.


2. Gateway API

> [!NOTE]  
> [Gateway API](https://gateway-api.sigs.k8s.io/guides/#getting-started-with-gateway-api) must be installed in your cluster

Helm values file example 👉 [values-https-gateway.yaml](./helm/values-https-gateway.yaml)

```yaml
proxy:
  tunnels:
    web:
      gateway:
        enabled: true
        # -- Set the gateway ClassName (leave empty to use default)
        className: ""
        listeners:
          - hostname: hook.private.com
            backend: hook.gitguardian.com
            tls:
              # Certificate for hook.private.com
              secretName: internal-crt
```

3. OpenShift Route

Helm values file example 👉 [values-https-openshift-route.yaml](./helm/values-https-openshift-route.yaml)

```yaml
proxy:
  resolver: dns-default.openshift-dns.svc.cluster.local
  tunnels:
    web:
      openShiftRoute::
        enabled: true
        listeners:
          - hostname: hook-gitguardian.internal.com
            backend: hook.gitguardian.com
            tls:
              termination: edge
```

> [!TIP]  
> You can verify that tunnel traffic is working by mapping `api.private.com` to `api.gitguardian.com` and issuing a request to the mapped hostname from within your environment :
> ```console
> $ curl -IsLk https://api.private.com
> HTTP/2 301 
> date: Wed, 15 Oct 2025 09:36:16 GMT
> location: https://api.gitguardian.com/docs
> server: nginx/1.29.2
> 
> HTTP/2 200 
> server: istio-envoy
> date: Wed, 15 Oct 2025 09:36:16 GMT
> content-type: text/html
> content-length: 3684076
> last-modified: Tue, 14 Oct 2025 11:44:21 GMT
> vary: Accept-Encoding
> etag: "68ee3795-3836ec"
> accept-ranges: bytes
> x-envoy-upstream-service-time: 1
> ```

## TLS Routing

![tls-routing](../../docs/images/ggbridge-tls-routing.drawio.png)

To enable **TLS Routing**, update your Helm values as follows:

```yaml
client:
  tunnels:
    tls:
      enabled: true
```

> [!NOTE]  
> **TLS Routing** does not work natively with standard `Ingress` objects. Some Ingress controllers offer specific configurations (annotations, custom parameters, etc.) to enable TLS Routing with `Ingress`, but this depends entirely on your controller’s capabilities and is highly implementation-specific. Please refer to your Ingress controller’s documentation for this particular use case.

The TLS port can be exposed using one of the following methods:

1. LoadBalancer Service

Helm values file example 👉 [values-tls-service.yaml](./helm/values-tls-service.yaml)

```yaml
proxy:
  tunnels:
    tls:
      service:
        type: LoadBalancer
        ports:
          tls:
            # 443 is the default TLS port
            port: 443
```

> [!NOTE]  
> `LoadBalancer` Service exposes the Service externally using an external load balancer. Kubernetes does not directly offer a load balancing component; you must provide one, or you can integrate your Kubernetes cluster with a cloud provider. Refer to official [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) for more information.

2. Gateway API

> [!NOTE]  
> [Gateway API](https://gateway-api.sigs.k8s.io/guides/#getting-started-with-gateway-api) must be installed in your cluster

Helm values file example 👉 [values-tls-gateway.yaml](./helm/values-tls-gateway.yaml)

```yaml
proxy:
  tunnels:
    tls:
      gateway:
        enabled: true
        # This will create the gateway resource, you can disable it if you want to mange it on you own.
        create: true
        className: ""
```

3. OpenShift Route

Helm values file example 👉 [values-tls-openshift-route.yaml](./helm/values-tls-openshift-route.yaml)

```yaml
proxy:
  resolver: dns-default.openshift-dns.svc.cluster.local
  tunnels:
    tls:
      openShiftRoute:
        enabled: true
```

With these configurations, your ggbridge client can securely forward traffic through the tunnel from the client side to approved GitGuardian services.