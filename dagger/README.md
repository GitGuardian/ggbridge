# ggbridge - Dagger Integration

This directory contains the [Dagger](https://dagger.io/) module for the `ggbridge` project.

Dagger is a portable devkit for CI/CD pipelines. It allows you to define pipelines as code using familiar programming languages, run them locally or in CI, and share logic across projects with reusable modules.

## ✨ What Dagger Enables

With Dagger, we can:

- Build and test Docker images / Helm chart for `ggbridge` in a consistent environment.
- Push signed images to a container registry.
- Reuse pipeline logic across development and CI/CD.
- Rapidly iterate using a local setup before pushing to CI.

## 🚀 Getting Started

### 1. Install Dagger

To use Dagger locally, you'll need the Dagger CLI and Engine installed. Follow the official installation guide:

👉 [Install Dagger](https://docs.dagger.io/installation)

### 2. Set Up the Python Dev Environment

We use the Python SDK for defining our Dagger pipelines.

You can create a virtual environment and install the required dependencies as follows:

👉 [Dagger Doc](https://docs.dagger.io/api/ide-integration)

```bash
cd dagger
dagger develop --sdk=python
uv sync
```

👉 [Dagger Python SDK Doc](https://dagger-io.readthedocs.io/en/sdk-python-v0.18.5/index.html)

### 3. Run Pipeline Functions with `dagger call`

You can also run specific pipeline functions defined in the Dagger Python module using the `dagger call` command, which provides a more structured and CLI-friendly way to invoke actions.

For example, you can run the following commands form the project root:

Show all available commands

```shell
dagger call --help
```

Build the ggbridge image

```shell
# Build the image
dagger call build

# Build the image for arm64 and amd64
dagger call build --platform linux/amd64,linux/arm64

# Build the image and export the tarball
dagger call build export --path /tmp/ggbridge.tar
```

Build the ggbridge container an open an interactive terminal, usefull to check your image after updating it

```shell
dagger call container terminal
```

Scan the ggbridge image using `grype`

```shell
# By default, we set the severity to “high”, meaning the process will fail if any vulnerability with a severity level greater than or equal to high is detected.
dagger call scan

# Set the severity
dagger call scann --severity critical
```

Test `ggbridge`: This will build the `ggbridge` image and start a ggbridge server, which a ggbridge client will then connect to, ensuring that the tunnels are properly established between the client and the server running some `curl` commands.

```shell
dagger call test
```

As a result you should have this output:

```shell
| Test                              | Result |
|-----------------------------------|--------|
| Client → Server health tunnel     | ✅     |
| Server → Client health tunnel     | ✅     |
| Server → Client proxy tunnel      | ✅     |
```

Publish the ggbridge image: this will run the following steps:

- Build the image
- Scan the image
- publish the image on the specified registry (By default, the image is published on [ttl.sh](https://ttl.sh/) registry)

```shell
dagger call publish

# Publish multi-arch image
dagger call publish \
  --platform linux/amd64,linux/arm64

# Publish the image to a pri
dagger call \
  with-registry-auth \
    --username ${DOCKERHUB_USERNAME} \
    --secret env:DOCKERHUB_TOKEN \
    --address docker.io \
  publish \
    --repository docker.io/${DOCKERHUB_USERNAME}/ggbridge \
    --version 1.0.0 \
    --paltform linux/amd64,linux/arm64

# Publish and sign the image
```

Build the Helm chart

```shell
dagger call build-chart

# Build the helm chart and export the Chart archive
dagger call build-chart export --path ggbridge.tar.gz

# Build the helm chart specifying versions
dagger call build-chart \
  --version "1.0.0" \
  --app-version "1.12.0"
```

Test the Helm chart

```shell
dagger call test-chart
```

Publish the Helm chart

```shell
dagger call publish-chart

# Specifying a custom registry
dagger call \
  with-registry-auth \
    --username ${DOCKERHUB_USERNAME} \
    --secret env:DOCKERHUB_TOKEN \
    --address docker.io \
  publish-chart \
    --registry oci://docker.io/${DOCKERHUB_USERNAME}/helm-ggbridge \
    --version "1.0.0" \
    --app-version "1.12.0"
```
