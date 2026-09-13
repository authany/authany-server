import React, { useCallback } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";

const MASK = "********";

export interface YunpianFormState {
  yunpianAPIKey: string | null;
}

export interface YunpianFormProps<T extends YunpianFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function YunpianForm<T extends YunpianFormState>({
  state,
  setState,
}: YunpianFormProps<T>): React.ReactElement {
  const isSecretMasked = state.yunpianAPIKey == null;

  const onAPIKeyChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, yunpianAPIKey: value }));
    },
    [setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <Text as="p" size="1" color="gray">
        <FormattedMessage id="SMSProviderConfigurationScreen.form.yunpian.description" />
      </Text>
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.yunpian.apiKey" />
        }
        required={true}
        value={state.yunpianAPIKey ?? MASK}
        onChange={onAPIKeyChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="apikey"
      />
    </div>
  );
}
