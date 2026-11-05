// Authany: upstream sends these requests to its own help desk form; ours go to
// hello@authany.com as an email with the known details prefilled.
const HELP_DESK_EMAIL = "hello@authany.com";

export type HelpDeskPurpose =
  | "project-limit"
  | "sms-otp"
  | "custom-messages"
  | "resume-suspended"
  | "other";

export interface HelpDeskURLOptions {
  purpose: HelpDeskPurpose;
  projectID?: string | null;
  email?: string | null;
  fullName?: string | null;
}

// Unknown values are omitted so the email leaves those lines out.
export function makeHelpDeskURL(options: HelpDeskURLOptions): string {
  const { purpose, projectID, email, fullName } = options;
  const lines: string[] = [];
  if (projectID) {
    lines.push(`Project: ${projectID}`);
  }
  if (email) {
    lines.push(`Email: ${email}`);
  }
  if (fullName) {
    lines.push(`Name: ${fullName}`);
  }
  lines.push(`Purpose: ${purpose}`);
  const subject = `Authany console request: ${purpose}`;
  return (
    `mailto:${HELP_DESK_EMAIL}` +
    `?subject=${encodeURIComponent(subject)}` +
    `&body=${encodeURIComponent(lines.join("\n") + "\n\n")}`
  );
}
