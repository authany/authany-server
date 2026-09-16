import React, { useCallback } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";
import { SMSTemplateCodes, TemplateCodeFields } from "./TemplateCodeFields";

const MASK = "********";

export interface AliyunMASFormState {
  aliyunMASAccessKeyID: string;
  aliyunMASAccessKeySecret: string | null;
  aliyunMASSignName: string;
  aliyunMASTemplateCode: string;
  aliyunMASTemplateCodes: SMSTemplateCodes;
}

export interface AliyunMASFormProps<T extends AliyunMASFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function AliyunMASForm<T extends AliyunMASFormState>({
  state,
  setState,
}: AliyunMASFormProps<T>): React.ReactElement {
  const isSecretMasked = state.aliyunMASAccessKeySecret == null;

  const onAccessKeyIDChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, aliyunMASAccessKeyID: value }));
    },
    [setState]
  );

  const onAccessKeySecretChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, aliyunMASAccessKeySecret: value }));
    },
    [setState]
  );

  const onChangeSignName = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, aliyunMASSignName: value }));
    },
    [setState]
  );

  const onChangeTemplateCode = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, aliyunMASTemplateCode: value }));
    },
    [setState]
  );

  const onChangeTemplateCodes = useCallback(
    (value: SMSTemplateCodes) => {
      setState((prev) => ({ ...prev, aliyunMASTemplateCodes: value }));
    },
    [setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <Text as="p" size="1" color="gray">
        <FormattedMessage id="SMSProviderConfigurationScreen.form.aliyunMAS.description" />
      </Text>
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.aliyunMAS.accessKeyID" />
        }
        required={true}
        value={state.aliyunMASAccessKeyID}
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
          <FormattedMessage id="SMSProviderConfigurationScreen.form.aliyunMAS.accessKeySecret" />
        }
        required={true}
        value={state.aliyunMASAccessKeySecret ?? MASK}
        onChange={onAccessKeySecretChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="access_key_secret"
      />
      <TemplateCodeFields
        term="templateCode"
        signName={state.aliyunMASSignName}
        templateCode={state.aliyunMASTemplateCode}
        templateCodes={state.aliyunMASTemplateCodes}
        disabled={isSecretMasked}
        onChangeSignName={onChangeSignName}
        onChangeTemplateCode={onChangeTemplateCode}
        onChangeTemplateCodes={onChangeTemplateCodes}
      />
    </div>
  );
}
