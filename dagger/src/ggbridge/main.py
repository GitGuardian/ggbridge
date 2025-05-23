import random
from typing import Annotated, Self

import dagger
from dagger import DefaultPath, Doc, Name, dag, function, object_type

UUID: str = random.randrange(10**8)


@object_type
class Ggbridge:
    """ggbridge module"""

    source: dagger.Directory

    container_: dagger.Container | None
    cosign: dagger.Cosign | None
    crane: dagger.Crane | None
    helm: dagger.Helm | None

    @classmethod
    async def create(
        cls,
        source: Annotated[dagger.Directory, DefaultPath("/"), Doc("Source directory")],
    ):
        """Constructor"""
        return cls(
            source=source,
            container_=dag.container(),
            cosign=dag.cosign(),
            crane=dag.crane(),
            helm=dag.helm(),
        )

    # =============================================================================
    # Private Functions
    # =============================================================================

    # =============================================================================
    # Public Functions
    # =============================================================================

    @function
    def melange_keygen(
        self,
    ) -> dagger.Directory:
        """Generate a signing key for the APK repository"""
        return dag.melange().keygen()

    @function
    def cosign_keygen(
        self, password: Annotated[dagger.Secret | None, Doc("Key password")] = None
    ) -> dagger.Directory:
        """Generate a cosign key pair for image signing"""
        return self.cosign.generate_key_pair(password=password)

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
        self.crane = self.crane.with_registry_auth(
            address=address, username=username, secret=secret
        )
        self.cosign = self.cosign.with_registry_auth(
            address=address, username=username, secret=secret
        )
        self.helm = self.helm.with_registry_auth(
            address=address, username=username, secret=secret
        )
        return self

    @function
    async def publish(
        self,
        repositories: Annotated[
            list[str], Doc("Image repository"), Name("repository")
        ] = (),
        version: Annotated[str, Doc("Version")] = "0.1.0",
        latest: Annotated[bool, Doc("Tag as latest")] = False,
        variant: Annotated[str, Doc("Image variant")] = "prod",
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = None,
        sign: Annotated[bool, Doc("Sign image")] = False,
        cosign_private_key: Annotated[
            dagger.Secret | None, Doc("Private key for image signing")
        ] = None,
        cosign_password: Annotated[
            dagger.Secret | None, Doc("Password for image signing key")
        ] = None,
        cosign_oidc_provider: Annotated[
            str, Doc("OIDC provider for image signing")
        ] = "",
        cosign_oidc_issuer: Annotated[str, Doc("OIDC issuer for image signing")] = "",
        melange_repository: Annotated[
            dagger.Directory | None, Doc("APK repository")
        ] = None,
        melange_signing_key: Annotated[
            dagger.File | None, Doc("Melange Signing key")
        ] = None,
        severity: Annotated[
            str, Doc("Specify the minimum vulnerability severity to trigger an error")
        ] = "high",
    ) -> str:
        """Publish the ggbridge image"""
        platforms = platforms or [await dag.default_platform()]

        # Build the image
        build: dagger.Directory = await self.build(
            variant=variant,
            platforms=platforms,
            repository=melange_repository,
            signing_key=melange_signing_key,
        )
        sbom: dagger.Directory = build.directory("sbom")
        tarball: dagger.File = build.file("image.tar")
        tarball_digest: str = await tarball.digest()

        # Scan the image for vulnerabilities
        dag.grype().scan_file(
            source=tarball,
            source_type="oci-archive",
            severity_cutoff=severity,
            fail=True,
            output_format="table",
        )

        # Publish image
        platform_variants: list[dagger.Container] = []
        for platform in platforms:
            platform_variants.append(dag.container(platform=platform).import_(tarball))

        address: str = (
            f"{repositories[0]}:{version}"
            if repositories
            else f"ttl.sh/ggbridge-{tarball_digest.split(':')[1][:8]}:{version}"
        )
        digest: str = await self.container_.publish(
            address=address, platform_variants=platform_variants
        )

        if sign:
            # Sign image
            await self.cosign.sign(
                image=digest,
                private_key=cosign_private_key,
                password=cosign_password,
                oidc_provider=cosign_oidc_provider,
                recursive=True,
            )
            if len(platforms) > 1:
                # Attest index SBOM
                await self.cosign.attest(
                    image=digest,
                    private_key=cosign_private_key,
                    password=cosign_password,
                    oidc_provider=cosign_oidc_provider,
                    oidc_issuer=cosign_oidc_issuer,
                    type_="spdxjson",
                    predicate=sbom.file("sbom-index.spdx.json"),
                )
            # Attest platforms SBOMs
            sbom_file: dagger.File = None
            for platform in platforms:
                if platform == dagger.Platform("linux/amd64"):
                    sbom_file = sbom.file("sbom-x86_64.spdx.json")
                elif platform == dagger.Platform("linux/arm64"):
                    sbom_file = sbom.file("sbom-aarch64.spdx.json")
                platform_digest: str = digest
                if len(platforms) > 1:
                    platform_digest = await self.crane.digest(
                        image=address, platform=platform, full_ref=True
                    )
                await self.cosign.attest(
                    image=platform_digest,
                    private_key=cosign_private_key,
                    password=cosign_password,
                    oidc_provider=cosign_oidc_provider,
                    oidc_issuer=cosign_oidc_issuer,
                    type_="spdxjson",
                    predicate=sbom_file,
                )

        if latest:
            # Add latest tag
            await self.crane.tag(image=address, tag="latest")

        for repository in repositories[1:]:
            # Copy image (including signatures, attestations and SBOMs)
            await self.container_.publish(
                address=f"{repository}:{version}", platform_variants=platform_variants
            )
            await self.cosign.copy(
                source=address,
                destination=f"{repository}:{version}",
                force=True,
            )
            if latest:
                await self.crane.tag(image=f"{repository}:{version}", tag="latest")

        return digest

    @function
    async def publish_chart(
        self,
        registry: Annotated[str, Doc("Helm registry")] = "",
        version: Annotated[str, Doc("Helm chart version")] = "0.1.0",
        app_version: Annotated[str, Doc("App version")] = "unstable",
    ) -> str:
        """Publish the Helm chart"""
        chart: dagger.File = await self.build_chart(
            version=version, app_version=app_version
        )
        chart_digest: str = await chart.digest()
        if not registry:
            registry = f"oci://ttl.sh/ggbridge-helm-{chart_digest.split(':')[1][:8]}"
        return await self.helm.push(chart=chart, registry=registry)

    @function
    async def scan(
        self,
        variant: Annotated[str, Doc("Variant to scan")] = "prod",
        severity: Annotated[
            str, Doc("Specify the minimum vulnerability severity to trigger an error")
        ] = "high",
        output_format: Annotated[str, Doc("Report output formatter")] = "table",
    ) -> str:
        """Scan the ggbridge image using grype"""
        build: dagger.Directory = await self.build(variant=variant)
        tarball: dagger.File = build.file("image.tar")
        report: dagger.File = dag.grype().scan_file(
            source=tarball,
            source_type="oci-archive",
            severity_cutoff=severity,
            fail=False,
            output_format=output_format,
        )
        return await report.contents()

    @function
    async def test(
        self,
    ) -> str:
        """Test ggbridge"""
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
            .from_("ghcr.io/gitguardian/wolfi/bash:latest")
            .with_service_binding("ggbridge-server", ggbridge_server_svc)
            .with_service_binding("ggbridge-client", ggbridge_client_svc)
            .with_file(
                "/tests/test.sh",
                self.source.file("dagger/tests/test.sh"),
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
                ]
            )
        )
        stdout: str = await container.stdout()
        return stdout

    @function
    async def test_chart(
        self,
    ) -> str:
        """Test Helm chart"""
        return await self.helm.lint(self.source.directory("helm/ggbridge"), strict=True)

    @function
    async def client(
        self,
        server: Annotated[str, Doc("Server address")],
        ca: Annotated[dagger.File | None, Doc("Certificate authority")] = None,
        cert: Annotated[dagger.File | None, Doc("Client certificate")] = None,
        key: Annotated[dagger.File | None, Doc("Client certificate key")] = None,
        tunnel_health_port: Annotated[int, Doc("Health tunnel port")] = 9081,
    ) -> dagger.Container:
        """Return a ggbridge client container"""
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
                container.with_mounted_file("/etc/ggbridge/tls/ca.crt", source=ca)
                .with_mounted_file("/etc/ggbridge/tls/client.crt", source=cert)
                .with_mounted_file("/etc/ggbridge/tls/client.key", source=key)
            )
        return container

    @function
    async def server(
        self,
        ca: Annotated[dagger.File | None, Doc("Certificate authority")] = None,
        cert: Annotated[dagger.File | None, Doc("Client certificate")] = None,
        key: Annotated[dagger.File | None, Doc("Client certificate key")] = None,
        port: Annotated[int, Doc("Server port")] = 9000,
        tunnel_health_port: Annotated[int, Doc("Health port")] = 9081,
        tunnel_socks_port: Annotated[int, Doc("Socks port")] = 9180,
    ) -> dagger.Container:
        """Return a ggbridge server container"""
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
                container.with_mounted_file("/etc/ggbridge/tls/ca.crt", source=ca)
                .with_mounted_file("/etc/ggbridge/tls/server.crt", source=cert)
                .with_mounted_file("/etc/ggbridge/tls/server.key", source=key)
            )
        return container

    @function
    async def container(
        self,
        variant: Annotated[str, Doc("Variant to build")] = "shell",
    ) -> dagger.Container:
        """Return a ggbridge container"""
        build: dagger.Directory = await self.build(variant=variant)
        tarball: dagger.File = build.file("image.tar")
        container: dagger.Container = dag.container().import_(tarball)
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
    async def build(
        self,
        tag: Annotated[str, Doc("Image tag")] = "ggbridge",
        variant: Annotated[str, Doc("Variant to build")] = "prod",
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = None,
        repository: Annotated[dagger.Directory | None, Doc("APK repository")] = None,
        signing_key: Annotated[dagger.Secret | None, Doc("Signing key")] = None,
    ) -> dagger.Directory:
        """Build the ggbridge image using apko"""
        if not signing_key:
            keys: dagger.Directory = await self.melange_keygen()
            signing_key_file: dagger.File = keys.file("melange.rsa")
            signing_key = dag.set_secret(
                name="melange_signing_key", plaintext=await signing_key_file.contents()
            )

        if not repository:
            repository: dagger.Directory = await self.build_repository(
                signing_key=signing_key,
                platforms=platforms,
            )

        apko: dagger.Apko = dag.apko()
        apko_build: dagger.ApkoBuild = apko.build(
            source=self.source.filter(include=["apko/"]),
            config=self.source.file(f"apko/{variant}.yaml"),
            keyring_append=keys.file("melange.rsa.pub"),
            repository_append=repository,
            tag=tag,
            arch=platforms,
        )
        return apko_build.as_directory()

    @function
    async def build_repository(
        self,
        signing_key: Annotated[dagger.Secret | None, Doc("Signing key")] = None,
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = None,
    ) -> dagger.Directory:
        """Build the ggbridge APK packages using melange"""
        melange: dagger.Melange = dag.melange()
        return await melange.with_build(
            config=self.source.file("melange/wstunnel.yaml"),
            signing_key=signing_key,
            arch=platforms,
        ).build(
            config=self.source.file("melange/ggbridge.yaml"),
            source_dir=self.source.filter(
                include=["docker/", "go.mod", "go.sum", "main.go"]
            ),
            signing_key=signing_key,
            arch=platforms,
        )

    @function
    async def build_chart(
        self,
        version: Annotated[str, Doc("Helm chart version")] = "0.0.0",
        app_version: Annotated[str, Doc("App version")] = "unstable",
    ) -> dagger.File:
        """Build the Helm chart"""
        await self.test_chart()
        return self.helm.package(
            source=self.source.directory("helm/ggbridge"),
            version=version,
            app_version=app_version,
        )
