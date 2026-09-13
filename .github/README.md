# Authany Server

Authany is a self-hosted identity platform: one deployment, unlimited projects, each project with its own fully isolated user pool, login pages, branding and domain. Think of it as running your own Auth0 for every product, internal tool and customer you have, without per-project deployments and without a vendor's name anywhere your users can see it.

This repository is the server: the Go services, the hosted login/signup UI, the admin console and the platform admin API. Deployment, brand assets, translations and build scripts live in a separate private repository.

## What it does

**Projects as tenants.** Every project is a separate tenant with its own users, credentials, sessions, settings and admins. A person who signs up for one project does not exist in another. Project admins only see their own projects; platform admins see everything.

**Domains.** Each project gets `{project}.authanyid.com` and can bring its own domain (TLS is issued on demand). The platform itself runs on `id.authany.com` (platform login) and `manage.authany.com` (admin console).

**Login methods.** Email or phone with password; passwordless one-time codes over email, SMS and WhatsApp; passkeys (WebAuthn); social and enterprise login through Google, Apple, Facebook, GitHub, LinkedIn, Microsoft Entra ID, Azure AD B2C, ADFS and WeChat; LDAP directories. Multi-factor with TOTP, OTP and recovery codes.

**Protocols.** OpenID Connect and OAuth 2.0 for applications (web, SPA, mobile, machine-to-machine), SAML 2.0 identity provider for enterprise apps, a session resolver for reverse-proxy setups, and a GraphQL Admin API per project.

**Security.** Bot protection (Cloudflare Turnstile, reCAPTCHA), rate limits and account lockout, session management across devices, audit log, roles and groups, hooks that can block or observe events (webhooks and TypeScript hooks).

**Customisation.** Per-project branding of the hosted pages (logo, colours, CSS, templates), email and SMS templates, 23 end-user languages, custom UI via the authentication-flow API. The admin console itself is available in English and Simplified Chinese with an in-app switcher.

## Architecture

| Component | Role |
|---|---|
| `cmd/authgear`, `pkg/lib`, `pkg/auth` | Core identity server: authentication flows, OAuth/OIDC/SAML, sessions, hooks, the hosted UI |
| `cmd/portal`, `pkg/portal`, `portal/` | Admin console backend (GraphQL) and the React frontend project admins use |
| `pkg/siteadmin` | Platform-wide admin API: every project, every collaborator, usage |
| `authui/` | Frontend assets for the hosted login and account pages |
| `resources/` | Built-in templates, translations and static assets, overridable per deployment and per project |

Runtime dependencies: PostgreSQL (app, audit and search databases), Redis, S3-compatible object storage. Project configuration is stored in the database and hot-reloaded. Services run as containers behind Caddy with wildcard and on-demand certificates.

Local development, tests and conventions are documented in [`CONTRIBUTING.md`](../CONTRIBUTING.md) and [`AGENTS.md`](../AGENTS.md).

## Branches and releases

- `main` is what production runs: the current base release plus Authany commits on top. Authany commits follow [Conventional Commits](https://www.conventionalcommits.org/), `type(scope): subject (Authany)`: the scope is `server`, `portal`, `authui` or `ci`, the common types are `feat`, `fix`, `docs`, `test`, `refactor`, `chore` and `ci`, the subject starts lower case, is imperative and carries no full stop, and every commit has a body saying what changed and why. This convention overrides the commit style of the upstream [`AGENTS.md`](../AGENTS.md).
- Base release tags (`YYYY-MM-DD.N`) are mirrored in this repository. Currently based on `2026-09-09.0`.
- Authany changes so far (`git log --oneline <base tag>..main` is the authoritative list; the highlights; the four subjects below were pushed before the convention above and are kept as written):
  - `[Portal] Add locale selection and language switcher (Authany)` — the admin console picks its locale from `localStorage` / browser language and offers a Language submenu; translations are loaded at runtime from the deployment's resource directory.
  - `[Portal] Strip residual upstream vendor links (Authany)` — external links to the base project's website, docs, community and mailboxes render as plain text; header contact/docs links, the Get Started contact and resource columns, the Billing nav entry and the Starter Kit section are removed.
  - `[CI] Replace the upstream CI with an Authany portal check (Authany)` — the workflow below.
  - `[Server] De-brand the collaborator invitation subject (Authany)` — the collaborator invitation email subject says "in Authany". The email body is overridden at deployment level (`PORTAL_CUSTOM_RESOURCE_DIRECTORY/templates/en/messages/`); the subject is hard-coded, so this change only reaches production once the portal image is built from this repository.
  - `feat(server): add seven native SMS providers (Authany)` — aliyun, aliyun_mas, tencent, yunpian, smsbao, gatewayapi and smsaero each get secret configuration, a sending client, resolver wiring, admin console backend support and a provider form.

Taking a new base release:

```bash
git fetch upstream --tags
git merge <tag>                # conflicts, if any, are confined to the portal files touched above
git push origin main --tags
```

Then rebuild the admin console and redeploy from authany-deploy.

## CI

The only workflow that runs here is [`authany-portal.yaml`](workflows/authany-portal.yaml): typecheck, eslint, stylelint, prettier, tests and a build of `portal/`, on pushes to `main` and pull requests that touch `portal/`. Run the same checks locally with `npm run typecheck && npm run eslint && npm run prettier` in `portal/` before pushing.

The upstream workflows (`ci-branches.yaml`, `ci-prs.yaml`, `ci-tags.yaml`, `custom-build.yaml`, `mirror.yaml`, `oursky.yaml`, `periodic-check-*.yaml`, `chromatic.yaml`, `authgear-once.yaml`) are kept in the tree untouched so base-release merges never conflict on them, but they are **disabled in the repository's Actions settings**: they are wired for the upstream project: they push images to `quay.io/theauthgear` under upstream's credentials, and their release paths are gated on `github.repository == 'authgear/authgear-server'`, so here they would either fail or do nothing useful. Their build jobs run on ordinary `ubuntu-24.04` runners, so runner size is not the reason. To turn one back on, e.g. after adding Go changes:

```bash
gh workflow enable "CI - Branches" --repo authany/authany-server
```

If a base release adds a new workflow file, disable it the same way (`gh workflow disable <name>`) after the merge.

## License

Apache License 2.0, unchanged from upstream: see [`LICENSE.txt`](../LICENSE.txt).

Authany Server is a modified version of the [Authgear server](https://github.com/authgear/authgear-server), copyright Oursky Limited, used under the Apache License 2.0. Upstream ships no `NOTICE` file, so there is none to reproduce here. Every file Authany has changed or added is listed under "Branches and releases" above and carries a commit whose subject ends with `(Authany)`; upstream's own [`README.md`](../README.md) and [`CONTRIBUTING.md`](../CONTRIBUTING.md) are kept for reference, each with a notice at the top marking it as upstream's and otherwise unchanged.

"Authgear" is a trademark of Oursky Limited. It is used here only to describe where this software comes from. Authany is not affiliated with or endorsed by Oursky.
