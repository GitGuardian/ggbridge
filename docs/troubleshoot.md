## Troubleshooting
### Debug image

GGBridge provides two types of images for different use cases:

- **Production image**: `ghcr.io/gitguardian/ggbridge:latest` - Minimal, secure image without shell access
- **Debug image**: `ghcr.io/gitguardian/ggbridge:latest-shell` - Includes debugging tools

**Available debug tools**: `bash`, `curl`, `net-tools`, `bind-tools`, `openssl`, `dig`, `nslookup`

**How to switch to debug image**:
- **Docker Compose**: Update the image tag in `docker-compose.yaml`
- **Helm**: Update the image tag in `values.yaml`

### Basic checks

Some very basic commands can be executed to check deployment healthiness before going further into debugging.

Check pods status:
```bash
kubectl get pods -n ggbridge -o wide
```

You should be looking for `Running` status, Ready column showing `1/1` (or `2/2`), and low restart count.

Check pods details:
```bash
kubectl describe pod $pod_name -n ggbridge 
```

Have a look at the `Events` section for suspicious warnings or errors.


### Connectivity Tests

#### 1. Client-Side Healthcheck

Verify basic connectivity from the client to the server:

```bash
kubectl exec -it $ggbridge_pod -- bash -c "curl http://127.0.0.1:9081/healthz"
```

Expected output:
```console
OK
```

#### 2. SOCKS Proxy Test (Server-Side)
Test SOCKS proxy connectivity and DNS resolution.

> [!IMPORTANT]  
> If you want to execute this test on **client side**, you need first to enable `socks` tunnel in your `values.yaml` by adding these lines :
> ```yaml
> client:
>   tunnels:
>     socks:
>       enabled: true
> ```
> Then upgrade your deployment:
> ```bash
> helm -n ggbridge upgrade -i gbridge oci://ghcr.io/gitguardian/ggbridge/helm/ggbridge -f values.yaml
> ```
> And finally, test the connection. By default, the service name will be `ggbridge-proxy` (different from the server side). Only endpoints in the allowed list can be accessed - for testing, you can use `https://api.gitguardian.com`.
> ```bash
> kubectl run debug -it --rm \
>                       --restart=Never \
>                       -n ggbridge \
>                       --image=nicolaka/netshoot:latest \
>                       -- zsh -c "curl -sILk --proxy socks5h://ggbridge-proxy.ggbridge.svc.cluster.local:1080 https://api.gitguardian.com" 

Quick test (HTTP status code only):

```bash
curl -sLk \
     -o /dev/null \
     -w "%{http_code}" \
     --connect-timeout 60 \
     --proxy "socks5h://${PROXY_HOST}:${PROXY_PORT}" "${VCS_URL}"
```

Verbose test (with headers):

```bash
curl -sILk --connect-timeout 60 \
           --proxy "socks5h://${PROXY_HOST}:${PROXY_PORT}" "${VCS_URL}"
```

Real-world example:

```bash
# Replace $uid with your actual bridge UID
kubectl run debug -it --rm \
                      --restart=Never \
                      -n ggbridge \
                      --image=nicolaka/netshoot:latest \
                      -- zsh -c "curl -sILk --proxy socks5h://$uid.ggbridge.svc.cluster.local https://vcs.example.local"
```

Expected responses:
- `200`: Success
- `301/302`: Redirect 

> [!NOTE]  
> The `socks5h` is intended for remote DNS lookup.

#### 3. Git Repository Test (Server-Side)
Test Git operations through the SOCKS proxy:

```bash
git -c http.proxy="socks5h://${PROXY_HOST}" \
    -c http.sslVerify=false \
    -c http.timeout=30 \
    ls-remote --heads "${REPO_URL_WITH_AUTH}"
```

Example with authentication:
```bash
git -c http.proxy="socks5h://$uid-proxy-socks:1080" \
    -c http.sslVerify=false \
    -c http.timeout=30 \
    ls-remote --heads "https://admin:token@gitlab.example.local/group1/myrepo.git"
```
Expected output: List of Git branches and their commit hashes

> [!TIP]
> Please consider using the `CronJob` probes available [here](../tests/) if you want a permanent check.

#### 4. Reverse tunneling

When reverse tunneling is enabled on client side, you can check if you are able to connect to `api.gitguardian.com`. Execute this command on the customer's cluster:

```bash
kubectl run debug -it --rm \
                      --restart=Never \
                      -n ggbridge \
                      --image=nicolaka/netshoot \
                      -- zsh -c "curl -IL --resolve api.gitguardian.com:443:$(kubectl get svc ggbridge-proxy-tls -n ggbridge -o jsonpath='{.spec.clusterIP}') https://api.gitguardian.com"
