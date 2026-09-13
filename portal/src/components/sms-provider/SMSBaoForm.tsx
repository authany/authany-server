import React, { useCallback } from "react";
import { Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";

const MASK = "********";

export interface SMSBaoFormState {
  smsbaoUsername: string;
  smsbaoPasswordOrAPIKey: string | null;
  smsbaoGoodsID: string;
}

export interface SMSBaoFormProps<T extends SMSBaoFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function SMSBaoForm<T extends SMSBaoFormState>({
  state,
  setState,
}: SMSBaoFormProps<T>): React.ReactElement {
  const isSecretMasked = state.smsbaoPasswordOrAPIKey == null;

  const onUsernameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, smsbaoUsername: value }));
    },
    [setState]
  );

  const onPasswordOrAPIKeyChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, smsbaoPasswordOrAPIKey: value }));
    },
    [setState]
  );

  const onGoodsIDChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, smsbaoGoodsID: value }));
    },
    [setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <Text as="p" size="1" color="gray">
        <FormattedMessage id="SMSProviderConfigurationScreen.form.smsbao.description" />
      </Text>
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsbao.username" />
        }
        required={true}
        value={state.smsbaoUsername}
        onChange={onUsernameChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="username"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsbao.passwordOrAPIKey" />
        }
        hint={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsbao.passwordOrAPIKey.hint" />
        }
        required={true}
        value={state.smsbaoPasswordOrAPIKey ?? MASK}
        onChange={onPasswordOrAPIKeyChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="password_or_api_key"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsbao.goodsID" />
        }
        hint={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.smsbao.goodsID.hint" />
        }
        optional={true}
        value={state.smsbaoGoodsID}
        onChange={onGoodsIDChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="goods_id"
      />
    </div>
  );
}
