import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { Badge } from "@/components/ui/badge";
import { Asset, AssetURLResponse } from "@/typings/asset";
import { fetchApi } from "@/utils/http-util";
import { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import {
  Check,
  FileAudioIcon,
  ImageIcon,
  Link2,
  VideoIcon,
} from "lucide-react";
import type { MouseEvent } from "react";
import { Link, useParams } from "react-router";
import { useCopyToClipboard } from "@/hooks/use-copy";
import { cn } from "@/lib/utils";

export function formatFilename(key: string) {
  const parts = key.split("/");
  return parts[parts.length - 1];
}

function assetDisplayName(asset: Asset) {
  return asset.originalName || asset.key.replace(/^(image|video|audio)\//, "");
}

const typeIcon: Record<string, { Icon: typeof ImageIcon; tint: string }> = {
  image: { Icon: ImageIcon, tint: "text-info" },
  video: { Icon: VideoIcon, tint: "text-primary" },
  audio: { Icon: FileAudioIcon, tint: "text-success" },
};

export function assetColumns(): ColumnDef<Asset>[] {
  return [
    {
      accessorKey: "id",
      header: "Asset",
      cell: (info) => {
        const asset = info.row.original;
        const displayValue = assetDisplayName(asset);
        const { Icon, tint } = typeIcon[asset.type] ?? typeIcon.image;

        return (
          <Link
            to={asset.id}
            className="group flex items-center gap-2.5 py-0.5"
          >
            <span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-muted">
              <Icon className={cn("size-4", tint)} />
            </span>
            <span className="min-w-0">
              <span className="block truncate font-medium group-hover:text-primary">
                {formatFilename(displayValue)}
              </span>
            </span>
          </Link>
        );
      },
    },
    {
      accessorKey: "type",
      header: "Type",
      cell: (info) => {
        const type = info.getValue() as string;
        return (
          <Badge variant="outline" className="font-normal capitalize">
            {type}
          </Badge>
        );
      },
    },
    {
      accessorKey: "createdAt",
      header: "Created",
      cell: (info) => {
        const value = info.getValue();
        if (!value) return <span className="text-muted-foreground">—</span>;
        return (
          <span className="text-muted-foreground tabular-nums">
            {format(new Date(value as Date), "yyyy-MM-dd HH:mm")}
          </span>
        );
      },
    },
    {
      accessorKey: "key",
      header: () => <span className="sr-only">URL</span>,
      cell: (info) => <UrlCell asset={info.row.original} />,
    },
  ];
}

function UrlCell({ asset }: { asset: Asset }) {
  const { organizationId } = useParams<{ organizationId: string }>();
  const { copied, copy } = useCopyToClipboard();

  async function copySignedURL(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    if (!organizationId) return;

    const response = await fetchApi<AssetURLResponse>(
      `/api/dash/storage/${organizationId}/file/${asset.id}/url`,
    );
    if (response?.url) copy(response.url, "Link copied to clipboard");
  }

  return (
    <div className="flex justify-end">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <button
              onClick={copySignedURL}
              aria-label="Copy signed URL"
              className="flex size-8 cursor-pointer items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            >
              {copied ? (
                <Check className="size-4 text-success" />
              ) : (
                <Link2 className="size-4" />
              )}
            </button>
          </TooltipTrigger>
          <TooltipContent>
            <p>Copy signed URL</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}
