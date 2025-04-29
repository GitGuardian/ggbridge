from typing import Annotated

import dagger
from dagger import Doc, Name, dag, function, object_type

from .repository import Repository


@object_type
class Image:
    """image"""

    source: dagger.Directory
    repository: Repository
    container_: dagger.Container
    apko: dagger.Apko
    cosign: dagger.Cosign
    crane: dagger.Crane
    grype: dagger.Grype
    uuid: str

    build_: dagger.ApkoBuild | None = None

    @function
    async def publish(
        self,
        tags: Annotated[list[str], Doc("Tags"), Name("tag")] = (),
        version: Annotated[str, Doc("Image version. Used when no tags provided")] = "",
        variant: Annotated[str, Doc("Image variant")] = "prod",
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Platforms"), Name("platform")
        ] = None,
        scan: Annotated[bool, Doc("Scan the image for vulnerabilities")] = True,
        sign: Annotated[bool, Doc("Sign and Attest the image with cosign")] = False,
    ) -> str:
        """Publish the ggbridge image"""
        platforms = platforms or [await dag.default_platform()]
        # Build the image
        build: dagger.ApkoBuild = await self._build(
            variant=variant,
            platforms=platforms,
        )
        build_digest: str = await build.as_tarball().digest()
        self.container_ = self.container_.import_(build.as_tarball())

        if scan:
            # Scan the image for vulnerabilities
            scan_reports: dict[dagger.Platform, dagger.File] = {}
            for platform in platforms:
                scan_reports[platform] = self.grype.scan_file(
                    source=build.tarball(platform=platform),
                    severity="",
                    fail=False,
                    output_format="json",
                )

        # Publish the image
        full_ref: str = ""
        platform_variants: list[dagger.Container] = []
        for platform in platforms:
            if platform != await dag.default_platform():
                platform_variants.append(build.container(platform=platform))

        # When tags not provided, compute image address.
        if not tags:
            # Retrieve image title from config
            registry: str = "ttl.sh"
            repository: str = f"{self.uuid}/ggbridge"
            if not version:
                version = build_digest.split(":")[1][:8]
            if variant != "prod":
                version = f"{version}-shell"
            tags = [f"{registry}/{repository}:{version}"]

        full_ref: str = await self.container_.publish(
            address=tags[0], platform_variants=platform_variants
        )

        # Sign and attest
        if sign:
            # Clean all existing attestations with cosign
            await self.cosign.clean(full_ref)

            # Sign the image with cosign
            await self.cosign.sign(
                image=full_ref,
                recursive=True,
            )

            # Attest SBOMs
            if len(platforms) > 1:
                # Attest index SBOM
                await self.cosign.attest(
                    image=full_ref,
                    predicate=build.sbom_file(),
                    type_="spdxjson",
                )

            # Attest platforms SBOMs
            for platform in platforms:
                platform_digest: str = await self.crane.digest(
                    full_ref, platform=platform, full_ref=True
                )
                await self.cosign.attest(
                    image=platform_digest.strip(),
                    predicate=build.sbom_file(platform=platform),
                    type_="spdxjson",
                )

                if scan_reports:
                    # Attest vulnerability reports
                    await self.cosign.attest(
                        image=platform_digest.strip(),
                        predicate=scan_reports[platform],
                        type_="openvex",
                    )

        # Publish other tags
        for tag in tags[1:]:
            await self.cosign.copy(source=full_ref, destination=tag, force=True)

        return full_ref

    @function
    async def scan(
        self,
        variant: Annotated[str, Doc("Variant to scan")] = "prod",
        severity: Annotated[
            str, Doc("Specify the minimum vulnerability severity to trigger an error")
        ] = "",
        output_format: Annotated[str, Doc("Report output formatter")] = "table",
    ) -> dagger.File:
        """Scan the image using grype"""
        build: dagger.Directory = await self.build(variant=variant)
        return self.grype.scan_file(
            source=build.file("image.tar"),
            source_type="docker-archive",
            severity=severity,
            fail=False,
            output_format=output_format,
        )

    @function
    async def container(
        self,
        variant: Annotated[str, Doc("Variant to build")] = "shell",
    ) -> dagger.Container:
        """Return the image container"""
        build: dagger.Directory = await self.build(variant=variant)
        tarball: dagger.File = build.file("image.tar")
        return dag.container().import_(tarball)

    @function
    async def build(
        self,
        tag: Annotated[str, Doc("Image tag")] = "ggbridge",
        variant: Annotated[str, Doc("Variant to build")] = "prod",
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = None,
    ) -> dagger.Directory:
        """Build the ggbridge image using apko"""
        build: dagger.ApkoBuild = await self._build(
            tag=tag, variant=variant, platforms=platforms
        )
        return build.as_directory()

    async def _build(
        self,
        tag: Annotated[str, Doc("Image tag")] = "ggbridge",
        variant: Annotated[str, Doc("Variant to build")] = "prod",
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = None,
    ) -> dagger.ApkoBuild:
        """Build the ggbridge image using apko and return ApkoBuild"""
        repository: dagger.Directory = await self.repository.build(
            platforms=platforms,
        )
        return self.apko.build(
            config=self.source.file(f"apko/{variant}.yaml"),
            tag=tag,
            keyring_append=[repository.file("melange.rsa.pub")],
            repository_append=[repository],
            arch=platforms,
        )
