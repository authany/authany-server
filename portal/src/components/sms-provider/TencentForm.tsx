import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";
import { SMSTemplateCodes, TemplateCodeFields } from "./TemplateCodeFields";

const MASK = "********";

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
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, tencentRegion: value }));
    },
    [setState]
  );

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
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.region" />
        }
        hint={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.tencent.region.hint" />
        }
        optional={true}
        value={state.tencentRegion}
        onChange={onRegionChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="region"
      />
      <TemplateCodeFields
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
