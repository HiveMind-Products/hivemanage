import { cn } from "@/lib/utils";

/**
 * hivemanage hive-cell mark. The outer cell uses the brand accent; the inner
 * cell is filled. Sized via className (defaults to 1.5rem square).
 */
export function Logo({ className }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 32 32"
      fill="none"
      role="img"
      aria-hidden="true"
      className={cn("size-6 shrink-0", className)}
    >
      <path
        d="M26 16 L21.5 23.79 L12.5 23.79 L8 16 L12.5 8.21 L21.5 8.21 Z"
        fill="none"
        stroke="var(--primary)"
        strokeWidth="2.1"
        strokeLinejoin="round"
      />
      <path
        d="M19.6 16 L17.8 19.12 L14.2 19.12 L12.4 16 L14.2 12.88 L17.8 12.88 Z"
        fill="var(--primary)"
      />
    </svg>
  );
}

/**
 * Full lockup: hive mark + "hivemanage" wordmark. The "hive" syllable is
 * weighted to read as the product name; the rest stays quiet.
 */
export function Wordmark({
  className,
  markClassName,
}: {
  className?: string;
  markClassName?: string;
}) {
  return (
    <span className={cn("flex items-center gap-2", className)}>
      <Logo className={markClassName} />
      <span className="text-[0.95rem] font-semibold tracking-tight text-foreground">
        hive<span className="text-muted-foreground font-medium">manage</span>
      </span>
    </span>
  );
}
