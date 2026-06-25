import { Asset } from "@/typings/asset";
import { Card, CardFooter } from "@/components/ui/card";
import {
  ImageIcon,
  VideoIcon,
  FileAudioIcon,
  MoreVertical,
} from "lucide-react";
import { Link, useSearchParams } from "react-router";
import { Button } from "@/components/ui/button";

interface AssetGridViewProps {
  assets: Asset[];
  totalCount: number;
}

export function AssetGridView({ assets, totalCount }: AssetGridViewProps) {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = Number(searchParams.get("page") ?? "0");
  const pageSize = 20;
  const pageCount = Math.ceil(totalCount / pageSize);

  const handlePageChange = (newPage: number) => {
    const params = new URLSearchParams(searchParams);
    params.set("page", newPage.toString());
    setSearchParams(params, { replace: true });
  };

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
        {assets.map((asset) => (
          <AssetCard key={asset.id} asset={asset} />
        ))}
      </div>

      <div className="flex items-center justify-end space-x-2 py-4">
        <Button
          variant="outline"
          size="sm"
          onClick={() => handlePageChange(page - 1)}
          disabled={page <= 0}
        >
          Previous
        </Button>
        <div className="text-sm font-medium">
          Page {page + 1} of {pageCount || 1}
        </div>
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
  );
}

function AssetCard({ asset }: { asset: Asset }) {
  const type = asset.type;

  let Icon = ImageIcon;
  if (type === "video") {
    Icon = VideoIcon;
  } else if (type === "audio") {
    Icon = FileAudioIcon;
  }

  const filename = asset.originalName || asset.key.split("/").pop() || asset.key;

  return (
    <Card className="overflow-hidden group cursor-pointer hover:border-primary/50 transition-colors py-0 gap-0">
      <Link to={asset.id} className="contents">
        <div className="aspect-video w-full bg-muted flex items-center justify-center relative overflow-hidden">
          <Icon className="w-10 h-10 text-muted-foreground" />
        </div>
        <CardFooter className="p-3 flex items-center justify-between">
          <div className="truncate flex-1">
            <p className="text-sm font-medium truncate" title={filename}>
              {filename}
            </p>
            <p className="text-xs text-muted-foreground uppercase">{type}</p>
          </div>
          <button className="p-1 hover:bg-accent rounded-md opacity-0 group-hover:opacity-100 transition-opacity">
            <MoreVertical className="w-4 h-4" />
          </button>
        </CardFooter>
      </Link>
    </Card>
  );
}
