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
#syntax=gucharbon/microb                         # [1]  Enable automatic microb syntax support

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

[tool.microb.default]
api_version = "v1"                               # [2] Configure the microb API version used
python_version = "3.11"                          # [3] Configure the python interpreter version to use
build_deps = ["build-essential", "libffi-dev"]   # [4] Additional apt packages to install during build
env = { "FOO": "bar" }                           # [5] Additional environment variables to set in final image
indices = [{ "url": "https://pypi.org/simple" }] # [6] Configure pip index to use
labels = { "com.example.foo": "bar" }            # [7] Additional labels to add to the final image
entrypoint = ["micro", "run"]                    # [8] Configure the entrypoint used in the final image
command = ["example:setup"]                      # [9] Configure the command used in the final image
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

Several build configurations can be defined in a single `pyproject.toml`. Each build configuration is defined in a separate section under `[tool.microb]`. The default build configuration is defined under `[tool.microb.default]`.

The frontend is compatible with linux, windows and mac. It also supports various cpu architectures.
Currently `i386`, `amd64`, `arm/v6`, `arm/v7`, `arm64/v8` are supported. Buildkit automatically picks the right version
for you from docker hub.

Available configuration options are listed in the table below.

|     | required | description                                                                                                                                                                                                                                                                              | default | type                    |
|-----|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|---------|-------------------------|
| 1   | no      | instruct Docker to use `pyproject.toml` syntax for parsing this file                                                                                                                                                                                                                           | -       | docker syntax directive |
| 2   | no       | api version of `microb` frontend. This is mainly due to future development to prevent incompatibilities                                                                                                                                                                                 | v1      | enum: [`v1`]            |
| 3   | no      | the python interpreter version to use. Versions format is: `3`, `3.9` or `3.9.1`                                                                                                                                                                                                         | -       | string                  |
| 4   | no       | additional `apt` packages to install before staring the build. These are not part of the final image                                                                                                                                                                                     | -       | string[]                |
| 5   | no       | additional environment variables. These are present in the build and in the run stage                                                                                                                                                                                                    | -       | map\[string]\[string]   |
| 6   | no       | additional list of index to consider for installing dependencies. The only required filed is `url`.                                                                                                                                                                             
| 7   | no       | additional labels to add to the final image. These have precedence over automatically added                                                                                                                                                                                              | -       | map\[string]\[string]   |
| 8   | no       | the entrypoint to use in the final image. This is the command that is run when the container starts                                                                                                                                                                                      | -       | string[]                |
| 9   | no       | the command to use in the final image. This is the command that is run when the container starts if no arguments are given                                                                                                                                                               | -       | string[]                |

#### Index

| name     | required | description                                                                                                 | default | type    |
|----------|----------|-------------------------------------------------------------------------------------------------------------|---------|---------|
| url      | yes      | url of the additional index                                                                                 | -       | string  |
| username | no       | optional username to authenticate. If you got a token for instance, as single factor, just set the username | -       | string  |
| password | no       | optional password to use. If username is not set, this is ignored                                           | -       | string  |
| trust    | no       | used to add the indices domain as trusted. Useful if the index uses a self-signed certificate or uses http  | false   | boolean |

The [example folder](example) contains a few examples how you can use `microb`.

## Recommendations for using `microb`

- use `https` in favor of `http` if possible (for registries, for direct `whl` files and for `git`)
- try to avoid setting `trust` in an index definition, rather use a trusted `https` url
- prefer `git+ssh://git@github.com/moskomule/anatome.git` over http / https links
  like `git+https://user:secret@github.com/moskomule/anatome.git`
- in general prefer setting up an index under the `indices` key for authentication of existing pip registries, rather
  than using in-url credentials

## Build `microb`

`microb` can be build with every docker buildkit compatible cli. The following are a few examples:

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

| name       |              description              |    type |       default |
|------------|:-------------------------------------:|--------:|--------------:|
| llb        |     output created llb to stdout      | boolean |         false |
| dockerfile | print equivalent Dockerfile to stdout | boolean |         false |
| buildkit   |  connect to buildkit and build image  | boolean |          true |
| filename   |           path to pyproject.toml            |  string | pyproject.toml |

For instance to show the created equivalent Dockerfile, use the
command `go run ./cmd/microb/main.go -dockerfile -filename example/minimal/pyproject.toml`.

## Credits

- https://github.com/cmdjulian/mopy
- https://earthly.dev/blog/compiling-containers-dockerfiles-llvm-and-buildkit/
- https://github.com/moby/buildkit/blob/master/docs/merge%2Bdiff.md
- https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/syntax.md
- https://gitlab.wikimedia.org/repos/releng/blubber
