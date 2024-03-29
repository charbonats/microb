[![Docker Pulls](https://badgen.net/docker/pulls/gucharbon/microb?icon=docker&label=pulls)](https://hub.docker.com/r/gucharbon/microb/)
[![Docker Image Size](https://badgen.net/docker/size/gucharbon/microb?icon=docker&label=image%20size)](https://hub.docker.com/r/gucharbon/microb/)

# `microb` - build python based container images the easy way

> Microb stands for Micro Build. The name originates from the fact that I wanted a simple and easy way to build Docker images for [NATS Micro services](https://charbonats.github.io/nats-micro)

🐳 `microb` is an alternative to the Dockerfile format for creating best practice Python based container images. It relies on the same `pyproject.toml` file which is already used to configure Python projects.

As a buildkit frontend, `microb` does not need to be installed. It is seamlessly integrated and run by [docker buildkit](https://github.com/moby/buildkit)
(respectively [docker](https://github.com/docker/buildx)).  
Create best practice docker images for packaging your python app with ease, without beeing a docker pro!  

## Pyproject.toml

Modern Python packages can contain a `pyproject.toml` file, first introduced in [PEP 518](https://peps.python.org/pep-0518/) and later expanded in [PEP 517](https://peps.python.org/pep-0517/), [PEP 621](https://peps.python.org/pep-0621/) and [PEP 660](https://peps.python.org/pep-0660/). This file contains build system requirements and information, which are used by pip or other package managers to build the package.

`microb` uses the `pyproject.toml` file to gather the required dependencies and build the container image. No need for an additional Dockerfile or a separate build configuration !

[//]: # (@formatter:off)

```toml
#syntax=gucharbon/microb                           # [1]  Enable automatic microb syntax support
[project]
name = "my_example"
authors = [
    {name = "Guillaume Charbonnier", email = "gu.charbon@gmail.com"},
]
requires-python = ">=3.8,<3.12"
dependencies = [
    "nats-py",
    "nkeys"
]
dynamic = ["version"]

[tool.setuptools]
py-modules = ["example"]

[tool.setuptools.dynamic]
version = {attr = "example.__version__"}

[tool.microb.target.default]
api_version = "v1"                                 # [2] Configure the microb API version used
python_version = "3.11"                            # [3] Configure the python interpreter version to use
build_deps = ["build-essential", "libffi-dev"]     # [4] Additional apt packages to install during build (not installed in final image)
system_deps = ["gettext"]                          # [5] Additional apt packages to install in final image (not installed in build image)
env = { "FOO" = "bar" }                            # [6] Additional environment variables to set in final image
indices = [{ "url" = "https://pypi.org/simple" }]  # [7] Configure pip index to use
labels = { "com.example.foo" = "bar" }             # [8] Additional labels to add to the final image
entrypoint = ["micro", "run"]                      # [9] Configure the entrypoint used in the final image
command = ["example:setup"]                        # [10] Configure the command used in the final image
```

[//]: # (@formatter:on)

The above `pyproject.toml` file can be used to produce a docker image with the following command:

```bash
docker build -t example:latest -f pyproject.toml .
```

The most important part of the file is the first line `#syntax=gucharbon/microb`. It tells docker buildkit to use the
`microb` frontend. This can also be achieved by setting the frontend to solve the dockerfile by the running engine itself.
For instance for the docker build command one can append the following build-arg to tell docker to use `microb` without
the in-file syntax directive: `--build-arg BUILDKIT_SYNTAX=gucharbon/microb:v1`. However, the recommended way is to set it
in the `pyproject.toml`, as this is independent of the used builder cli.

The `pyproject.toml` file is a standard file for python projects. It is used to configure the project and its dependencies.

Several build configurations can be defined in a single `pyproject.toml`. Each build configuration is defined in a separate section under `[tool.microb.target]`. The default build configuration is defined under `[tool.microb.target.default]`. To build a specific target, use the `microb_target` build argument:

```bash
docker build -t example:latest --build-arg microb_target=default -f pyproject.toml .
```

The frontend is compatible with linux, windows and mac. It also supports various cpu architectures.
Currently `i386`, `amd64`, `arm/v6`, `arm/v7`, `arm64/v8` are supported. Buildkit automatically picks the right version
for you from docker hub.

Available configuration options are listed in the table below.

|     | name                      | required | description                                                                                                                                                                                                                                                                                                                                 | default | type                    |
| --- | ------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- | ----------------------- |
| 1   | -                         | no       | instruct Docker to use `pyproject.toml` syntax for parsing this file                                                                                                                                                                                                                                                                        | -       | docker syntax directive |
| 2   | `api_version`             | no       | api version of `microb` frontend. This is mainly due to future development to prevent incompatibilities                                                                                                                                                                                                                                     | `"v1"`  | enum: `["v1"]`          |
| 3   | `python_version`          | no       | the python interpreter version to use. Versions format is: `3`, `3.9` or `3.9.1`. If a file named `.python-version` is present in the build context, this variable defaults to the version written in `.python-version`.                                                                                                                                                                                                                                                            | -       | `string`                |
| 4   | `build_deps`              | no       | additional [`apt` packages](https://packages.debian.org/search?keywords=apt) to install before staring the build. These are not part of the final image. When `flavor` is `"alpine"`, packages must be valid [`apk` packages](https://pkgs.alpinelinux.org/packages) instead.                                                                                                                                                                                                                                        | -       | `string[]`              |
| 5   | `system_deps`             | no       | additional [`apt` packages](https://packages.debian.org/search?keywords=apt) to install in the final image. These are not part of the build image. When `flavor` is `"alpine"`, packages must be valid [`apk` packages](https://pkgs.alpinelinux.org/packages).                                                                                                                                                                                                                                              | -       | `string[]`              |
| 6   | `env`                     | no       | additional [environment variables](https://docs.docker.com/reference/dockerfile/#env). These are present in the build and in the run stage. It's possible to use shell substitution to use a value provided as a build argument.                                                                                                                                                                 | -       | `map[string][string]`   |
| 7   | `indices`                 | no       | additional list of index to consider for installing dependencies. The only required filed is `url`.                                                                                                                                                                                                                                         | -       | `Index[]`               |
| 8   | `labels`                  | no       | additional [labels](https://docs.docker.com/config/labels-custom-metadata/) to add to the final image. These have precedence over automatically added. It's possible to use shell substitution to use a value provided as a build argument.                                                                                                                                                           | -       | `map[string][string]`   |
| 9   | `entrypoint`              | no       | the [entrypoint](https://docs.docker.com/reference/dockerfile/#entrypoint) to use in the final image. This is the command that is run when the container starts                                                                                                                                                                                                                                         | -       | `string[]`              |
| 10  | `command`                 | no       | the [command](https://docs.docker.com/reference/dockerfile/#cmd) to use in the final image. This is the command that is run when the container starts if no arguments are given                                                                                                                                                                                                                  | -       | `string[]`              |
| -   | `extras`                  | no       | install additional [extra dependency group](https://packaging.python.org/en/latest/specifications/pyproject-toml/#dependencies-optional-dependencies). Each extra must be an optional dependency group defined in the pyproject.toml                                                                                                                                                                                                                    | -       | `string[]`              |
| -   | `requirements`            | no       | Path to a [requirements.txt](https://pip.pypa.io/en/stable/reference/requirements-file-format/) file used to install project dependencies. When requirements is specified, extras cannot be used, and dependencies listed in pyproject.toml are ignored. Use requirements when project dependencies are locked using a third-party tool like pip-tools or poetry and can be exported as a requirements.txt file. | -       | `string`                |
| -   | `copy_files`              | no       | additional files to [copy](https://docs.docker.com/reference/dockerfile/#copy) into the final image. Files are not copied to the build stage.                                                                                                                                                                                                                                                     | -       | `Copy[]`                |
| -   | `add_files`               | no       | additional files to [add](https://docs.docker.com/reference/dockerfile/#add) into the final image. Files are not added to the build stage.                                                                                                                                                                                                                                                       | -       | `Add[]`                 |
| -   | `copy_files_before_build` | no       | additional files to [copy](https://docs.docker.com/reference/dockerfile/#copy) into the build stage. Files are not copied to the final image.                                                                                                                                                                                                                                                     | -       | `Copy[]`                |
| -   | `add_files_before_build`  | no       | additional files to [add](https://docs.docker.com/reference/dockerfile/#add) into the build stage. Files are not added to the final image.                                                                                                                                                                                                                                                       | -       | `Add[]`                 |
| - | `flavor` | no | flavor to use for the base image. The flavor is used to select a base image with additional tools and libraries. Use `alpine` image if you want to reduce image size, but be careful as it might require additional build dependencies (alpine base image comes without many tools installed in the debian base image). | `"debian"` | enum: `["debian", "alpine"]` |

#### Copy

> Refer to https://docs.docker.com/reference/dockerfile/#copy for more information. Note that only `--from` option is supported. The other options (such as `--chmod` or `--chown`) are not currently supported.

| name   | required | description                          | default | type     |
| ------ | -------- | ------------------------------------ | ------- | -------- |
| `src`  | yes      | source path                          | -       | `string` |
| `dst`  | yes      | destination path                     | -       | `string` |
| `from` | no       | stage, context or image to copy from | -       | `string` |


#### Add

> Refer to https://docs.docker.com/reference/dockerfile/#add for more information. Note that only `--checksum` option is supported. The other options (such as `--chown`) are not currently supported.

| name       | required | description                            | default | type     |
| ---------- | -------- | -------------------------------------- | ------- | -------- |
| `src`      | yes      | source path                            | -       | `string` |
| `dst`      | yes      | destination path                       | -       | `string` |
| `checksum` | no       | checksum used to verify file integrity | -       | `string` |

#### Index

| name              | required | description                                                                                                 | default | type      |
| ----------------- | -------- | ----------------------------------------------------------------------------------------------------------- | ------- | --------- |
| `url`             | yes      | url of the additional index                                                                                 | -       | `string`  |
| `username`        | no       | optional username to authenticate. If you got a token for instance, as single factor, just set the username | -       | `string`  |
| `username_secret` | no       | optional id of secret containing the username. This option takes precedence over `username`.                | -       | `string`  |
| `password`        | no       | optional password to use. If username is not set, this is ignored                                           | -       | `string`  |
| `password_secret` | no       | optional id of secret containing the password. This option takes precedence over `password`.                | -       | `string`  |
| `trust`           | no       | used to add the indices domain as trusted. Useful if the index uses a self-signed certificate or uses http  | `false` | `boolean` |

Note that when `username_secret` or `password_secret` are used, the secrets must be provided to the build command. For example:

```bash
docker buildx build --secret id=az_feed_token,env=AZ_FEED_TOKEN -t example -f pyproject.toml .
```

> In the example above, the secret `az_feed_token` is provided using the value from the environment variable `AZ_FEED_TOKEN` on the host. Checkout https://docs.docker.com/build/building/secrets/ to learn more about build secrets. 

The [example folder](example) contains a few examples how you can use `microb`.

## Which problems does this solve ?

- **No need to learn Dockerfile syntax**: `microb` uses the `pyproject.toml` file to gather the required dependencies and build the container image. No need for an additional Dockerfile or a separate build configuration !

- **No need to install an additional tool**: `microb` is a buildkit frontend. It is seamlessly integrated and run by [buildkit](https://docs.docker.com/build/buildkit/).

- **No need to choose a base image**: `microb` uses the python version defined in the `pyproject.toml` file to find the right base image for you.

- **Automatically install python project**: `microb` automates the installation of a python package and its dependencies within a container image.

- **Multi-stage build: smaller image**: `microb` uses the `pyproject.toml` file to gather the required dependencies and install them in the build stage. The final image only contains the required runtime dependencies.

- **Install pip dependencies using SSH easily**: `microb` takes care of adding an [SSH mount](https://docs.docker.com/reference/dockerfile/#run---mounttypessh) to the pip command used to install packages during build.

- **Install dependencies from third-party index**: `microb` allows you to configure additional one or sevreal pip indices to install dependencies from. Build secrets can be used to securely authenticate against the index.

- **Install apt dependencies easily**: `microb` takes care of adding a [cache mount](https://docs.docker.com/reference/dockerfile/#example-cache-apt-packages) to the apt command used to install packages during build for caching and clean up the `apt` cache if any package is installed in the final image.


## Recommendations for using `microb`

- use `https` in favor of `http` if possible (for registries, for direct `whl` files and for `git`)
- try to avoid setting `trust` in an index definition, rather use a trusted `https` url
- prefer `git+ssh://git@github.com/moskomule/anatome.git` over http / https links
  like `git+https://user:secret@github.com/moskomule/anatome.git`
- in general prefer setting up an index under the `indices` key for authentication of existing pip registries, rather
  than using in-url credentials

## Build images with `microb`

`microb` can be used with every docker buildkit compatible cli. The following are a few examples:

#### docker:

```bash
DOCKER_BUILDKIT=1 docker build --ssh default --build-arg BUILDKIT_SYNTAX=gucharbon/microb:v1 -t example:latest -f pyproject.toml .
```

In that particular case even the syntax directive from `[1]` is not required anymore, as it is set on the `docker build`
command directly.  
If the syntax directive is set in the `pyproject.toml`, `--build-arg BUILDKIT_SYNTAX=gucharbon/microb:v1` can be omitted in the
command.

#### nerdctl:

```bash
nerdctl build --ssh default -t example:latest -f pyproject.toml .
```

#### buildctl:

```bash
buildctl build \
--frontend=gateway.v0 \
--opt source=gucharbon/microb:v1 \
--ssh default \
--local context=. \
--local dockerfile=pyproject.tml \
--output type=docker,name=example:latest \
| docker load
```

In that particular case even the syntax directive from `[1]` is not required anymore, as it is set on the `buildctl`
command directly.  
If the syntax directive is set in the `pyproject.toml`, `--opt source=gucharbon/microb:v1` can be omitted in the command.

The resulting image is build as a best practice docker image and employs a multistage build- It
uses [official python slim images](https://hub.docker.com/_/python) image as final base image. It runs as
non-root user and only includes the minimal required runtime dependencies.

### SSH dependencies

If at least one ssh dependency is present in the deps list, pay attention to add the `--ssh default`
flag to the build command. Also make sure, that your ssh-key is loaded inside the ssh agent.  
If you receive an error `invalid empty ssh agent socket, make sure SSH_AUTH_SOCK is set` your SSH agent is not running
or improperly set up. You can start or configure it and adding your ssh key by executing:

```bash
eval `ssh-agent`
ssh-add /path/to/ssh-key
```

The `ssh` flag is only required if you're including a ssh dependency. If no ssh dependency is present, the ssh flag can
be omitted.

## Run a container from the built image

The built image can be run like any other container:

```bash
$ docker run --rm example:latest
```

## microb development

### Installation as cmd

```bash
$ go install github.com/charbonats/microb
```

### Arguments

The following arguments are supported running the frontend:

| name       |              description              |      type |            default |
| ---------- | :-----------------------------------: | --------: | -----------------: |
| llb        |     output created llb to stdout      | `boolean` |            `false` |
| dockerfile | print equivalent Dockerfile to stdout | `boolean` |            `false` |
| buildkit   |  connect to buildkit and build image  | `boolean` |             `true` |
| filename   |        path to pyproject.toml         |  `string` | `"pyproject.toml"` |

For instance to show the created equivalent Dockerfile, use the
command `go run ./cmd/microb/main.go -dockerfile -filename example/debian/pyproject.toml`.

## Example generated Dockerfile

The example present in [example/debian](example/debian) contains a [`pyproject.toml`](example/debian/pyproject.toml) file. The dockerfile produced by `microb` is the following:

```Dockerfile
FROM docker.io/python:3.11 AS builder

RUN --mount=type=cache,target=/var/cache/apt --mount=type=cache,target=/var/lib/apt apt-get update && apt-get install -y --no-install-recommends build-essential libffi-dev

ENV PIP_DISABLE_PIP_VERSION_CHECK=1
ENV PIP_NO_WARN_SCRIPT_LOCATION=0
ENV PIP_USER=1
ENV PYTHONPYCACHEPREFIX=/.pycache
RUN --mount=type=cache,target=/root/.cache --mount=type=ssh,required=true GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no' python -m pip install --user --retries 2 --extra-index-url https://pypi.org/simple nats-py nats-micro@git+ssh://git@github.com/charbonats/nats-micro.git nkeys
COPY . /projectdir
RUN --mount=type=cache,target=/root/.cache python -m pip install --no-deps /projectdir
RUN find /root/.local/lib/python*/ -name 'tests' -exec rm -r '{}' + && find /root/.local/lib/python*/site-packages/ -name '*.so' -exec sh -c 'file "{}" | grep -q "not stripped" && strip -s "{}"' \; && find /root/.local/lib/python*/ -type f -name '*.pyc' -delete && find /root/.local/lib/python*/ -type d -name '__pycache__' -delete

FROM python:3.11-slim

RUN apt-get update && apt-get install -y --no-install-recommends  gettext  && rm -rf /var/lib/apt/lists/*
RUN useradd --uid=65532 --user-group --home-dir=/home/nonroot --create-home nonroot
USER 65532:65532

COPY --from=builder /root/.local /home/nonroot/.local
ENV PATH=$PATH:/home/nonroot/.local/bin

ENTRYPOINT ["micro","run"]
CMD ["example:setup"]

LABEL org.opencontainers.image.description="autogenerated by microb"
LABEL moby.buildkit.frontend="microb"
LABEL microb.version="v1"
LABEL com.example.foo="World"
LABEL org.opencontainers.image.authors="Guillaume Charbonnier <gu.charbon@gmail.com>, Someone else <someone.else@gmail.com>"
```

## Credits

- https://github.com/cmdjulian/mopy
- https://earthly.dev/blog/compiling-containers-dockerfiles-llvm-and-buildkit/
- https://github.com/moby/buildkit/blob/master/docs/merge%2Bdiff.md
- https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/syntax.md
- https://gitlab.wikimedia.org/repos/releng/blubber