```

Check that DNS resolution on customer's environment properly resolve to the custom Kubernetes endpoint (implentation specific) instead of the public IP address of `hook.gitguardian.com`/`api.gitguardian.com`. Execute following commands from the customer's VCS server for example:

```bash
dig hook.gitguardian.com
dig api.gitguardian.com
```

```bash
traceroute hook.gitguardian.com
traceroute api.gitguardian.com
```

### Log Analysis

> [!TIP]
> To collect and package **client side** logs in a `.tgz` archive, you can use the [dedicated script](./client-log-fetcher.sh)

#### 1. Client/Server Healthcheck Logs
Check nginx sidecar logs for connectivity issues :

Server side:
```bash
# Check ggbridge server nginx container logs for a specific tenant and index
kubectl logs -l tenant=$uid,index=$index,app.kubernetes.io/component=server -c nginx -n ggbridge
```
Client side:
```bash
# Check ggbridge client nginx container logs for a specific index
kubectl logs -l index=$index,app.kubernetes.io/instance=ggbridge -c nginx -n ggbridge
```

Healthy connection log example for the Healthcheck probe (nginx container for server/client pod):
```console
health 127.0.0.1 [30/Sep/2025:12:04:38 +0000] 127.0.0.1 "GET /healthz HTTP/1.1" 200 3 "-" "Go-http-client/1.1"
```
No logs = No connectivity from the other tunnel endpoint.

#### 2. Client/Server tunnel Logs

Server side:
```bash
# Check ggbridge server main container logs for a specific tenant and index
kubectl logs -l tenant=$uid,index=$index,app.kubernetes.io/component=server -c ggbridge -n ggbridge
```
Client side:
```bash
# Check ggbridge client main container logs for a specific index
kubectl logs -l index=$index,app.kubernetes.io/instance=ggbridge -c ggbridge -n ggbridge
```

You should see `INFO` logs mentionning opened/closed connections:
```console
2025-10-16T08:21:06.156024Z  INFO wstunnel::protocols::tls::server: Doing TLS handshake using SNI DnsName("jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com") with the server jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com:443
2025-10-16T08:21:06.570872Z  INFO tunnel{id="0199ec1b-c14b-7f41-9492-e538c7a90f97" remote="127.0.0.1:8081"}: wstunnel::tunnel::transport::io: Closing local => remote tunnel
2025-10-16T08:21:06.571213Z  INFO tunnel{id="0199ec1b-c14b-7f41-9492-e538c7a90f97" remote="127.0.0.1:8081"}: wstunnel::tunnel::transport::io: Closing local <= remote tunnel
2025-10-16T08:21:08.489704Z  INFO tunnel{id="0199ec1b-af3c-70e3-8595-cd82aaf74cf4" remote="0.0.0.0:9081"}: wstunnel::tunnel::transport::io: Closing local => remote tunnel
2025-10-16T08:21:08.738773Z  INFO wstunnel::protocols::tls::server: Doing TLS handshake using SNI DnsName("jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com") with the server jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com:443
2025-10-16T08:21:10.693531Z  INFO tunnel{id="0199ec1b-b789-7362-b52a-5853e726c484" remote="0.0.0.0:9081"}: wstunnel::tunnel::transport::io: Closing local => remote tunnel
2025-10-16T08:21:10.947501Z  INFO wstunnel::protocols::tls::server: Doing TLS handshake using SNI DnsName("jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com") with the server jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com:443
```

> [!NOTE]
> Any log entries at `WARN` or `ERROR` level are worth highlighting if present.

> [!NOTE]
> You can also increase verbosity if needed, at `DEBUG` or `TRACE` level (default `INFO`):
> ```yaml
> logLevel: INFO # --> set de DEBUG or TRACE on server/client side values.yaml
> ```

#### 3. Server-Side Proxy Logs

Monitor traffic through the SOCKS proxy:
```bash
kubectl logs -l tenant=$uid,index=$index,app.kubernetes.io/component=proxy -n ggbridge
```

Port meanings:
- `8081`: Health checks
- `1080`: SOCKS proxy traffic
- `443`: HTTPS/TLS traffic
- `80`: HTTP traffic

Log format explanation:

| Position | Value | Nginx variable | Description | Unit |
| --- | --- | --- | --- | --- |
| 1 | `127.0.0.1` | `$remote_addr` | Local client(health check) | IP |
| 2 | `[24/Sep/2025:09:46:28 +0000]` | `[$time_local]` | Connection timestamp | Date |
| 3 | `TCP` | `$protocol` | Transport protocol | Protocol |
| 4 | `200` | `$status` | Status code | Code |
| 5 | `150` | `$bytes_sent` | Bytes sent by nginx → client | Bytes |
| 6 | `102` | `$bytes_received` | Bytes received by nginx ← client | Bytes |
| 7 | `0.077` | `$session_time` | Session duration | Seconds |
| 8 | `"172.20.167.124:8081"` | `"$upstream_addr"` | Healthcheck backend server | IP:Port |
| 9 | `"102"` | `"$upstream_bytes_sent"` | Data sent nginx → backend | Bytes |
| 10 | `"150"` | `"$upstream_bytes_received"` | Data received nginx ← backend | Bytes |
| 11 | `"0.000"` | `"$upstream_connect_time"` | Connection time | Seconds |

> [!TIP]
> If `Session duration` reach 5sec for healtcheck (port 8081), it means time out occured 

## Client Monitoring/Alerting Guidelines
### Overview

> [!NOTE]
> This guide provides generic recommendations for monitoring GGBridge client health and stability. These guidelines are platform-agnostic and can be adapted to your existing monitoring infrastructure. 

#### Replica count
Ensure that all 3 GGBridge client deployments are properly deployed, each with 1 replica:

```console
$ kubectl get deployments -n ggbridge
NAME                READY   UP-TO-DATE   AVAILABLE   AGE
ggbridge-client-0   1/1     1            1           25h
ggbridge-client-1   1/1     1            1           25h
ggbridge-client-2   1/1     1            1           25h
```

**What to monitor**: 
- All deployments should show 1/1 in the READY column

**Alert condition**: 
- Any deployment showing 0/1 or missing deployments

**Prometheus query example**:
```
kube_deployment_status_replicas_ready{namespace="ggbridge", deployment=~"ggbridge-client-.*"}
```
Count deployment with correct status (should be 3):
```
sum(
  (kube_deployment_status_replicas_ready{namespace="ggbridge", deployment=~"ggbridge-client-.*"} == 1) and
  (kube_deployment_spec_replicas{namespace="ggbridge", deployment=~"ggbridge-client-.*"} == 1) and
  (kube_deployment_status_replicas_available{namespace="ggbridge", deployment=~"ggbridge-client-.*"} == 1)
)
```

#### Pod Status and Readiness
Check that all pods are running and ready to accept connections:

```console
$ kubectl get pods -n ggbridge
NAME                                  READY   STATUS    RESTARTS   AGE
ggbridge-client-0-76687c7f6f-h6zrj   2/2     Running   0          25h
ggbridge-client-1-89abc123de-xyz45   2/2     Running   0          25h
ggbridge-client-2-12def456gh-abc78   2/2     Running   0          25h
```

**What to monitor**:
- All pods should show 2/2 in the READY column (ggbridge + nginx containers)
- STATUS should be Running
- Monitor restart count - frequent restarts indicate issues

**Alert conditions**:
- Pod showing 1/2 ready (connection issues with server)
- Pod in `CrashLoopBackOff`, `Error`, or `Pending` status
- High restart count (>5 restarts in 1 hour)

**Prometheus query example**:
```
kube_pod_status_ready{condition="true", namespace="ggbridge", pod=~"ggbridge-client-.*"} 
```

#### Container Logs Analysis
Monitor logs from the `ggbridge` container for connection issues:

**Key error patterns to watch for**:

WebSocket handshake failures (server connectivity issues):

```console
2025-09-30T15:35:11.627155Z ERROR tunnel{id="01999b43-6b64-7a61-bab6-6ff55b03aade" remote="127.0.0.1:8081"}: wstunnel::tunnel::client::client: failed to do websocket handshake with the server wss://jpynh30wscp60zs4lbdf4m4p8qe9idgu.ggbridge.gitguardian.com:443
```

**What to monitor**:
- Frequency of ERROR log entries
- Specific error patterns indicating connectivity issues
- Connection establishment success/failure rates

**Loki query example**:
```
{k8s_namespace_name="ggbridge", k8s_pod_name=~"ggbridge-client-.*"} |= "ERROR"
```

#### Resource Usage
Monitor pod resource consumption:
```console
$ kubectl top pods -n ggbridge
NAME                                 CPU(cores)   MEMORY(bytes)   
ggbridge-client-0-76687c7f6f-h6zrj   8m           7Mi             
ggbridge-client-1-bd75768f4-cr59l    10m          8Mi             
ggbridge-client-2-689f9d7c5-bz9k5    9m           7Mi
```
**What to monitor**:
- CPU usage
- Memory usage
- Sudden spikes in resource usage

**Prometheus query example**:
```
# CPU (millicores)
rate(container_cpu_usage_seconds_total{namespace="ggbridge", pod=~"ggbridge-client-.*", container!="POD", container!=""}[5m]) * 1000

# Memory (MB) 
container_memory_working_set_bytes{namespace="ggbridge", pod=~"ggbridge-client-.*", container!="POD", container!=""} / 1024 / 1024
```

### Getting Support
For technical support, please contact [support@gitguardian.com](mailto:support@gitguardian.com) with:
1. Environment details: Kubernetes version, GGBridge version
2. Error logs: Include relevant nginx and application logs
3. Configuration: Sanitized `values.yaml` or `docker-compose.yaml`
4. Test results: Output from the connectivity tests above
5. Network setup: Information about firewalls, proxies, DNS configuration