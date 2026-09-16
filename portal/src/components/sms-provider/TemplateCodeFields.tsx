import React, { useCallback, useContext, useMemo, useState } from "react";
import { Text } from "@radix-ui/themes";
import { Context as MFContext, FormattedMessage } from "../../intl";
import { Accordion } from "../common/Accordion";
import { TextField } from "../v2/TextField/TextField";
import { ErrorParseRule, makeLocalErrorParseRule } from "../../error/parse";
import { LocalError } from "../../error/error";

export const localErrorSignNameRequired: LocalError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.validation.required",
    },
  },
};

const signNameErrorRules: ErrorParseRule[] = [
  makeLocalErrorParseRule(
    localErrorSignNameRequired,
    localErrorSignNameRequired.info.error
  ),
];

export const localErrorTemplateCodeRequired: LocalError = {
  errorName: "__local",
  reason: "__local",
  info: {
    error: {
      messageID: "errors.validation.required",
    },
  },
};

const templateCodeErrorRules: ErrorParseRule[] = [
  makeLocalErrorParseRule(
    localErrorTemplateCodeRequired,
    localErrorTemplateCodeRequired.info.error
  ),
];

export const SMS_TEMPLATE_NAMES = [
  "verification_sms.txt",
  "setup_primary_oob_sms.txt",
  "setup_secondary_oob_sms.txt",
  "authenticate_primary_oob_sms.txt",
  "authenticate_secondary_oob_sms.txt",
  "forgot_password_oob_sms.txt",
] as const;

export type SMSTemplateName = (typeof SMS_TEMPLATE_NAMES)[number];

export type SMSTemplateCodes = Partial<Record<SMSTemplateName, string>>;

/**
 * The provider-specific noun for the template identifier, so that the labels
 * match the wording of the provider console. Tencent Cloud calls it a template
 * ID while Alibaba Cloud calls it a template code.
 */
export type SMSTemplateTerm = "templateID" | "templateCode";

const templateTermMessageIDs: Record<SMSTemplateTerm, string> = {
  templateID:
    "SMSProviderConfigurationScreen.form.templateCodes.term.templateID",
  templateCode:
    "SMSProviderConfigurationScreen.form.templateCodes.term.templateCode",
};

const templateNameMessageIDs: Record<SMSTemplateName, string> = {
  "verification_sms.txt":
    "SMSProviderConfigurationScreen.form.templateCodes.template.verification",
  "setup_primary_oob_sms.txt":
    "SMSProviderConfigurationScreen.form.templateCodes.template.setupPrimaryOOB",
  "setup_secondary_oob_sms.txt":
    "SMSProviderConfigurationScreen.form.templateCodes.template.setupSecondaryOOB",
  "authenticate_primary_oob_sms.txt":
    "SMSProviderConfigurationScreen.form.templateCodes.template.authenticatePrimaryOOB",
  "authenticate_secondary_oob_sms.txt":
    "SMSProviderConfigurationScreen.form.templateCodes.template.authenticateSecondaryOOB",
  "forgot_password_oob_sms.txt":
    "SMSProviderConfigurationScreen.form.templateCodes.template.forgotPassword",
};

export function parseSMSTemplateCodes(raw: unknown): SMSTemplateCodes {
  const out: SMSTemplateCodes = {};
  if (raw == null || typeof raw !== "object") {
    return out;
  }
  const record = raw as Record<string, unknown>;
  for (const templateName of SMS_TEMPLATE_NAMES) {
    const value = record[templateName];
    if (typeof value === "string" && value !== "") {
      out[templateName] = value;
    }
  }
  return out;
}

export function serializeSMSTemplateCodes(
  codes: SMSTemplateCodes
): Record<string, string> {
  const out: Record<string, string> = {};
  for (const templateName of SMS_TEMPLATE_NAMES) {
    const value = codes[templateName];
    if (value != null && value !== "") {
      out[templateName] = value;
    }
  }
  return out;
}

function TemplateCodeOverrideField({
  templateName,
  value,
  placeholder,
  disabled,
  onChange,
}: {
  templateName: SMSTemplateName;
  value: string;
  placeholder: string;
  disabled?: boolean;
  onChange: (templateName: SMSTemplateName, value: string) => void;
}): React.ReactElement {
  const onChangeValue = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onChange(templateName, e.target.value);
    },
    [onChange, templateName]
  );

  return (
    <TextField
      size="2"
      labelSize="2"
      type="text"
      label={<FormattedMessage id={templateNameMessageIDs[templateName]} />}
      hint={
        <FormattedMessage id={`${templateNameMessageIDs[templateName]}.hint`} />
      }
      optional={true}
      placeholder={placeholder}
      value={value}
      onChange={onChangeValue}
      disabled={disabled}
      parentJSONPointer={/\/secrets\/\d+\/data\/template_codes/}
      fieldName={templateName}
    />
  );
}

