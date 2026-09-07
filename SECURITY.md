# Security Policy

## Reporting a vulnerability

Report security issues in **this repository** privately, through GitHub's
[private vulnerability reporting](https://github.com/authany/authany-server/security/advisories/new)
on the Security tab. If that is unavailable to you, email **hello@authany.com** with `SECURITY` in
the subject line.

Please include enough detail to reproduce: affected version or commit, the request or configuration
involved, and what an attacker gains. Please do not open a public issue or pull request for a
security problem, and please give us a chance to ship a fix before disclosing it publicly.

We will acknowledge your report and keep you informed as we work on it. We are a small team, so we
cannot promise a fixed response time.

## Scope

This policy covers the code in this repository and the Authany service operated at `authany.com`.

Authany Server is derived from [Authgear](https://github.com/authgear/authgear-server) by Oursky
Limited. If you find a vulnerability that is inherited from Authgear rather than introduced by
Authany, it affects Authgear's own users too, and the right place to report it is
[Authgear's security policy](https://github.com/authgear/authgear-server/security). Telling us as
well is welcome, and we will coordinate rather than duplicate.

## Out of scope

Reports generated purely by automated scanners with no demonstrated impact, findings against
third-party services we do not operate, and issues that require an already-compromised account or
host.
