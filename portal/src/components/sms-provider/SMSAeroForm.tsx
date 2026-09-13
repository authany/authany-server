import React, { useCallback } from "react";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";

const MASK = "********";

export interface SMSAeroFormState {
  smsAeroEmail: string;
  smsAeroAPIKey: string | null;
  smsAeroSenderName: string;
}

export interface SMSAeroFormProps<T extends SMSAeroFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function SMSAeroForm<T extends SMSAeroFormState>({
  state,
  setState,
}: SMSAeroFormProps<T>): React.ReactElement {
  const isSecretMasked = state.smsAeroAPIKey == null;

  const onEmailChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, smsAeroEmail: value }));
    },
    [setState]
  );

  const onAPIKeyChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, smsAeroAPIKey: value }));
    },
    [setState]
  );

  const onSenderNameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, smsAeroSenderName: value }));
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
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsaero.email" />
        }
        required={true}
        value={state.smsAeroEmail}
        onChange={onEmailChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="email"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsaero.apiKey" />
        }
        required={true}
        value={state.smsAeroAPIKey ?? MASK}
        onChange={onAPIKeyChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="api_key"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsaero.senderName" />
        }
        hint={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsaero.senderName.hint" />
        }
        required={true}
        value={state.smsAeroSenderName}
        onChange={onSenderNameChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="sender_name"
      />
    </div>
  );
}
