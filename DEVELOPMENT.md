# DEVELOPMENT

This guide helps through the development process

## Contribution

Read the [CONTRIBUTING.md guide](CONTRIBUTING.md).

## Development environment

We recommend you to create a dedicated environment for your developments with Pyenv.

```sh
pyenv install 3.9
pyenv virtualenv 3.9 freyja
pyenv activate freyja
pip install --upgrade pip
```

## Running

While you develop, stick to the Poetry usage to leverage your development environment :

```sh
poetry update
poetry install
# use poetry to run freyja development version
poetry run freyja --help
```
