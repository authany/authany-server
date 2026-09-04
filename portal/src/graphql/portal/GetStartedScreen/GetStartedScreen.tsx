import React, { useEffect, useMemo } from "react";
import { useParams } from "react-router-dom";
import { FormattedMessage } from "../../../intl";
import { Text } from "@fluentui/react";
import ShowLoading from "../../../ShowLoading";
import ShowError from "../../../ShowError";
import ScreenLayoutScrollView from "../../../ScreenLayoutScrollView";
import { useCapture } from "../../../gtm_v2";
import { useAppAndSecretConfigQuery } from "../query/appAndSecretConfigQuery";

import HeroLoginCard from "./HeroLoginCard";
import HeroIntegrateCard from "./HeroIntegrateCard";
import FeatureCard from "./FeatureCard";

import {
  EyeOpenIcon,
  GearIcon,
  LightningBoltIcon,
  LockClosedIcon,
  PersonIcon,
  RulerSquareIcon,
} from "@radix-ui/react-icons";

import styles from "./GetStartedScreen.module.css";

const DOCS_CUSTOM_UI = "https://docs.authgear.com/customization/custom-ui";

interface GetStartedScreenContentProps {
  publicOrigin: string;
  numberOfClients: number;
}

function GetStartedScreenContent(
  props: GetStartedScreenContentProps
): React.ReactElement {
  const { publicOrigin, numberOfClients } = props;
  const { appID } = useParams() as { appID: string };
  const capture = useCapture();
  const hasApp = numberOfClients > 0;
  const captureData = useMemo(() => ({ has_app: hasApp }), [hasApp]);

  useEffect(() => {
    capture("getStarted.viewed", captureData);
  }, [capture, captureData]);

  const featureCards = useMemo(
    () => [
      {
        Icon: LockClosedIcon,
        titleMessageID: "GetStartedScreen.feature.2fa.title",
        descriptionMessageID: "GetStartedScreen.feature.2fa.description",
        actionMessageID: "GetStartedScreen.feature.2fa.action",
        internalHref: `/project/${appID}/configuration/authentication/2fa`,
        onClick: () => capture("getStarted.clicked-feature_2fa", captureData),
      },
      {
        Icon: EyeOpenIcon,
        titleMessageID: "GetStartedScreen.feature.bot-protection.title",
        descriptionMessageID:
          "GetStartedScreen.feature.bot-protection.description",
        actionMessageID: "GetStartedScreen.feature.bot-protection.action",
        internalHref: `/project/${appID}/attack-protection/bot-protection`,
        onClick: () =>
          capture("getStarted.clicked-feature_bot_protection", captureData),
      },
      {
        Icon: PersonIcon,
        titleMessageID: "GetStartedScreen.feature.user-management.title",
        descriptionMessageID:
          "GetStartedScreen.feature.user-management.description",
        actionMessageID: "GetStartedScreen.feature.user-management.action",
        internalHref: `/project/${appID}/users`,
        onClick: () =>
          capture("getStarted.clicked-feature_user_management", captureData),
      },
      {
        Icon: GearIcon,
        titleMessageID: "GetStartedScreen.feature.admin-api.title",
        descriptionMessageID: "GetStartedScreen.feature.admin-api.description",
        actionMessageID: "GetStartedScreen.feature.admin-api.action",
        internalHref: `/project/${appID}/advanced/admin-api`,
        onClick: () =>
          capture("getStarted.clicked-feature_admin_api", captureData),
      },
      {
        Icon: RulerSquareIcon,
        titleMessageID: "GetStartedScreen.feature.custom-ui.title",
        descriptionMessageID: "GetStartedScreen.feature.custom-ui.description",
        actionMessageID: "GetStartedScreen.feature.custom-ui.action",
        externalHref: DOCS_CUSTOM_UI,
        onClick: () =>
          capture("getStarted.clicked-feature_custom_ui", captureData),
      },
      {
        Icon: LightningBoltIcon,
        titleMessageID: "GetStartedScreen.feature.hooks.title",
        descriptionMessageID: "GetStartedScreen.feature.hooks.description",
        actionMessageID: "GetStartedScreen.feature.hooks.action",
        internalHref: `/project/${appID}/advanced/hooks`,
        onClick: () => capture("getStarted.clicked-feature_hooks", captureData),
      },
    ],
    [appID, capture, captureData]
  );

  return (
    <ScreenLayoutScrollView>
      <div className={styles.root}>
        <div className={styles.heroRow}>
          <HeroLoginCard
            appID={appID}
            publicOrigin={publicOrigin}
            hasApp={hasApp}
          />
          <HeroIntegrateCard appID={appID} hasApp={hasApp} />
        </div>

        <section className={styles.featureSection}>
          <Text as="h2" block={true} className={styles.sectionHeading}>
            <FormattedMessage id="GetStartedScreen.features.heading" />
          </Text>
          <div className={styles.featureGrid}>
            {featureCards.map((card) => (
              <FeatureCard key={card.titleMessageID} {...card} />
            ))}
          </div>
        </section>
      </div>
    </ScreenLayoutScrollView>
  );
}

export default function GetStartedScreen(): React.ReactElement {
  const { appID } = useParams() as { appID: string };

  const { effectiveAppConfig, isLoading, loadError, refetch } =
    useAppAndSecretConfigQuery(appID);

  if (isLoading || !effectiveAppConfig) {
    return <ShowLoading />;
  }

  if (loadError) {
    // eslint-disable-next-line @typescript-eslint/strict-void-return
    return <ShowError error={loadError} onRetry={refetch} />;
  }

  return (
    <GetStartedScreenContent
      publicOrigin={effectiveAppConfig.http?.public_origin ?? ""}
      numberOfClients={effectiveAppConfig.oauth?.clients?.length ?? 0}
    />
  );
}
