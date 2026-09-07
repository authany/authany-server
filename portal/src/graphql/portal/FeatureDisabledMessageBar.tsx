import React, { useMemo } from "react";
import { IMessageBarProps } from "@fluentui/react";
import { FormattedMessage, Values } from "../../intl";
import BlueMessageBar from "../../BlueMessageBar";
import ExternalLink from "../../ExternalLink";

export interface FeatureDisabledMessageBarProps extends IMessageBarProps {
  messageID: string;
  messageValues?: Values;
}

// useFeatureDisabledMessageValues provides the standard rich-text values
// (plan-page link, contact-us link, bold) for FeatureConfig.*.disabled
// messages. Shared by the Fluent message bar below and the v2
// FeatureDisabledCallout.
export function useFeatureDisabledMessageValues(
  messageValues?: Values
): Values {
  return useMemo(() => {
    const planPagePath = "mailto:hello@authany.com"; // Authany: no billing page; every upgrade link becomes a contact link
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

const FeatureDisabledMessageBar: React.VFC<FeatureDisabledMessageBarProps> =
  function FeatureDisabledMessageBar(props: FeatureDisabledMessageBarProps) {
    const { messageID, messageValues, ...rest } = props;
    const values = useFeatureDisabledMessageValues(messageValues);

    return (
      <BlueMessageBar {...rest}>
        <FormattedMessage id={messageID} values={values} />
      </BlueMessageBar>
    );
  };

export default FeatureDisabledMessageBar;
