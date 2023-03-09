# CONTRIBUTING

## Contributor License Agreements

This repository use the Apache-2.0 license.

## Contributing repo

To contribute to this project, start creating an issue.

Once the issue is created, create a new branch or a new fork related to it and push your modification
inside.

Speak with the maintainers and the developers, get advice and remain active.

## Contributing code

Install Poetry to contribute to the implementation. Use poetry to add new dependency or to run python
tests.

## Pull Request Checklist

- Test your implementation and verify it with the [SECURITY.md](./SECURITY.md) documentation.  
- Squash your commits into fewer meaningful commits, then create a pull request.
- Notify a maintainer to validate.

### Testing

#### Running sanity check

Use a linter to scan your source code.

A linter will be use by a maintainer to verify the quality of your code.

#### Running unit tests

```sh
poetry run pytest --cov=freyja freyja/tests/
```

#### Running vulnerability scanner

You should run a vulnerability scanner every time you add a new dependency in projects :

```sh
poetry run -m python safety check
```
