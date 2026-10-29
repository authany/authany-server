import React, { useCallback, useMemo, useRef, useState } from "react";
import cn from "classnames";
import { useParams } from "react-router-dom";
import { Heading, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import Link from "../../Link";
import ScreenContent from "../../ScreenContent";
import ShowError from "../../ShowError";
import ShowLoading from "../../ShowLoading";
import { FormContainerBase } from "../../FormContainerBase";
import { useFormWithExternalInitialState } from "../../hook/useFormWithExternalInitialState";
import { makeReasonErrorParseRule } from "../../error/parse";
import { TextField } from "../../components/v2/TextField/TextField";
import { CopyIconButton } from "../../components/v2/CopyIconButton/CopyIconButton";
import { SaveFunctionBar } from "../../components/v2/SaveFunctionBar/SaveFunctionBar";
import { SettingsSectionCard } from "../../components/v2/SettingsSectionCard/SettingsSectionCard";
import { useAppAndSecretConfigQuery } from "../../graphql/portal/query/appAndSecretConfigQuery";
import { useProjectNameQueryQuery } from "../../graphql/portal/query/projectNameQuery.generated";
import { useUpdateProjectNameMutationMutation } from "../../graphql/portal/mutations/updateProjectNameMutation.generated";
import { usePortalClient } from "../../graphql/portal/apollo";
import { AppListQueryDocument } from "../../graphql/portal/query/appListQuery.generated";
import {
  PROJECT_NAME_MAX_LENGTH,
  isProjectNameTooLong,
} from "../../util/projectname";
import styles from "./ProjectSettingsScreen.module.css";

const PROJECT_NAME_ERROR_RULES = [
  makeReasonErrorParseRule(
    "InvalidProjectName",
    "ProjectSettingsScreen.project-name.error.invalid"
  ),
];

interface FormState {
  projectName: string;
}

interface ProjectSettingsContentProps {
  appID: string;
  rawAppID: string;
  publicOrigin: string;
  projectName: string;
}

function ProjectSettingsContent(
  props: ProjectSettingsContentProps
): React.ReactElement {
  const { appID, rawAppID, publicOrigin, projectName } = props;
  const contentWidthAnchorRef = useRef<HTMLDivElement>(null);
  const client = usePortalClient();
  const [updateProjectName] = useUpdateProjectNameMutationMutation({
    client,
    // The project list is not normalized by id, so refresh it for the
    // project selector and the projects page.
    refetchQueries: [{ query: AppListQueryDocument }],
  });

  const [initialState] = useState<FormState>({ projectName });

  const form = useFormWithExternalInitialState<FormState, null>({
    defaultState: initialState,
    submit: async (s) => {
      const state = { projectName: s.projectName.trim() };
      await updateProjectName({
        variables: { appID, projectName: state.projectName },
      });
      return { result: null, nextInitialState: state };
    },
  });

  const isDirty = useMemo(() => form.getIsDirty(), [form]);
  const tooLong = isProjectNameTooLong(form.state.projectName);

  const onChangeProjectName = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      form.setState((prev) => ({ ...prev, projectName: value }));
    },
    [form]
  );

  return (
    <FormContainerBase form={form} canSave={!tooLong}>
      <ScreenContent className={cn(isDirty ? styles.contentWithSaveBar : null)}>
        <div
          ref={contentWidthAnchorRef}
          className={cn(styles.widget, styles.pageHeader)}
        >
          <Heading as="h1" size="5" weight="bold" className={styles.pageTitle}>
            <FormattedMessage id="ProjectSettingsScreen.title" />
          </Heading>
          <Text as="p" size="2" color="gray" className={styles.pageDescription}>
            <FormattedMessage id="ProjectSettingsScreen.description" />
          </Text>
        </div>

        <SettingsSectionCard
          className={styles.widget}
          title={
            <FormattedMessage id="ProjectSettingsScreen.project-name.title" />
          }
          description={
            <FormattedMessage
              id="ProjectSettingsScreen.project-name.description"
              values={{
                // eslint-disable-next-line react/no-unstable-nested-components
                designLink: (chunks: React.ReactNode) => (
                  <Link to={`/project/${appID}/branding/design`}>{chunks}</Link>
                ),
              }}
            />
          }
        >
          <TextField
            size="2"
            labelSize="2"
            label={
              <FormattedMessage id="ProjectSettingsScreen.project-name.label" />
            }
            placeholder={rawAppID}
            value={form.state.projectName}
            onChange={onChangeProjectName}
            fieldName="projectName"
            errorRules={PROJECT_NAME_ERROR_RULES}
            error={
              tooLong ? (
                <FormattedMessage
                  id="ProjectSettingsScreen.project-name.error.too-long"
                  values={{ max: PROJECT_NAME_MAX_LENGTH }}
                />
              ) : undefined
            }
            hint={
              <FormattedMessage id="ProjectSettingsScreen.project-name.hint" />
            }
          />
        </SettingsSectionCard>

        <SettingsSectionCard
          className={cn(
            styles.widget,
            isDirty && styles.settingsCardSaveBarClearance
          )}
          contentClassName="gap-6"
          title={<FormattedMessage id="ProjectSettingsScreen.details.title" />}
        >
          <div className={styles.readOnlyField}>
            <Text as="p" size="2" weight="medium">
              <FormattedMessage id="ProjectSettingsScreen.project-id.label" />
            </Text>
            <div className={styles.readOnlyValueRow}>
              <code className={styles.readOnlyValue}>{rawAppID}</code>
              <CopyIconButton textToCopy={rawAppID} />
            </div>
            <Text as="p" size="1" color="gray">
              <FormattedMessage id="ProjectSettingsScreen.project-id.hint" />
            </Text>
          </div>
          <div className={styles.readOnlyField}>
            <Text as="p" size="2" weight="medium">
              <FormattedMessage id="ProjectSettingsScreen.public-origin.label" />
            </Text>
            <div className={styles.readOnlyValueRow}>
              <code className={styles.readOnlyValue}>{publicOrigin}</code>
              <CopyIconButton textToCopy={publicOrigin} />
            </div>
            <Text as="p" size="1" color="gray">
              <FormattedMessage
                id="ProjectSettingsScreen.public-origin.hint"
                values={{
                  // eslint-disable-next-line react/no-unstable-nested-components
                  domainsLink: (chunks: React.ReactNode) => (
                    <Link to={`/project/${appID}/branding/custom-domains`}>
                      {chunks}
                    </Link>
                  ),
                }}
              />
            </Text>
          </div>
        </SettingsSectionCard>

        <SaveFunctionBar anchorRef={contentWidthAnchorRef} />
      </ScreenContent>
    </FormContainerBase>
  );
}

