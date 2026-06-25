import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Asset, AssetURLResponse } from "@/typings/asset";
import { fetchApi } from "@/utils/http-util";
import { ColumnDef } from "@tanstack/react-table";
import { format } from "date-fns";
import { ClipboardCopy, FileAudioIcon, ImageIcon, VideoIcon } from "lucide-react";
import type { MouseEvent } from "react";
import { Link, useParams } from "react-router";

export function formatFilename(key: string) {
  const parts = key.split("/");
  return parts[parts.length - 1];
}

function assetDisplayName(asset: Asset) {
  return asset.originalName || asset.key.replace(/^(image|video|audio)\//, "");
}

export function assetColumns(): ColumnDef<Asset>[] {
  return [
    {
      accessorKey: "id",
      header: "Asset",
      cell: (info) => {
        const asset = info.row.original;
        const displayValue = assetDisplayName(asset);
        let Icon = ImageIcon;

        if (asset.type === "video") Icon = VideoIcon;
        else if (asset.type === "audio") Icon = FileAudioIcon;

        return (
          <Link to={asset.id} className="flex items-center">
            <Icon size={18} className="mr-2 text-gray-400" />
            <span>
              <p className="hover:underline">{formatFilename(displayValue)}</p>
              <p className="text-xs text-muted-foreground">{asset.type}</p>
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
        return type.charAt(0).toUpperCase() + type.slice(1);
      },
    },
    {
      accessorKey: "createdAt",
      header: "Created At",
      cell: (info) => format(new Date(info.getValue() as Date), "yyyy-MM-dd HH:mm:ss"),
    },
    {
      accessorKey: "key",
      header: "URL",
      cell: (info) => <UrlCell asset={info.row.original} />,
    },
  ];
}

function UrlCell({ asset }: { asset: Asset }) {
  const { organizationId } = useParams<{ organizationId: string }>();

  async function copySignedURL(event: MouseEvent<HTMLButtonElement>) {
    event.preventDefault();
    if (!organizationId) return;

    const response = await fetchApi<AssetURLResponse>(
      `/api/dash/storage/${organizationId}/file/${asset.id}/url`,
    );
    if (response?.url) await navigator.clipboard.writeText(response.url);
  }

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <button onClick={copySignedURL} className="cursor-pointer rounded-md bg-accent/60 p-2 hover:bg-accent">
            <ClipboardCopy size={16} />
          </button>
        </TooltipTrigger>
        <TooltipContent>
          <p>Copy signed URL</p>
        </TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}