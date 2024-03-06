# Requirements file

This is a minimal example of building an alpine docker image out of a `pyproject.toml` and a requirements file.

Alpine images are smaller than their debian counterparts. They are a good choice for production images when the project has no dependencies that require a debian image.

The buildkit syntax directive at the top of the pyproject.toml file is used to indicate to Docker how to interpret the file:

```toml
#syntax=gucharbon/microb
```

An additional section in the `pyproject.toml` file is used to configure the requirements file, the entrypoint and command of the produced image. An additional option named `flavor` is used to specify the base image to use. The default value is `debian`:

```toml
[tool.microb.target.default]
flavor = "alpine"
requirements = "requirements.txt"
entrypoint = ["python"]
command = ["-m", "example"]
```

A docker image can be built using the following command:

```bash
docker build -f pyproject.toml -t minimal .
```

Because the syntax directive is present at the top of the file, it is not required to use `--build-arg BUILDKIT_SYNTAX=gucharbon/microb`.

The produced image can be started using the following command:

```bash
docker run -it --rm minimal
```

The produced image has the python project installed and the entrypoint and command are defined. You should see the following output:

```plaintext
Hello, world!
```