export interface TemplateCodeFieldsProps {
  /** The noun this provider uses for the template identifier. */
  term: SMSTemplateTerm;
  signName: string;
  templateCode: string;
  templateCodes: SMSTemplateCodes;
  /** When omitted, the overseas template code field is not rendered. */
  overseasTemplateCode?: string;
  disabled?: boolean;
  onChangeSignName: (value: string) => void;
  onChangeTemplateCode: (value: string) => void;
  onChangeTemplateCodes: (value: SMSTemplateCodes) => void;
  onChangeOverseasTemplateCode?: (value: string) => void;
}

export function TemplateCodeFields({
  term,
  signName,
  templateCode,
  templateCodes,
  overseasTemplateCode,
  disabled,
  onChangeSignName,
  onChangeTemplateCode,
  onChangeTemplateCodes,
  onChangeOverseasTemplateCode,
}: TemplateCodeFieldsProps): React.ReactElement {
  const { renderToString } = useContext(MFContext);
  const termText = renderToString(templateTermMessageIDs[term]);
  const termValues = useMemo(() => ({ term: termText }), [termText]);
  const overridePlaceholder = renderToString(
    "SMSProviderConfigurationScreen.form.templateCodes.overrides.placeholder",
    { term: termText }
  );

  // Only the value on mount matters, so that expanding or collapsing the
  // overrides afterwards is not undone by editing the fields inside.
  const [hasTemplateCodeOverride] = useState(() =>
    SMS_TEMPLATE_NAMES.some((templateName) => {
      const value = templateCodes[templateName];
      return value != null && value !== "";
    })
  );

  const onSignNameChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onChangeSignName(e.target.value);
    },
    [onChangeSignName]
  );

  const onTemplateCodeChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onChangeTemplateCode(e.target.value);
    },
    [onChangeTemplateCode]
  );

  const onOverseasTemplateCodeChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      onChangeOverseasTemplateCode?.(e.target.value);
    },
    [onChangeOverseasTemplateCode]
  );

  const onTemplateCodeOverrideChange = useCallback(
    (templateName: SMSTemplateName, value: string) => {
      const newTemplateCodes: SMSTemplateCodes = { ...templateCodes };
      if (value === "") {
        delete newTemplateCodes[templateName];
      } else {
        newTemplateCodes[templateName] = value;
      }
      onChangeTemplateCodes(newTemplateCodes);
    },
    [onChangeTemplateCodes, templateCodes]
  );

  return (
    <div className="flex flex-col gap-y-4">
      <Text as="p" size="1" color="gray">
        <FormattedMessage id="SMSProviderConfigurationScreen.form.templateCodes.description" />
      </Text>
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage id="SMSProviderConfigurationScreen.form.templateCodes.signName" />
        }
        required={true}
        value={signName}
        onChange={onSignNameChange}
        disabled={disabled}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="sign_name"
        errorRules={signNameErrorRules}
      />
      <TextField
        size="2"
        labelSize="2"
        type="text"
        label={
          <FormattedMessage
            id="SMSProviderConfigurationScreen.form.templateCodes.templateCode"
            values={termValues}
          />
        }
        hint={
          <FormattedMessage
            id="SMSProviderConfigurationScreen.form.templateCodes.templateCode.hint"
            values={termValues}
          />
        }
        required={true}
        value={templateCode}
        onChange={onTemplateCodeChange}
        disabled={disabled}
        parentJSONPointer={/\/secrets\/\d+\/data/}
        fieldName="template_code"
        errorRules={templateCodeErrorRules}
      />
      {overseasTemplateCode !== undefined ? (
        <TextField
          size="2"
          labelSize="2"
          type="text"
          label={
            <FormattedMessage
              id="SMSProviderConfigurationScreen.form.templateCodes.overseasTemplateCode"
              values={termValues}
            />
          }
          hint={
            <FormattedMessage id="SMSProviderConfigurationScreen.form.templateCodes.overseasTemplateCode.hint" />
          }
          optional={true}
          value={overseasTemplateCode}
          onChange={onOverseasTemplateCodeChange}
          disabled={disabled}
          parentJSONPointer={/\/secrets\/\d+\/data/}
          fieldName="overseas_template_code"
        />
      ) : null}
      <Accordion
        text={
          <FormattedMessage
            id="SMSProviderConfigurationScreen.form.templateCodes.overrides.label"
            values={termValues}
          />
        }
        defaultExpanded={hasTemplateCodeOverride}
      >
        <div className="flex flex-col gap-y-4">
          <Text as="p" size="1" color="gray">
            <FormattedMessage
              id="SMSProviderConfigurationScreen.form.templateCodes.overrides.description"
              values={termValues}
            />
          </Text>
          {SMS_TEMPLATE_NAMES.map((templateName) => (
            <TemplateCodeOverrideField
              key={templateName}
              templateName={templateName}
              value={templateCodes[templateName] ?? ""}
              placeholder={overridePlaceholder}
              disabled={disabled}
              onChange={onTemplateCodeOverrideChange}
            />
          ))}
        </div>
      </Accordion>
    </div>
  );
}
