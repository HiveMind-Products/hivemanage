import { Asset, AssetURLResponse } from "@/typings/asset";
import { Card } from "@/components/ui/card";
import {
  ImageIcon,
  Link2,
  VideoIcon,
  FileAudioIcon,
  X,
} from "lucide-react";
import { Link, useParams, useSearchParams } from "react-router";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { useQuery } from "@tanstack/react-query";
import { useEffect, useRef, useState } from "react";
import { fetchApi } from "@/utils/http-util";
import { useCopyToClipboard } from "@/hooks/use-copy";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

interface AssetGridViewProps {
  assets: Asset[];
  totalCount: number;
}

const typeMeta: Record<string, { Icon: typeof ImageIcon; tint: string }> = {
  image: { Icon: ImageIcon, tint: "text-info" },
  video: { Icon: VideoIcon, tint: "text-primary" },
  audio: { Icon: FileAudioIcon, tint: "text-success" },
};

function assetName(asset: Asset) {
  return asset.originalName || asset.key.split("/").pop() || asset.key;
}

async function fetchAssetUrl(organizationId: string, fileId: string) {
  const res = await fetchApi<AssetURLResponse>(
    `/api/dash/storage/${organizationId}/file/${fileId}/url`,
  );
  return res?.url ?? "";
}

export function AssetGridView({ assets, totalCount }: AssetGridViewProps) {
  const params = useParams<{ organizationId: string }>();
  const organizationId = params.organizationId as string;
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "0");
  const pageSize = 20;
  const pageCount = Math.ceil(totalCount / pageSize);

  const [selected, setSelected] = useState<Set<string>>(new Set());

  const handlePageChange = (newPage: number) => {
    const next = new URLSearchParams(searchParams);
    next.set("page", newPage.toString());
    setSearchParams(next, { replace: true });
    setSelected(new Set());
  };

  const toggle = (id: string) =>
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });

  const copySelectedLinks = async () => {
    const ids = [...selected];
    const urls = await Promise.all(
      ids.map((id) => fetchAssetUrl(organizationId, id)),
    );
    const joined = urls.filter(Boolean).join("\n");
    if (!joined) return toast.error("Couldn't fetch links");
    await navigator.clipboard.writeText(joined);
    toast.success(
      `Copied ${ids.length} link${ids.length > 1 ? "s" : ""} to clipboard`,
    );
  };

  const from = totalCount === 0 ? 0 : page * pageSize + 1;
  const to = Math.min((page + 1) * pageSize, totalCount);

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5">
        {assets.map((asset) => (
          <AssetCard
            key={asset.id}
            asset={asset}
            organizationId={organizationId}
            selected={selected.has(asset.id)}
            selecting={selected.size > 0}
            onToggle={() => toggle(asset.id)}
          />
        ))}
      </div>

      <div className="flex items-center justify-between border-t border-border/60 pt-4">
        <p className="text-sm text-muted-foreground tabular-nums">
          {from}–{to} of {totalCount.toLocaleString()}
        </p>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(page - 1)}
            disabled={page <= 0}
          >
            Previous
          </Button>
          <span className="px-1 text-sm tabular-nums text-muted-foreground">
            {page + 1} / {pageCount || 1}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => handlePageChange(page + 1)}
            disabled={page >= pageCount - 1}
          >
            Next
          </Button>
        </div>
      </div>

      {/* Floating selection bar */}
      {selected.size > 0 && (
        <div className="fixed bottom-6 left-1/2 z-30 flex -translate-x-1/2 items-center gap-3 rounded-full border bg-popover/95 py-2 pl-4 pr-2 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-popover/80">
          <span className="text-sm font-medium tabular-nums">
            {selected.size} selected
          </span>
          <Button size="sm" variant="secondary" onClick={copySelectedLinks}>
            <Link2 className="size-4" />
            Copy links
          </Button>
          <Button
            size="icon"
            variant="ghost"
            className="size-8 rounded-full"
            onClick={() => setSelected(new Set())}
            aria-label="Clear selection"
          >
            <X className="size-4" />
          </Button>
        </div>
      )}
    </div>
  );
}

function AssetCard({
  asset,
  organizationId,
  selected,
  selecting,
  onToggle,
}: {
  asset: Asset;
  organizationId: string;
  selected: boolean;
  selecting: boolean;
  onToggle: () => void;
}) {
  const meta = typeMeta[asset.type] ?? typeMeta.image;
  const { Icon, tint } = meta;
  const filename = assetName(asset);
  const { copy } = useCopyToClipboard();

  const ref = useRef<HTMLDivElement>(null);
  const inView = useInViewOnce(ref);
  const isImage = asset.type === "image";

  // Shares the same query key as useFileURL so the cache is reused elsewhere.
  const { data } = useQuery({
    queryKey: ["file-url", organizationId, asset.id],
    enabled: isImage && inView && !!organizationId,
    staleTime: 1000 * 60 * 5,
    queryFn: () => fetchAssetUrl(organizationId, asset.id),
  });

  const handleCopy = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const url = data || (await fetchAssetUrl(organizationId, asset.id));
    if (url) copy(url, "Link copied to clipboard");
  };

  return (
    <Card
      ref={ref}
      className={cn(
        "group relative gap-0 overflow-hidden py-0 transition-all duration-150 hover:border-border-strong hover:shadow-sm",
        selected && "border-primary ring-1 ring-primary",
      )}
    >
      {/* Selection checkbox — visible on hover or when selecting */}
      <div
        className={cn(
          "absolute left-2 top-2 z-10 transition-opacity",
          selected || selecting
            ? "opacity-100"
            : "opacity-0 group-hover:opacity-100",
        )}
      >
        <Checkbox
          checked={selected}
          onCheckedChange={onToggle}
          aria-label={`Select ${filename}`}
          className="border-border-strong bg-background/80 backdrop-blur"
        />
      </div>

      {/* Copy-link affordance */}
      <button
        type="button"
        onClick={handleCopy}
        aria-label="Copy link"
        className="absolute right-2 top-2 z-10 flex size-7 items-center justify-center rounded-md border border-border-strong bg-background/80 text-muted-foreground opacity-0 backdrop-blur transition-all hover:text-foreground group-hover:opacity-100"
      >
        <Link2 className="size-3.5" />
      </button>

      <Link to={asset.id} className="contents">
        <div className="relative flex aspect-video w-full items-center justify-center overflow-hidden bg-muted/60">
          {isImage && data ? (
            <img
              src={data}
              alt={filename}
              loading="lazy"
              className="size-full object-cover"
            />
          ) : (
            <Icon className={cn("size-9", tint, "opacity-80")} />
          )}
        </div>
        <div className="flex items-center gap-2 p-3">
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium" title={filename}>
              {filename}
            </p>
            <p className="text-xs uppercase tracking-wide text-muted-foreground">
              {asset.type}
            </p>
          </div>
        </div>
      </Link>
    </Card>
  );
}

/** Fire-once IntersectionObserver so off-screen thumbnails don't fetch. */
function useInViewOnce(ref: React.RefObject<HTMLElement | null>) {
  const [inView, setInView] = useState(false);
  useEffect(() => {
    const el = ref.current;
    if (!el || inView) return;
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setInView(true);
          observer.disconnect();
        }
      },
      { rootMargin: "200px" },
    );
    observer.observe(el);
    return () => observer.disconnect();
  }, [ref, inView]);
  return inView;
}
