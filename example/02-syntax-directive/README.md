# Syntax directive

This is a minimal example of building a docker image out of a `pyproject.toml`.

The only configuration specific to `microb` is the buildkit syntax directive at the top of the pyproject.toml file:

```toml
#syntax=gucharbon/microb
```

A docker image can be built using the following command:

```bash
docker build -f pyproject.toml -t minimal .
```

Because the syntax directive is present at the top of the file, it is not required to use `--build-arg BUILDKIT_SYNTAX=gucharbon/microb`.

The produced image does not have any entrypoint or command defined, but has the python project installed.
