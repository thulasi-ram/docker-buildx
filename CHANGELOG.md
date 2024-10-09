# Changelog

## [5.0.0](https://codeberg.org/woodpecker-plugins/docker-buildx/releases/tag/v5.0.0) - 2024-10-09

### ❤️ Thanks to all contributors! ❤️

@6543, @flyingscorpio, @pat-s, @woodpecker-bot

### 💥 Breaking changes

- Hide insecure configuration options behind insecure tag [[#194](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/194)]
- chore(deps): update docker docker tag to v27 [[#170](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/170)]

### ✨ Features

- Add `sbom` option [[#187](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/187)]

### 📈 Enhancement

- Optimize container layers [[#191](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/191)]
- Dont let proxy config overwrite dedicated buildkit_driveropt proxy options if set [[#175](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/175)]

### 🐛 Bug Fixes

- Quote `no_proxy` [[#185](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/185)]
- Update drone-plugin-lib to handle nil pointer [[#178](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/178)]

### 📚 Documentation

- update pipeline config example [[#180](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/180)]

### 📦️ Dependency

- chore(deps): update docker docker tag to v27.2.1 [[#183](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/183)]
- fix(deps): update github.com/urfave/cli/v3 digest to 31c5c84 [[#182](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/182)]
- Update to github.com/urfave/cli/v3 [[#173](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/173)]

### Misc

- dont ignore changelog ([f3f8166](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/f3f8166af3c420310ddd1dca69158a6bc70cb288))
- Add nix flake [[#192](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/192)]
- chore(deps): update dependency go to v1.23.2 ([9c4cea3](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/9c4cea3cccce7b3b668ed5a078d71e4e0b652ec4))
- fix(deps): update github.com/urfave/cli/v3 digest to 20ef97b ([b6a6f1c](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/b6a6f1c7728729882ce616a78aeb7bdda2de19be))
- Add release config for release-helper plugin [[#190](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/190)]
- Use ready-release-go plugin [[#189](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/189)]
- add-remote-builders [[#179](https://codeberg.org/woodpecker-plugins/docker-buildx/pulls/179)]
- chore(deps): update docker docker tag to v27.3.1 ([4fedea4](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/4fedea48552f8e7584abb39aa50a9e8e4d1aaea9))
- chore(deps): update docker/buildx-bin docker tag to v0.17.1 ([b807eda](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/b807edaa48ff0aadd5f5e154c40ba74956e2e2a9))
- chore(deps): update davidanson/markdownlint-cli2 docker tag to v0.14.0 ([8435819](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/84358199716d7e6b4529e454772dccf50ca5e067))
- chore(deps): update dependency go to v1.23.1 ([3ee6b8b](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/3ee6b8bb9dd1170434f48b7f9d0a389e1c85a569))
- chore(deps): update docker docker tag to v27.2.0 ([c661b7a](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/c661b7a48d39b4329944f435727c06af9336a6ee))
- fix(deps): update github.com/urfave/cli/v3 digest to 3d76e1b ([ba44638](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/ba44638ddaba4aeff7fe9d59caa0f73e47f20a70))
- fix(deps): update module github.com/pelletier/go-toml/v2 to v2.2.3 ([0e1e881](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/0e1e8818e72d6fae335635be701fa43e6c10ae6b))
- fix(deps): update github.com/urfave/cli/v3 digest to 5386081 ([cf7d19c](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/cf7d19cbb1b1f94ce0d4cf868dde9d75c8b2aab5))
- fix(deps): update module honnef.co/go/tools to v0.5.1 ([23b0f61](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/23b0f6125596bdf4877262676e59490f3ac577dd))
- chore(deps): update golang docker tag to v1.23 ([b98307b](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/b98307bb3e495969737aba7dc861838720de3c21))
- fix(deps): update github.com/urfave/cli/v3 digest to 3110c0e ([13aa887](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/13aa8879a21410c5ac9a83515aa9a19ca75243cb))
- chore(deps): update docker docker tag to v27.1.2 ([fe28f12](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/fe28f12be182c261560c10e6b21c487e6beaee2f))
- chore(deps): update woodpeckerci/plugin-release docker tag to v0.2.1 ([ea32257](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/ea32257d1d634daf2132dae551473cd7d473df07))
- fix(deps): update module github.com/aws/aws-sdk-go to v1.55.5 ([af71dcd](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/af71dcdbdef4a03cda1e6ac4779d75a0263293f8))
- fix(deps): update module github.com/aws/aws-sdk-go to v1.55.3 ([1d5b8fd](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/1d5b8fdc17a39cb5469c05b392cedccf06437bd3))
- chore(deps): update docker/buildx-bin docker tag to v0.16.2 ([c671088](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/c671088766b7cb467c3cb10f0d7b1069c59e853a))
- chore(deps): update woodpeckerci/plugin-docker-buildx docker tag to v4.2.0 ([642f2b3](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/642f2b36d7736b0f875c985228a41ea4a8419078))
- chore(deps): update docker/buildx-bin docker tag to v0.16.1 ([f8b58f5](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/f8b58f573b2b56b53466c4e7871032244adbc8ca))
- fix(deps): update github.com/urfave/cli/v3 digest to 127cf54 ([03cd73f](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/03cd73f06864bc0a26c66aaa2e66b07b0e9c81c5))
- update version parse lib ([42395c5](https://codeberg.org/woodpecker-plugins/docker-buildx/src/commit/42395c51eca5407325619b38a90669d480ca4ef9))