const ProjectSettingsScreen: React.VFC = function ProjectSettingsScreen() {
  const { appID } = useParams() as { appID: string };
  const client = usePortalClient();
  const projectNameQuery = useProjectNameQueryQuery({
    client,
    variables: { id: appID },
  });
  const appConfigQuery = useAppAndSecretConfigQuery(appID);

  if (projectNameQuery.loading || appConfigQuery.isLoading) {
    return <ShowLoading />;
  }
  if (projectNameQuery.error != null) {
    return (
      <ShowError
        error={projectNameQuery.error}
        onRetry={() => {
          projectNameQuery.refetch().catch((e: unknown) => {
            console.error(e);
          });
        }}
      />
    );
  }
  if (appConfigQuery.loadError != null) {
    return (
      <ShowError
        error={appConfigQuery.loadError}
        // eslint-disable-next-line @typescript-eslint/strict-void-return
        onRetry={appConfigQuery.refetch}
      />
    );
  }

  const node = projectNameQuery.data?.node;
  const projectName = node?.__typename === "App" ? node.projectName : null;
  const config = appConfigQuery.effectiveAppConfig;

  return (
    <ProjectSettingsContent
      appID={appID}
      rawAppID={config?.id ?? ""}
      publicOrigin={config?.http?.public_origin ?? ""}
      projectName={projectName ?? ""}
    />
  );
};

export default ProjectSettingsScreen;
