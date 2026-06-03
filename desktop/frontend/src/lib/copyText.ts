export async function copyText(label: string, value: string): Promise<void> {
  if (!value) {
    throw new Error(`Nothing to copy for ${label}.`);
  }
  await navigator.clipboard.writeText(value);
}
