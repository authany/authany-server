import { describe, it, expect } from "@jest/globals";
import { makeHelpDeskURL } from "./helpDesk";

describe("makeHelpDeskURL", () => {
  it("includes every known value", () => {
    expect(
      makeHelpDeskURL({
        purpose: "sms-otp",
        projectID: "myapp",
        email: "user+1@example.com",
        fullName: "Chan Tai Man",
      })
    ).toEqual(
      "mailto:hello@authany.com?subject=Authany%20console%20request%3A%20sms-otp&body=Project%3A%20myapp%0AEmail%3A%20user%2B1%40example.com%0AName%3A%20Chan%20Tai%20Man%0APurpose%3A%20sms-otp%0A%0A"
    );
  });

  it("omits unknown values", () => {
    expect(
      makeHelpDeskURL({
        purpose: "project-limit",
        projectID: null,
        email: "user@example.com",
        fullName: "",
      })
    ).toEqual(
      "mailto:hello@authany.com?subject=Authany%20console%20request%3A%20project-limit&body=Email%3A%20user%40example.com%0APurpose%3A%20project-limit%0A%0A"
    );
  });
});
