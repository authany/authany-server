# Authany server

The identity server behind [Authany](https://authany.com): Authgear's source with Authany-specific changes on top. Upstream is [authgear/authgear-server](https://github.com/authgear/authgear-server) (Apache-2.0; `LICENSE` and `NOTICE` are kept as-is). The upstream README lives at [`README.md`](../README.md) and is left untouched so merges stay clean.

## Branches

- `main` — the upstream release tag currently deployed, plus Authany commits on top. Authany commits are prefixed `[Portal]`, `[AuthUI]`, `[Server]` and end with `(Authany)`.
- Remote `upstream` — `authgear/authgear-server`. Upstream tags are mirrored here.

Currently based on `2026-08-26.0`.

## What differs from upstream

- `[Portal] Add locale selection and language switcher (Authany)` — the admin console picks its locale from `localStorage` / browser language instead of being hard-wired to English, and gains a Language submenu. Translations themselves are served at runtime through `PORTAL_CUSTOM_RESOURCE_DIRECTORY`.
- `[Portal] Strip residual upstream vendor links (Authany)` — links to the upstream vendor's site, docs, Discord and mailboxes render as plain text; header contact/docs links, Get Started contact and resource columns, and the Billing nav entry are removed; the Starter Kit section is hidden while it points at upstream example repos.

Go services are unmodified; the official images run in production. Only the admin-console frontend built from this repo is deployed, as an override on top of the official portal image.

## Build and deploy

Building, translations, brand assets and deployment live in [teomyth/authany-deploy](https://github.com/teomyth/authany-deploy), which expects this repo checked out at `server/` and runs `i18n/scripts/build.sh`.

## Upgrading to a new upstream release

```bash
git fetch upstream --tags
git merge <tag>            # e.g. 2026-09-30.0; conflicts, if any, are confined to the portal files touched above
git push origin main --tags
```

Then rebuild and redeploy from authany-deploy.
