export function userFacingErrorMessage(error: unknown): string {
  const message = error instanceof Error ? error.message : String(error);
  if (
    message.includes("Cannot read properties of undefined") &&
    message.includes("main")
  ) {
    return "Desktop bridge unavailable. Open the Wails desktop app to scan or export workbooks.";
  }
  return message;
}
