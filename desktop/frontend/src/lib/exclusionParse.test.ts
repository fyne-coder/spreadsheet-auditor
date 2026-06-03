import { describe, expect, it } from "vitest";
import { parseExclusionList } from "./exclusionParse";

describe("parseExclusionList", () => {
  it("splits on commas and newlines", () => {
    expect(parseExclusionList("SheetA, SheetB\nModel!A1")).toEqual([
      "SheetA",
      "SheetB",
      "Model!A1",
    ]);
  });

  it("drops empty tokens", () => {
    expect(parseExclusionList("  ,,\n  ")).toEqual([]);
  });
});
