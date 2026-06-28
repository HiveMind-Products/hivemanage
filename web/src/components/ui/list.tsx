import { type HTMLAttributes } from "react";
import { ChevronRight } from "lucide-react";
import { cn } from "@/lib/utils";

interface ListProps {
  children: React.ReactNode;
  className?: string;
}

interface ListHeaderProps {
  title: string;
  children?: React.ReactNode;
}

interface ListItemProps extends HTMLAttributes<HTMLDivElement> {
  title: string;
  subtitle?: string;
  rightContent?: React.ReactNode;
}

export function List({ children, className }: ListProps) {
  return (
    <div
      className={cn(
        "overflow-hidden rounded-lg border bg-card",
        className,
      )}
    >
      {children}
    </div>
  );
}

function Header(props: ListHeaderProps) {
  return (
    <div className="flex items-center justify-between gap-2 border-b bg-muted/40 px-4 py-2.5">
      <h2 className="text-sm font-semibold tracking-tight">{props.title}</h2>
      {props.children}
    </div>
  );
}

function Item({ title, subtitle, rightContent, className, ...props }: ListItemProps) {
  const interactive = typeof props.onClick === "function";
  return (
    <div
      {...props}
      className={cn(
        "group flex items-center justify-between gap-3 border-t border-border/60 px-4 py-3 transition-colors first:border-t-0",
        interactive && "cursor-pointer hover:bg-accent/50",
        className,
      )}
    >
      <div className="min-w-0">
        <div className="truncate text-sm font-medium">{title}</div>
        {subtitle && (
          <p className="truncate text-xs text-muted-foreground">{subtitle}</p>
        )}
      </div>
      <div className="flex shrink-0 items-center gap-3">
        {rightContent}
        {interactive && (
          <ChevronRight className="size-4 text-muted-foreground/50 transition-colors group-hover:text-muted-foreground" />
        )}
      </div>
    </div>
  );
}

List.Header = Header;
List.Item = Item;
