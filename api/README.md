# API contract

`openapi.yaml` is Vizb's canonical public REST contract. It describes only
`POST /`, `POST /merge`, and `POST /ui`.

Errors use `application/problem+json`: malformed JSON returns `400`, bodies
over 10 MiB return `413`, and valid JSON that violates schema or semantic rules
returns `422`. Unknown paths return `404`; unsupported methods return `405`
with `Allow: POST`.

Run the contract checks from this directory:

```bash
pnpm install --frozen-lockfile
pnpm lint
pnpm bundle
go test -count=1 .
```

The Go test resolves every local reference, validates all operation examples
against the documented subset of JSON Schema, and compares reusable Dataset and
chart schemas with the Go wire structs. Redocly performs the full OpenAPI 3.1
lint and dereferenced-bundle validation.
