# Goods practices to follow

## Generic

:warning:**You must never store credentials information into source code or config file in a GitHub repository**
- Block sensitive data being pushed to GitHub by git-secrets or its likes as a git pre-commit hook
- Audit for slipped secrets with dedicated tools
- Use environment variables for secrets in CI/CD (e.g. GitHub Secrets) and secret managers in production

## Dependencies vulnerabilities scanner

You should run a vulnerability scanner every time you add a new dependency in projects :

```sh
poetry run -m python safety check
```

# Security Policy

## Supported Versions

Use this section to tell people about which versions of your project are currently being supported with security updates.

The current versions are supported

| Version | Supported          |
|---------|--------------------|
| 0.1.0   | :white_check_mark: |

## Reporting a Vulnerability

Report the vulnerabilities in this repository's issue tracker.    

Give the proof of the vulnerability: CVE, analysis report, etc...

Precise how it concerns the implementation of Freyja.

You can ask for support by contacting security@opensource.thalesgroup.com

## Security Update policy

You will get update of the vulnerabilities you have found through the issue tracker.

## Disclosure policy

The policy disclosure will depend on the context of the vulnerability, the proof provided to detect it and the means implemented to remediate.

The result will be discussed in the issue tracker.

## Security related configuration

Freyja is intended to be used in development environments and not in production contexts.

## Known security gaps & future enhancements

### Apparmor 

Freyja currently requires exception rules in Apparmor to work.  
This will be addressed in a future release.
