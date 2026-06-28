import { useCallback, useState } from "react";
import { toast } from "sonner";

/**
 * Copy text to the clipboard with a confirmation toast and a short-lived
 * `copied` flag for inline affordances (e.g. swapping a copy icon for a check).
 */
export function useCopyToClipboard(timeout = 1500) {
  const [copied, setCopied] = useState(false);

  const copy = useCallback(
    async (text: string, message = "Copied to clipboard") => {
      try {
        await navigator.clipboard.writeText(text);
        setCopied(true);
        toast.success(message);
        window.setTimeout(() => setCopied(false), timeout);
        return true;
      } catch {
        toast.error("Couldn't copy to clipboard");
        return false;
      }
    },
    [timeout],
  );

  return { copied, copy };
}
