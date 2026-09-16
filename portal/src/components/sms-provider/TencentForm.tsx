import React, { useCallback, useContext, useMemo } from "react";
import { Select } from "@radix-ui/themes";
import { Context as MFContext, FormattedMessage } from "../../intl";
import { useFormField } from "../../form";
import { FormField } from "../v2/FormField/FormField";
import { TextField } from "../v2/TextField/TextField";
import { SMSTemplateCodes, TemplateCodeFields } from "./TemplateCodeFields";

const MASK = "********";

// The Tencent Cloud SMS API only serves these three regions.
// An empty stored value means the backend default, which is ap-guangzhou.
const DEFAULT_REGION = "ap-guangzhou";

const REGIONS: { value: string; messageID: string }[] = [
  {
    value: DEFAULT_REGION,
    messageID: "SMSProviderConfigurationScreen.form.tencent.region.guangzhou",
  },
  {
    value: "ap-beijing",
    messageID: "SMSProviderConfigurationScreen.form.tencent.region.beijing",
  },
  {
    value: "ap-nanjing",
    messageID: "SMSProviderConfigurationScreen.form.tencent.region.nanjing",
  },
];

export interface TencentFormState {
  tencentSecretID: string;
  tencentSecretKey: string | null;
  tencentSDKAppID: string;
  tencentRegion: string;
  tencentSignName: string;
  tencentTemplateCode: string;
  tencentTemplateCodes: SMSTemplateCodes;
}

export interface TencentFormProps<T extends TencentFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function TencentForm<T extends TencentFormState>({
  state,
  setState,
}: TencentFormProps<T>): React.ReactElement {
  const isSecretMasked = state.tencentSecretKey == null;
  const { renderToString } = useContext(MFContext);
  // The other fields are TextField, which disables itself while the form is
  // saving. Select does not, so read the same flag here.
  const { loading: isFormLoading } = useFormField(undefined);

  const onSecretIDChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, tencentSecretID: value }));
    },
    [setState]
  );

  const onSecretKeyChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, tencentSecretKey: value }));
    },
    [setState]
  );

  const onSDKAppIDChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, tencentSDKAppID: value }));
    },
    [setState]
  );

  const onRegionChange = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, tencentRegion: value }));
    },
    [setState]
  );

  // An unknown stored region is kept as an extra option so that saving the
  // form does not silently rewrite it.
  const regionOptions = useMemo(() => {
    const options = REGIONS.map((region) => ({
      value: region.value,
      label: renderToString(region.messageID),
    }));
    if (
      state.tencentRegion !== "" &&
      !REGIONS.some((region) => region.value === state.tencentRegion)
    ) {
      options.push({
        value: state.tencentRegion,
        label: state.tencentRegion,
      });
    }
    return options;
  }, [renderToString, state.tencentRegion]);

  const onChangeSignName = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, tencentSignName: value }));
    },
    [setState]
  );

  const onChangeTemplateCode = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, tencentTemplateCode: value }));
    },
    [setState]
  );

  const onChangeTemplateCodes = useCallback(
    (value: SMSTemplateCodes) => {
      setState((prev) => ({ ...prev, tencentTemplateCodes: value }));
    },
    [setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.secretID" />
        }
        required={true}
        value={state.tencentSecretID}
        onChange={onSecretIDChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="secret_id"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.secretKey" />
        }
        required={true}
        value={state.tencentSecretKey ?? MASK}
        onChange={onSecretKeyChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="secret_key"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.sdkAppID" />
        }
        required={true}
        value={state.tencentSDKAppID}
        onChange={onSDKAppIDChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="sdk_app_id"
      />
      <FormField
        size="2"
        labelSize="2"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.region" />
        }
        hint={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.region.hint" />
        }
        labelSpace="1"
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="region"
      >
        <Select.Root
          value={
            state.tencentRegion === "" ? DEFAULT_REGION : state.tencentRegion
          }
          onValueChange={onRegionChange}
          disabled={isSecretMasked || isFormLoading}
        >
          <Select.Trigger variant="surface" />
          <Select.Content>
            {regionOptions.map((option) => (
              <Select.Item key={option.value} value={option.value}>
                {option.label}
              </Select.Item>
            ))}
          </Select.Content>
        </Select.Root>
      </FormField>
      <TemplateCodeFields
        term="templateID"
        signName={state.tencentSignName}
        templateCode={state.tencentTemplateCode}
        templateCodes={state.tencentTemplateCodes}
        disabled={isSecretMasked}
        onChangeSignName={onChangeSignName}
        onChangeTemplateCode={onChangeTemplateCode}
        onChangeTemplateCodes={onChangeTemplateCodes}
      />
    </div>
  );
}
