# Public-contract review checklist

Complete this checklist in the pull request that changes `openapi.yaml`. The
automated checks cannot decide the compatibility and product-policy questions.

- [ ] Operation IDs, schema names, and field names are suitable for generated clients.
- [ ] Dataset and chart configuration schemas remain compatible with the supported Vizb JSON wire format.
- [ ] The contract exposes only `POST /`, `POST /merge`, and `POST /ui`; remote URLs, filesystem paths, persistence, jobs, authentication, and command execution remain excluded.
- [ ] Every request object's strictness is deliberate, including unknown fields and chart options that are inapplicable to a selected chart.
- [ ] `application/json`, `text/html`, and `application/problem+json` media types and the `400`, `406`, `415`, `422`, and `500` semantics are consistent across applicable operations.
- [ ] Conversion, merge, UI, validation, and processing-failure examples describe actual handler behavior.
- [ ] The endpoint owners for `POST /`, `POST /merge`, and `POST /ui` have reviewed their operation against the shared components.
