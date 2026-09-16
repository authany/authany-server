import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";
import { SMSTemplateCodes, TemplateCodeFields } from "./TemplateCodeFields";

const MASK = "********";

export interface AliyunFormState {
  aliyunAccessKeyID: string;
  aliyunAccessKeySecret: string | null;
  aliyunSignName: string;
  aliyunTemplateCode: string;
  aliyunTemplateCodes: SMSTemplateCodes;
  aliyunOverseasTemplateCode: string;
}

export interface AliyunFormProps<T extends AliyunFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function AliyunForm<T extends AliyunFormState>({
  state,
  setState,
}: AliyunFormProps<T>): React.ReactElement {
  const isSecretMasked = state.aliyunAccessKeySecret == null;

  const onAccessKeyIDChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, aliyunAccessKeyID: value }));
    },
    [setState]
  );

  const onAccessKeySecretChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, aliyunAccessKeySecret: value }));
    },
    [setState]
  );

  const onChangeSignName = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, aliyunSignName: value }));
    },
    [setState]
  );

  const onChangeTemplateCode = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, aliyunTemplateCode: value }));
    },
    [setState]
  );

  const onChangeTemplateCodes = useCallback(
    (value: SMSTemplateCodes) => {
      setState((prev) => ({ ...prev, aliyunTemplateCodes: value }));
    },
    [setState]
  );

  const onChangeOverseasTemplateCode = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, aliyunOverseasTemplateCode: value }));
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
          <FormattedMessage id="SMSProviderConfigurationScreen.form.aliyun.accessKeyID" />
        }
        required={true}
        value={state.aliyunAccessKeyID}
        onChange={onAccessKeyIDChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="access_key_id"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.aliyun.accessKeySecret" />
        }
        required={true}
        value={state.aliyunAccessKeySecret ?? MASK}
        onChange={onAccessKeySecretChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="access_key_secret"
      />
      <TemplateCodeFields
        term="templateCode"
        signName={state.aliyunSignName}
        templateCode={state.aliyunTemplateCode}
        templateCodes={state.aliyunTemplateCodes}
        overseasTemplateCode={state.aliyunOverseasTemplateCode}
        disabled={isSecretMasked}
        onChangeSignName={onChangeSignName}
        onChangeTemplateCode={onChangeTemplateCode}
        onChangeTemplateCodes={onChangeTemplateCodes}
        onChangeOverseasTemplateCode={onChangeOverseasTemplateCode}
      />
    </div>
  );
}
