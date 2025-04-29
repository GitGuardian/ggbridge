from typing import Annotated

import dagger
from dagger import Doc, function, object_type


@object_type
class Chart:
    """chart"""

    source: dagger.Directory
    helm: dagger.Helm

    uuid: str

    @function
    async def push(
        self,
        registry: Annotated[str, Doc("Helm registry")] = "",
        version: Annotated[str, Doc("Helm chart version")] = "0.1.0",
        app_version: Annotated[str, Doc("App version")] = "0.1.0",
    ) -> str:
        """Publish the Helm chart and returns the digest"""
        chart: dagger.File = await self.build(version=version, app_version=app_version)
        if not registry:
            registry = f"ttl.sh/{self.uuid}/ggbridge-helm"
        return await self.helm.push(chart=chart, registry=registry)

    @function
    async def build(
        self,
        version: Annotated[str, Doc("Helm chart version")] = "0.1.0",
        app_version: Annotated[str, Doc("App version")] = "0.1.0",
    ) -> dagger.File:
        """Build the Helm chart"""
        await self.lint()
        return self.helm.package(
            source=self.source,
            version=version,
            app_version=app_version,
        )

    @function
    async def lint(
        self,
    ) -> str:
        """Test Helm chart"""
        return await self.helm.lint(self.source, strict=True)
