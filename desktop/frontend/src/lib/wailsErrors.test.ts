import { describe, expect, it } from "vitest";
import { userFacingErrorMessage } from "./wailsErrors";

describe("userFacingErrorMessage", () => {
  it("maps missing Wails bridge errors to a user-facing message", () => {
    expect(
      userFacingErrorMessage(
        new TypeError("Cannot read properties of undefined (reading 'main')"),
      ),
    ).toBe(
      "Desktop bridge unavailable. Open the Wails desktop app to scan or export workbooks.",
    );
  });

  it("preserves ordinary backend errors", () => {
    expect(userFacingErrorMessage(new Error("permission denied"))).toBe(
      "permission denied",
    );
  });
});
