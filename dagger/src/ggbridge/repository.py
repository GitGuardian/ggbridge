from typing import Annotated

import dagger
from dagger import Doc, Name, function, object_type


@object_type
class Repository:
    """repository"""

    source: dagger.Directory
    melange: dagger.Melange

    @function
    async def container(
        self,
    ) -> dagger.Container:
        """Return container"""
        return self.melange.container().with_mounted_directory(
            "$MELANGE_WORK_DIR", source=await self.build(), expand=True
        )

    @function
    async def build(
        self,
        platforms: Annotated[
            list[dagger.Platform] | None, Doc("Target platforms"), Name("platform")
        ] = None,
    ) -> dagger.Directory:
        """Build APK repository"""
        if not await self.melange.has_signing_key():
            self.melange = self.melange.with_keygen()
        # Build wstunell/ggbridge packages and add the melange public key
        return await self.melange.with_build(
            config=self.source.file("melange/wstunnel.yaml"), arch=platforms
        ).build(
            config=self.source.file("melange/ggbridge.yaml"),
            source_dir=self.source.filter(
                include=["docker/", "go.mod", "go.sum", "main.go"]
            ),
            arch=platforms,
        )
