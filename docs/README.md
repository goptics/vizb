# Vizb documentation site

Astro + Starlight site in this directory. From the repo root:

```bash
task dev:docs    # local dev server
task build:docs  # production build to docs/dist/
```

Content lives in `src/content/docs/`. Sidebar and site config are in `astro.config.mjs`.

Open Graph image source is `scripts/og-image/index.html`. From the repo root:

```bash
google-chrome --headless --disable-gpu --force-device-scale-factor=1 \
  --window-size=1200,600 \
  --screenshot=docs/public/og-image.png \
  "file://$(pwd)/docs/scripts/og-image/index.html"
```