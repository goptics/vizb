# API contract

`openapi.yaml` is Vizb's canonical public REST contract. It describes only
`POST /`, `POST /merge`, and `POST /ui`.

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

Use [REVIEW.md](REVIEW.md) for the required human public-contract review.
