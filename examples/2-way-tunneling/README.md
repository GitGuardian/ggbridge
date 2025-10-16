# Enabling 2-Way Tunneling

By default, the `ggbridge` client does not expose any **client → server** proxies that allow accessing remote resources through the established tunnel.

However, `ggbridge` supports client-to-server proxy with **TLS Routing**.

**TLS Routing** exposes a TLS port to forward all incoming TLS traffic through a TCP tunnel. This requires **DNS redirection**, e.g., redirecting `hook.gitguardian.com` requests to the ggbridge client.

> [!CAUTION]  
> On the server side, only requests to `hook.gitguardian.com` and `api.gitguardian.com` are allowed for client → server traffic.

---

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

### 1. LoadBalancer Service

Helm values file example 👉 [values-tls-service.yaml](./helm/values-tls-service.yaml)

```yaml
proxy:
  tunnels:
    tls:
      service:
        type: LoadBalancer
        ports:
          tls:
            port: 443
```
Ensure that your VCS (and any assets using the GitGuardian hook) resolve `hook.gitguardian.com` and `api.gitguardian.com` to the external IP address of the created LoadBalancer Service (default name `ggbridge-proxy-tls`).

> [!NOTE]  
> `LoadBalancer` Service exposes the Service externally using an external load balancer. Kubernetes does not directly offer a load balancing component; you must provide one, or you can integrate your Kubernetes cluster with a cloud provider. Refer to official [Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) for more information.

### 2. Custom Resource (CRD)

Some Ingress controllers provide Custom Resources that enable enhanced capabilities—such as TCP handling, authentication, and other advanced features. This is the case with Traefik, Istio, and other well-known Kubernetes networking solutions.

Currently we support Traefik [`IngressRouteTCP`](https://doc.traefik.io/traefik/reference/routing-configuration/kubernetes/crd/tcp/ingressroutetcp/) Custom Resource. Ensure that Traefik ingress controller is installed in your cluster first.

Helm values file example 👉 [values-tls-traefik.yaml](./helm/values-tls-traefik.yaml)

```yaml
proxy:
  tunnels:
    tls:
      ingress:
        enabled: true
        controller: "traefik"
```

Ensure that your VCS (and any assets using the GitGuardian hook) resolve `hook.gitguardian.com` and `api.gitguardian.com` to the external IP address of the Traefik LoadBalancer Service.
```console
$ kubectl get service -n traefik
NAME             TYPE           CLUSTER-IP     EXTERNAL-IP     PORT(S)                      AGE
traefik          LoadBalancer   10.43.65.162   172.31.26.173   80:20572/TCP,443:27014/TCP   29h
```

### 3. Gateway API

> [!NOTE]  
> [Gateway API](https://gateway-api.sigs.k8s.io/guides/#getting-started-with-gateway-api) must be installed in your cluster first.

Helm values file example 👉 [values-tls-gateway.yaml](./helm/values-tls-gateway.yaml)

```yaml
proxy:
  tunnels:
    tls:
      gateway:
        enabled: true
        gateway:
          # This will create the gateway resource, you can disable it if you want to manage it on your own.
          create: true
          # This will set the gatewayClass of the Gateway
          className: "my-gatewayclass"
```

This will create a [`Gateway`](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gateway) and [`TLSRoute`](https://gateway-api.sigs.k8s.io/concepts/api-overview/#tlsroute) object. You need to deploy first a [`GatewayClass`](https://gateway-api.sigs.k8s.io/concepts/api-overview/#gatewayclass) in your cluster.

If your teams already manage `GatewayClass` and `Gateway` resources, you can reference the `Gateway` in the `TLSRoute` with :
```yaml
proxy:
  tunnels:
    tls:
      gateway:
        enabled: true
        parentRefs:
        - name: my-gateway
          namespace: my-gateway-namespace
        gateway:
          create: false
```

Ensure that your VCS (and any assets using the GitGuardian hook) resolve `hook.gitguardian.com` and `api.gitguardian.com` to the IP address of the `Gateway` Custom Resource.
```console
$ kubectl get gateway
NAME             CLASS             ADDRESS         PROGRAMMED    AGE
my-gateway       my-gatewayclass   192.168.1.200   True          54h
```

### 4. OpenShift Route

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