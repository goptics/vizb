# Security Policy

## Reporting a Vulnerability

If you discover a security issue in Vizb, please report it privately rather than
opening a public issue.

**Email or DM the maintainers via [GitHub Security Advisories](https://github.com/goptics/vizb/security/advisories/new)**
or open a confidential issue if advisories are unavailable.

Include:

- A description of the vulnerability
- Steps to reproduce
- Impact assessment (what an attacker could achieve)
- Suggested fix, if you have one

We aim to acknowledge reports within a few business days and will keep you
updated on remediation progress.

## Scope

Vizb generates self-contained HTML files from user-supplied benchmark output,
CSV, and JSON. Treat all input as untrusted when embedding generated HTML in
shared or hosted environments. Do not serve untrusted vizb output from the same
origin as sensitive applications without reviewing the content first.

Most CLI commands run locally, but `vizb serve` exposes an unauthenticated HTTP
API. The Docker image runs this API on `0.0.0.0:8080` inside the container. Do
not publish it on a non-loopback host address without your own authentication,
TLS, and appropriate access controls, such as a protected reverse proxy or
firewall rules.

The other primary risk surface is malicious or malformed input files processed
by the parsers and rendered into HTML/JavaScript bundles.

## Supported Versions

Security fixes are applied to the latest release on the `main` branch. Older
releases may not receive backports unless the issue is critical.
