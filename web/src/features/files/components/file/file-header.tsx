import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Asset } from "@/typings/asset";
import { ArrowLeft, Copy, Download, ExternalLink } from "lucide-react";
import { Link } from "react-router";

interface FileHeaderProps {
  organizationId: string | undefined;
  fileUrl: string;
  fileName: string;
  file: Asset;
  onCopyLink: (url: string) => void;
  onDownload: (url: string, filename: string) => void;
}

function formatBytes(bytes: number, decimals = 2) {
  if (!+bytes) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

export function FileHeader({
  fileUrl,
  fileName,
  file,
  ...props
}: FileHeaderProps) {
  return (
    <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div className="space-y-1">
        <div className="flex items-center gap-2 text-sm text-muted-foreground mb-2">
          <Link
            to={`/app/${props?.organizationId}/storage`}
            className="hover:text-foreground transition-colors flex items-center gap-1"
          >
            <ArrowLeft className="h-3 w-3" /> Back to files
          </Link>
        </div>
        <h1
          className="max-w-2xl truncate text-xl font-semibold tracking-tight"
          title={fileName}
        >
          {fileName}
        </h1>
        <div className="flex items-center gap-2">
          <Badge variant="secondary" className="font-mono text-xs capitalize">
            {file.type}
          </Badge>
          <span className="text-sm text-muted-foreground">•</span>
          <span className="text-sm text-muted-foreground">
            {formatBytes(file.size)}
          </span>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => props.onCopyLink(fileUrl)}
        >
          <Copy className="mr-2 h-4 w-4" />
          Copy Link
        </Button>
        <Button variant="outline" size="sm" asChild>
          <a href={fileUrl} target="_blank" rel="noopener noreferrer">
            <ExternalLink className="mr-2 h-4 w-4" />
            Open Original
          </a>
        </Button>
        <Button size="sm" onClick={() => props.onDownload(fileUrl, fileName)}>
          <Download className="mr-2 h-4 w-4" />
          Download
        </Button>
      </div>
    </div>
  );
}
