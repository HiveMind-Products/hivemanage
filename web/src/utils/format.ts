const BYTE_UNITS = ["Bytes", "KB", "MB", "GB", "TB"];

export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return "0 Bytes";
  const k = 1024;
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(k)), BYTE_UNITS.length - 1);
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + BYTE_UNITS[i];
}

export function formatNumber(value: number | undefined | null): string {
  return (value ?? 0).toLocaleString();
}
