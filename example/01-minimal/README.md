# Minimal example

This is a minimal example of building a docker image out of a `pyproject.toml`.

The project and the pyproject.toml file do not contain any configuration specific to `microb`.

A docker image can still be built using the following command:

```bash
docker build --build-arg BUILDKIT_SYNTAX=gucharbon/microb -f pyproject.toml -t minimal .
```

By using `--build-arg BUILDKIT_SYNTAX=gucharbon/microb`, the `pyproject.toml` is used as a Dockerfile.

The produced image does not have any entrypoint or command defined, but has the python project installed.
