# K3d cluster deployment guide

This guide will help you deploy and manage local Kubernetes clusters using k3d for GGBridge deployment.

> [!WARNING]  
> This setup is not recommended for production due to single point of failure and lack of high availability.
>
> **For production setup**, we recommend deploying on a multi-node Kubernetes cluster across multiple AZs with proper redundancy and monitoring.

## 🔧 Prerequisites
For this installation method, you will need a single server (VM, bare metal...) and install following components:

- [Docker](https://docs.docker.com/get-docker/) installed and running
- [K3d](https://k3d.io/v5.8.3/#installation) version 5.8.3 or superior
- [helm](https://helm.sh/docs/intro/install/) to install GGBridge in the k3d cluster
- [kubectl](https://kubernetes.io/docs/tasks/tools/) to interact with the cluster

## ⚡ Quick Start

> [!IMPORTANT]  
> There are two methods for installing GGBridge on k3d, the [Basic CLI installation](#-basic-cli-installation), and [using config files](#-config-file-installation). While the basic CLI works, we recommend the second one for a long-term and maintainable lifecycle.


### 💻 Basic CLI installation

> [!NOTE]  
> Single node cluster here is minimal installation and low footprint


1. Basic cluster creation (single node cluster):
```bash
k3d cluster create ggbridge --agents 0 \
                            --servers 1 \
                            --k3s-arg "--disable=traefik@server:*" \
                            --k3s-arg "--disable=metrics-server@server:*" \
                            --k3s-arg "--disable=local-storage@server:*" \
                            --k3s-arg "--disable=servicelb@server:*" \
                            --k3s-node-label "project=ggbridge@server:*" \
                            --api-port 0.0.0.0:6445 \
                            --image rancher/k3s:v1.33.5-k3s1 \
                            --timeout 3m0s
```
2. Create the GGBridge namespace
```bash
kubectl create ns ggbridge
```

3. Create the client certificate secret

Extract the certificate bundle downloaded from the GitGuardian dashboard and create a Kubernetes secret with the certificate files
```bash
kubectl create secret generic ggbridge-client-crt -n ggbridge --from-file=tls.crt \
                                                              --from-file=tls.key \
                                                              --from-file=ca.crt
```

4. Install GGBridge client

Replace `$uid` here with the Bridge UID

```bash
helm -n ggbridge upgrade -i ggbridge oci://ghcr.io/gitguardian/ggbridge/helm/ggbridge \
                         --set hostname="$uid.ggbridge.gitguardian.com" \
                         --set tls.enabled=true \
                         --set tls.existingSecret="ggbridge-client-crt" \
                         --set tls.existingSecretKeys.caCrt="ca.crt" \
                         --set tls.existingSecretKeys.crt="tls.crt" \
                         --set tls.existingSecretKeys.key="tls.key" \
                         --set image.tag="latest"
```

5. Check installation is healthy

After few seconds, your client bridge should be `Running` and `2/2 Ready`.

> [!NOTE]  
> By default, 3 pods are deployed to ensure proper bridge functionality. This is the minimum required number and should not be reduced.

```console
$ kubectl get pods -n ggbridge
NAME                                 READY   STATUS    RESTARTS   AGE
ggbridge-client-0-58f49d45c8-rvjzt   2/2     Running   0          22s
ggbridge-client-1-75f69cdb75-5gpsv   2/2     Running   0          22s
ggbridge-client-2-76b98c699b-bk2q5   2/2     Running   0          22s
```

### 📝 Config file installation

> [!NOTE]  
> This installation is using declarative approach with explicit yaml files to describe cluster and Helm installation. It brings several advantages upon [basic CLI installation](#-basic-cli-installation). We recommend using this method if you are familiar with Kubernetes.
> 
> | Advantage             | Config File installation           | Basic CLI installation             |
> |-----------------------|------------------------------------|------------------------------------|
> | **Version Control**   | ✅ Easy to track changes           | ❌ Hard to version long commands   |
> | **Reproducibility**   | ✅ Identical deployments           | ❌ Prone to human error            |
> | **Documentation**     | ✅ Self-documenting                | ❌ Requires separate docs          |
> | **Collaboration**     | ✅ Easy to share & review          | ❌ Command sharing is cumbersome   |
> | **Complex Configs**   | ✅ Handles complexity well         | ❌ Commands become unwieldy        |
> | **Schema Validation** | ✅ IDE autocompletion & validation | ❌ No validation until execution   |
> | **Reusability**       | ✅ Template-friendly               | ❌ Hard to parameterize            |
> | **Maintenance**       | ✅ Easy updates & modifications    | ❌ Requires command reconstruction |
> 

1. Create the cluster with [dedicated configuration file](../k3d/cluster.yaml)

```bash
k3d cluster create --config cluster.yaml
```

2. Create the client certificate secret

Extract the certificate bundle downloaded from the GitGuardian dashboard and create a Kubernetes secret with the certificate files
```bash
kubectl create secret generic ggbridge-client-crt -n ggbridge --from-file=tls.crt \
                                                              --from-file=tls.key \
                                                              --from-file=ca.crt
```

3. Install GGBridge client

Edit the helm [install file](../k3d/helm-ggbridge.yaml) with your Bridge UID
```yaml
  valuesContent: |-
    hostname: $uid.ggbridge.gitguardian.com
```
Install GGBridge
```bash
kubectl apply -f helm-ggbridge.yaml
```

4. Check installation is healthy

After few seconds, your client bridge should be `Running` and `2/2 Ready`.

> [!NOTE]  
> By default, 3 pods are deployed to ensure proper bridge functionality. This is the minimum required number and should not be reduced.

```console
$ kubectl get pods -n ggbridge
NAME                                 READY   STATUS    RESTARTS   AGE
ggbridge-client-0-58f49d45c8-rvjzt   2/2     Running   0          22s
ggbridge-client-1-75f69cdb75-5gpsv   2/2     Running   0          22s
ggbridge-client-2-76b98c699b-bk2q5   2/2     Running   0          22s
```