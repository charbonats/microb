# python-journey

Describe your project here.

## Build docker image

The docker image will respect the `.python-version` file in this directry, and will use the `requirements.lock` file to install the dependencies.

Due to the `flavor` option in the `pyproject.toml` file, the docker image will be built upon an `alpine` base image.

```bash
docker build -t example-with-rye -f pyproject.toml .
```
