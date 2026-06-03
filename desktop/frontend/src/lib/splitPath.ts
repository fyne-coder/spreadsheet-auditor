export function splitPath(path: string): { filename: string; directory: string } {
  if (!path) {
    return { filename: "", directory: "" };
  }
  const parts = path.split("/");
  const filename = parts.pop() || path;
  const directory = parts.length > 0 ? `${parts.join("/")}/` : "";
  return { filename, directory };
}
