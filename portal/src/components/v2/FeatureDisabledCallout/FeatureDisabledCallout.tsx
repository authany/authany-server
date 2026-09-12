import React, { useMemo } from "react";
import { Callout } from "@radix-ui/themes";
import { InfoCircledIcon } from "@radix-ui/react-icons";
import { FormattedMessage, Values } from "../../../intl";
import ExternalLink from "../../../ExternalLink";

export interface FeatureDisabledCalloutProps {
  className?: string;
  messageID: string;
  messageValues?: Values;
}

// useFeatureDisabledMessageValues provides the standard rich-text values
// (plan-page link, contact-us link, bold) for FeatureConfig.*.disabled
// messages.
export function useFeatureDisabledMessageValues(
  messageValues?: Values
): Values {
  return useMemo(() => {
    // Authany: no billing page; every upgrade link becomes a contact link.
    const planPagePath = "mailto:hello@authany.com";
    const contactUsHref = "mailto:hello@authany.com";
    return {
      planPagePath,
      contactUsHref,

      b: (chunks: React.ReactNode) => <b>{chunks}</b>,

      ReactRouterLink: (chunks: React.ReactNode) => (
        <ExternalLink href={contactUsHref}>{chunks}</ExternalLink>
      ),

      ExternalLink: (chunks: React.ReactNode) => (
        <ExternalLink href={contactUsHref}>{chunks}</ExternalLink>
      ),
      ...messageValues,
    };
  }, [messageValues]);
}

// The standard FeatureConfig.*.disabled message (with plan-page /
// contact-us links) rendered as a Radix Callout, matching the info
// callouts used across the migrated screens.
export function FeatureDisabledCallout({
  className,
  messageID,
  messageValues,
}: FeatureDisabledCalloutProps): React.ReactElement {
  const values = useFeatureDisabledMessageValues(messageValues);

  return (
    <Callout.Root className={className} color="blue" variant="surface" size="1">
      <Callout.Icon>
        <InfoCircledIcon />
      </Callout.Icon>
      <Callout.Text>
        <FormattedMessage id={messageID} values={values} />
      </Callout.Text>
    </Callout.Root>
  );
}
