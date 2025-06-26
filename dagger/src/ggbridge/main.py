import random
from typing import Annotated, Self

import dagger
from dagger import DefaultPath, Doc, Name, dag, function, object_type

from .chart import Chart
from .repository import Repository
from .image import Image

UUID: str = random.randrange(10**8)


@object_type
class Ggbridge:
    """ggbridge module"""

    source: dagger.Directory
    container_: dagger.Container

    apko_: dagger.Apko
    cosign_: dagger.Cosign
    crane_: dagger.Crane
    grype_: dagger.Grype
    helm_: dagger.Helm
    melange_: dagger.Melange

    @classmethod
    async def create(
        cls,
        source: Annotated[dagger.Directory, DefaultPath("/"), Doc("Source directory")],
    ):
        """Constructor"""
        return cls(
            source=source.filter(
                include=[
                    "apko/",
                    "docker/",
                    "helm/",
                    "melange/",
                    "tests/",
                    "go.mod",
                    "go.sum",
                    "main.go",
                ]
            ),
            container_=dag.container(),
            apko_=dag.apko(source=source.filter(include=["apko/"])),
            crane_=dag.crane(),
            cosign_=dag.cosign(),
            grype_=dag.grype(),
            helm_=dag.helm(),
            melange_=dag.melange(),
        )

    # =============================================================================
    # Private Functions
    # =============================================================================

    # =============================================================================
    # Public Functions
    # =============================================================================

    @function
    def keygen(
        self,
    ) -> dagger.Directory:
        """Returns melange/cosign keypairs as directory"""
        return (
            dag.directory()
            .with_directory("cosign", self.cosign_.generate_key_pair())
            .with_directory("melange", self.melange_.keygen())
        )

    @function
    def with_registry_auth(
        self,
        username: Annotated[str, Doc("Registry username")],
        secret: Annotated[dagger.Secret, Doc("Registry password")],
        address: Annotated[str, Doc("Registry host")] = "ghcr.io",
    ) -> Self:
        """Authenticates with registry"""
        self.container_ = self.container_.with_registry_auth(
            address=address, username=username, secret=secret
        )
        self.apko_ = self.apko_.with_registry_auth(
            address=address, username=username, secret=secret
        )
        self.crane_ = self.crane_.with_docker_config(self.apko_.docker_config())
        self.cosign_ = self.cosign_.with_docker_config(self.apko_.docker_config())
        self.grype_ = self.grype_.with_docker_config(self.apko_.docker_config())
        self.helm_ = self.helm_.with_registry_auth(
            address=address, username=username, secret=secret
        )
        return self

    @function
    def with_melange_signing_key(
        self,
        key: Annotated[dagger.Secret, Doc("Key to use for signing")],
        name: Annotated[str | None, Doc("Key name")] = "melange.rsa",
    ) -> Self:
        """Set the Melange signing key"""
        self.melange_ = self.melange_.with_signing_key(
            key,
            name=name,
        )
        return self

    @function
    def with_cosign_private_key(
        self,
        key: Annotated[
            dagger.Secret | None, Doc("Private key to use for image signing")
        ] = None,
        password: Annotated[
            dagger.Secret | None, Doc("Password used to decrypt the Cosign Private key")
        ] = None,
    ) -> Self:
        """Set the private key to use for image signing with Cosign"""
        self.cosign_ = self.cosign_.with_private_key(
            key=key,
            password=password,
        )
        return self

    @function
    def with_cosign_oidc(
        self,
        provider: Annotated[
            str | None, Doc("Specify the provider to get the OIDC token from")
        ] = "",
        issuer: Annotated[
            str | None, Doc("OIDC provider to be used to issue ID token")
        ] = "",
    ) -> Self:
        """Set the OIDC parameters to use for image signing with Cosign"""
        self.cosign_ = self.cosign_.with_oidc(
            provider=provider,
            issuer=issuer,
        )
        return self

    @function
    def with_cosign_annotations(
        self,
        annotations: Annotated[list[str], Doc("Extra key=value pairs to sign")],
    ) -> Self:
        """Set the OIDC parameters to use for image signing with Cosign"""
        self.cosign_ = self.cosign_.with_annotations(annotations)
        return self

    @function
    def chart(self) -> Chart:
        """Chart functions"""
        return Chart(
            source=self.source.directory("helm/ggbridge"), helm=self.helm_, uuid=UUID
        )

    @function
    def repository(self) -> Repository:
        """Repository functions"""
        return Repository(
            source=self.source.filter(
                include=["docker/", "melange/", "go.mod", "go.sum", "main.go"]
            ),
            melange=self.melange_,
        )

    @function
    def image(self) -> Image:
        """Image functions"""
        return Image(
            source=self.source.filter(include=["apko/"]),
            repository=self.repository(),
            container_=self.container_,
            apko=self.apko_,
            cosign=self.cosign_,
            crane=self.crane_,
            grype=self.grype_,
            uuid=UUID,
        )

    @function
    async def publish(
        self,
        version: Annotated[str, Doc("Version")] = "0.1.0",
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = (dagger.Platform("linux/amd64"), dagger.Platform("linux/arm64")),
        scan: Annotated[bool, Doc("Scan the image for vulnerabilities")] = True,
        sign: Annotated[bool, Doc("Sign and Attest the image with cosign")] = False,
    ) -> list[str]:
        """Publish images and Helm chart"""
        artifacts: list[str] = []
        for variant in ("prod", "shell"):
            artifacts.append(
                await self.image().publish(
                    variant=variant,
                    version=version,
                    platforms=platforms,
                    scan=scan,
                    sign=sign,
                )
            )
        artifacts.append(await self.chart().push(version=version, app_version=version))
        return artifacts

    @function
    async def test(
        self,
    ) -> dagger.File:
        """Test ggbridge and return the test report"""
        ggbridge_server: dagger.Container = await self.server()
        ggbridge_server_svc: dagger.Service = ggbridge_server.as_service(
            args=["server"], use_entrypoint=True
        ).with_hostname("ggbridge-server")

        ggbridge_client: dagger.Container = await self.client(
            server="ggbridge-server:9000"
        )
        ggbridge_client_svc: dagger.Service = ggbridge_client.as_service(
            args=["client"], use_entrypoint=True
        ).with_hostname("ggbridge-client")

        container: dagger.Container = (
            dag.container()
            .from_("cgr.dev/chainguard/bash:latest")
            .with_service_binding("ggbridge-server", ggbridge_server_svc)
            .with_service_binding("ggbridge-client", ggbridge_client_svc)
            .with_file(
                "/tests/test.sh",
                self.source.file("tests/test.sh"),
                permissions=0o555,
            )
            .with_env_variable("GGBRIDGE_CLIENT_HOST", "ggbridge-client")
            .with_env_variable("GGBRIDGE_SERVER_HOST", "ggbridge-server")
            .with_env_variable("GGBRIDGE_TUNNEL_HEALTH_PORT", "9081")
            .with_env_variable("GGBRIDGE_TUNNEL_SOCKS_PORT", "9180")
            .with_exec(
                [
                    "sh",
                    "-c",
                    "sleep 5 && /tests/test.sh",
                ],
                redirect_stdout="/tmp/stdout",
            )
        )
        return container.file("/tmp/stdout")

    @function
    async def client(
        self,
        server: Annotated[str, Doc("Server address")],
        ca: Annotated[dagger.File | None, Doc("Certificate authority")] = None,
        cert: Annotated[dagger.File | None, Doc("Client certificate")] = None,
        key: Annotated[dagger.Secret | None, Doc("Client certificate key")] = None,
        tunnel_health_port: Annotated[int, Doc("Health tunnel port")] = 9081,
    ) -> dagger.Container:
        """Return the ggbridge client container"""
        container: dagger.Container = await self.container(variant="shell")
        tls_enabled: bool = bool(ca and cert and key)
        container = (
            container.with_env_variable("SERVER_ADDRESS", server)
            .with_env_variable("TLS_ENABLED", str(tls_enabled).lower())
            .with_env_variable("LOG_LEVEL", "debug")
            .with_env_variable("TUNNEL_HEALTH_PORT", str(tunnel_health_port))
            .with_exposed_port(
                tunnel_health_port,
                protocol=dagger.NetworkProtocol("TCP"),
                description="Health Tunnel",
            )
        )
        if tls_enabled:
            container = (
                container.with_mounted_file(
                    "/etc/ggbridge/tls/ca.crt", source=ca, owner="nonroot"
                )
                .with_mounted_file(
                    "/etc/ggbridge/tls/client.crt", source=cert, owner="nonroot"
                )
                .with_mounted_secret(
                    "/etc/ggbridge/tls/client.key", source=key, owner="nonroot"
                )
            )
        return container

    @function
    async def server(
        self,
        ca: Annotated[dagger.File | None, Doc("Certificate authority")] = None,
        cert: Annotated[dagger.File | None, Doc("Client certificate")] = None,
        key: Annotated[dagger.Secret | None, Doc("Client certificate key")] = None,
        port: Annotated[int, Doc("Server port")] = 9000,
        tunnel_health_port: Annotated[int, Doc("Health port")] = 9081,
        tunnel_socks_port: Annotated[int, Doc("Socks port")] = 9180,
    ) -> dagger.Container:
        """Return the ggbridge server container"""
        container: dagger.Container = await self.container(variant="shell")
        tls_enabled: bool = bool(ca and cert and key)
        container = (
            container.with_env_variable("SERVER_PORT", str(port))
            .with_env_variable("TLS_ENABLED", str(tls_enabled).lower())
            .with_env_variable("LOG_LEVEL", "debug")
            .with_env_variable("TUNNEL_HEALTH_PORT", str(tunnel_health_port))
            .with_env_variable("TUNNEL_SOCKS_PORT", str(tunnel_socks_port))
            .with_exposed_port(
                port,
                protocol=dagger.NetworkProtocol("TCP"),
                description="WebSocket endpoint",
            )
            .with_exposed_port(
                tunnel_health_port,
                protocol=dagger.NetworkProtocol("TCP"),
                description="Health Tunnel",
                experimental_skip_healthcheck=True,
            )
            .with_exposed_port(
                tunnel_socks_port,
                protocol=dagger.NetworkProtocol("TCP"),
                description="Socks Tunnel",
                experimental_skip_healthcheck=True,
            )
        )
        if tls_enabled:
            container = (
                container.with_mounted_file(
                    "/etc/ggbridge/tls/ca.crt", source=ca, owner="nonroot"
                )
                .with_mounted_file(
                    "/etc/ggbridge/tls/server.crt", source=cert, owner="nonroot"
                )
                .with_mounted_secret(
                    "/etc/ggbridge/tls/server.key", source=key, owner="nonroot"
                )
            )
        return container

    @function
    async def container(
        self,
        variant: Annotated[str, Doc("Variant to build")] = "shell",
    ) -> dagger.Container:
        """Return the ggbridge container"""
        container: dagger.Container = await self.image().container(variant=variant)
        user: str = await container.user()
        if variant == "shell":
            container = (
                container.with_user("0")
                .with_exec(
                    [
                        "sh",
                        "-c",
                        (
                            "sed -i"
                            " 's|resolver kube-dns.kube-system.svc.cluster.local|resolver 127.0.0.11|'"
                            " /etc/ggbridge/nginx.conf"
                        ),
                    ]
                )
                .with_user(user)
            )
        return container

    @function
    async def scan(
        self,
        variant: Annotated[str, Doc("Variant to scan")] = "prod",
        severity: Annotated[
            str, Doc("Specify the minimum vulnerability severity to trigger an error")
        ] = "",
        output_format: Annotated[str, Doc("Report output formatter")] = "table",
    ) -> dagger.File:
        """Scan the ggbridge image using grype"""
        return await self.image().scan(
            variant=variant, severity=severity, output_format=output_format
        )

    @function
    async def build(
        self,
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = (dagger.Platform("linux/amd64"), dagger.Platform("linux/arm64")),
    ) -> dagger.Directory:
        """Build images and Helm chart"""
        prod_image: dagger.Directory = await self.image().build(
            variant="prod", platforms=platforms
        )
        shell_image: dagger.Directory = await self.image().build(
            variant="shell", platforms=platforms
        )
        chart: dagger.File = await self.chart().build()
        return (
            dag.directory()
            .with_directory("images/prod", prod_image)
            .with_directory("images/shell", shell_image)
            .with_file(f"chart/{await chart.name()}", chart)
        )
