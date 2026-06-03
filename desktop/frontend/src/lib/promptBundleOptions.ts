import { model } from "../../wailsjs/go/models";
import { parseExclusionList } from "./exclusionParse";

export function buildPromptBundleOptions(
  excludeSheetsRaw: string,
  excludeCellsRaw: string,
  userObjectiveRaw: string,
  maxPacketBytes?: number,
): model.PromptBundleOptions {
  const excludeSheets = parseExclusionList(excludeSheetsRaw);
  const excludeCells = parseExclusionList(excludeCellsRaw);
  const userObjective = userObjectiveRaw.trim();
  return new model.PromptBundleOptions({
    exclude_sheets: excludeSheets.length > 0 ? excludeSheets : undefined,
    exclude_cells: excludeCells.length > 0 ? excludeCells : undefined,
    max_packet_bytes: maxPacketBytes,
    user_objective: userObjective || undefined,
  });
}
