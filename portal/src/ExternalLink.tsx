import React, { ReactElement } from "react";
// eslint-disable-next-line no-restricted-imports
import { ILinkProps, Link as FluentLink } from "@fluentui/react";

export interface ExternalLinkProps extends Omit<ILinkProps, "rel"> {}

export const DEFAULT_EXTERNAL_LINK_PROPS = {
  target: "_blank",
  rel: "noreferrer",
};

// Authany: the portal is white-labelled. Links that would lead to the upstream
// vendor's website, docs, community or mailboxes are rendered as plain text.
const UPSTREAM_LINK =
  /^(https?:\/\/([a-z0-9-]+\.)*authgear\.com(\/|$)|https?:\/\/discord\.gg\/|https?:\/\/github\.com\/authgear(\/|$)|mailto:[^@\s]+@authgear\.com)/i;

export function isUpstreamLink(href: string | undefined): boolean {
  return href != null && UPSTREAM_LINK.test(href);
}

export default function ExternalLink(props: ExternalLinkProps): ReactElement {
  if (isUpstreamLink(props.href)) {
    return <span className={props.className}>{props.children}</span>;
  }
  return <FluentLink {...DEFAULT_EXTERNAL_LINK_PROPS} {...props} />;
}
