import React, { ReactElement } from "react";
// eslint-disable-next-line no-restricted-imports
import { ILinkProps, Link as FluentLink } from "@fluentui/react";
import { AUTHANY_DOCS_ORIGIN, UPSTREAM_DOCS_PATHS } from "./authanyDocs";

export interface ExternalLinkProps extends Omit<ILinkProps, "rel"> {}

export const DEFAULT_EXTERNAL_LINK_PROPS = {
  target: "_blank",
  rel: "noreferrer",
};

// Authany: the portal is white-labelled. Links that would lead to the upstream
// vendor's website, docs, community or mailboxes are rendered as plain text.
const UPSTREAM_LINK =
  /^(https?:\/\/([a-z0-9-]+\.)*authgear\.com(\/|$)|https?:\/\/discord\.gg\/|https?:\/\/github\.com\/authgear(\/|$)|mailto:[^@\s]+@authgear\.com)/i;

// Authany: upstream documentation links are re-pointed at docs.authany.com when
// the page exists there (map generated from the docs site build), otherwise
// they fall back to plain text like any other vendor link.
const UPSTREAM_DOCS = /^https?:\/\/docs\.authgear\.com(\/[^#?]*)?([#?].*)?$/i;

export function isUpstreamLink(href: string | undefined): boolean {
  return href != null && UPSTREAM_LINK.test(href);
}

export function mapUpstreamDocsLink(
  href: string | undefined
): string | undefined {
  if (href == null) {
    return undefined;
  }
  const m = UPSTREAM_DOCS.exec(href);
  if (m == null) {
    return undefined;
  }
  const key = (optionalGroup(m, 1) ?? "").replace(/^\/+|\/+$/g, "");
  const target = UPSTREAM_DOCS_PATHS[key];
  if (target == null) {
    return undefined;
  }
  return AUTHANY_DOCS_ORIGIN + target + (optionalGroup(m, 2) ?? "");
}

// Capture groups that did not participate in the match are undefined at
// runtime even though RegExpExecArray types them as string.
function optionalGroup(m: RegExpExecArray, index: number): string | undefined {
  return m[index];
}

export default function ExternalLink(props: ExternalLinkProps): ReactElement {
  const docsHref = mapUpstreamDocsLink(props.href);
  if (docsHref != null) {
    return (
      <FluentLink {...DEFAULT_EXTERNAL_LINK_PROPS} {...props} href={docsHref} />
    );
  }
  if (isUpstreamLink(props.href)) {
    return <span className={props.className}>{props.children}</span>;
  }
  return <FluentLink {...DEFAULT_EXTERNAL_LINK_PROPS} {...props} />;
}
