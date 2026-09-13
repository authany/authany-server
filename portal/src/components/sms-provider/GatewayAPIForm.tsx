import React, { useCallback } from "react";
import { Flex, RadioGroup, Text } from "@radix-ui/themes";
import { FormattedMessage } from "../../intl";
import { TextField } from "../v2/TextField/TextField";
import { FormField } from "../v2/FormField/FormField";

const MASK = "********";

export const GATEWAY_API_ENDPOINTS = [
  "https://gatewayapi.com",
  "https://gatewayapi.eu",
] as const;

export const DEFAULT_GATEWAY_API_ENDPOINT = GATEWAY_API_ENDPOINTS[0];

export interface GatewayAPIFormState {
  gatewayAPIEndpoint: string;
  gatewayAPIAPIToken: string | null;
  gatewayAPISender: string;
}

export interface GatewayAPIFormProps<T extends GatewayAPIFormState> {
  state: T;
  setState: (fn: (prev: T) => T) => void;
}

export function GatewayAPIForm<T extends GatewayAPIFormState>({
  state,
  setState,
}: GatewayAPIFormProps<T>): React.ReactElement {
  const isSecretMasked = state.gatewayAPIAPIToken == null;

  const onEndpointChange = useCallback(
    (value: string) => {
      setState((prev) => ({ ...prev, gatewayAPIEndpoint: value }));
    },
    [setState]
  );

  const onAPITokenChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, gatewayAPIAPIToken: value }));
    },
    [setState]
  );

  const onSenderChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setState((prev) => ({ ...prev, gatewayAPISender: value }));
    },
    [setState]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <FormField
        size="2"
        labelSize="2"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.gatewayapi.endpoint" />
        }
        required={true}
        labelSpace="1"
      >
        <RadioGroup.Root
          value={state.gatewayAPIEndpoint}
          onValueChange={onEndpointChange}
          disabled={isSecretMasked}
        >
          <Flex direction="column" gap="2">
            {GATEWAY_API_ENDPOINTS.map((endpoint) => (
              <Text as="label" size="2" key={endpoint}>
                <Flex gap="2" align="center">
                  <RadioGroup.Item value={endpoint} />
                  {endpoint}
                </Flex>
              </Text>
            ))}
          </Flex>
        </RadioGroup.Root>
      </FormField>
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.gatewayapi.apiToken" />
        }
        required={true}
        value={state.gatewayAPIAPIToken ?? MASK}
        onChange={onAPITokenChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="api_token"
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.gatewayapi.sender" />
        }
        hint={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.gatewayapi.sender.hint" />
        }
        required={true}
        value={state.gatewayAPISender}
        onChange={onSenderChange}
        disabled={isSecretMasked}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="sender"
      />
    </div>
  );
}
