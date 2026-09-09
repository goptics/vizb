---
title: "Deploying"
description: "Deploy Vizb HTML reports to GitHub Pages, Netlify, Cloudflare Pages, and more."
---

Vizb generates a single self-contained HTML file. Upload that file (or a directory that contains it) to any static host.

## Platforms

  ### GitHub Pages

Deploy using [peaceiris/actions-gh-pages](https://github.com/peaceiris/actions-gh-pages):

  ```yaml
  - uses: peaceiris/actions-gh-pages@v4
    with:
      github_token: ${{ secrets.GITHUB_TOKEN }}
      publish_dir: .
  ```

  Configure GitHub Pages to serve from the `gh-pages` branch in **Settings → Pages**.

  ### Netlify

Deploy using [Netlify CLI action](https://github.com/marketplace/actions/netlify-deploy):

  ```yaml
  - uses: nwtgck/actions-netlify@v3
    with:
      publish-dir: '.'
      production-branch: main
      production-deploy: true
    env:
      NETLIFY_AUTH_TOKEN: ${{ secrets.NETLIFY_AUTH_TOKEN }}
      NETLIFY_SITE_ID: ${{ secrets.NETLIFY_SITE_ID }}
  ```

  ### Cloudflare Pages

Use [Cloudflare Pages action](https://github.com/cloudflare/wrangler-action):

  ```yaml
  - uses: cloudflare/wrangler-action@v3
    with:
      apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
      accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
      command: pages deploy . --project-name=my-benchmarks
  ```

  ### Artifact Only

Upload as a workflow artifact for manual download:

  ```yaml
  - uses: actions/upload-artifact@v4
    with:
      name: benchmark-report
      path: benchmark.html
  ```

> Since vizb generates a single HTML file with no external dependencies, any platform that can serve static files works.
