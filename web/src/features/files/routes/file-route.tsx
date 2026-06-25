import { useParams } from "react-router";
import { useFile, useFileURL } from "../api/use-file";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { Separator } from "@/components/ui/separator";
import {
  FileIcon,
  FileImage,
  FileVideo,
  ArrowLeft,
  Trash2,
} from "lucide-react";
import { toast } from "sonner";
import { Link } from "react-router";
import { FileViewer } from "../components/file/file-viewer";
import { FileHeader } from "../components/file/file-header";
import { usePermission } from "@/features/auth/hooks/use-permission";

function formatBytes(bytes: number, decimals = 2) {
  if (!+bytes) return "0 Bytes";

  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];

  const i = Math.floor(Math.log(bytes) / Math.log(k));

  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

export default function FileRoute() {
  const { organizationId, fileId } = useParams<{
    organizationId: string;
    fileId: string;
  }>();

  const canWrite = usePermission("storage", "write");
  const { data, isPending } = useFile(organizationId, fileId);
  const { data: signedURL } = useFileURL(organizationId, fileId);

  const handleCopyLink = (url: string) => {
    navigator.clipboard.writeText(url);
    toast.success("Link copied to clipboard");
  };

  const handleDownload = (url: string, filename: string) => {
    const link = document.createElement("a");
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (isPending) {
    return (
      <div className="container mx-auto py-10 max-w-5xl space-y-8 animate-in fade-in duration-500">
        <div className="flex items-center gap-4">
          <Skeleton className="h-10 w-10 rounded-full" />
          <div className="space-y-2">
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-4 w-32" />
          </div>
        </div>
        <Skeleton className="h-[400px] w-full rounded-xl" />
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-32 w-full" />
          <Skeleton className="h-32 w-full" />
        </div>
      </div>
    );
  }

  if (!data) {
    return (
      <div className="flex flex-col items-center justify-center min-h-[50vh] space-y-4">
        <div className="p-6 rounded-full bg-muted">
          <FileIcon className="h-12 w-12 text-muted-foreground" />
        </div>
        <h2 className="text-2xl font-bold tracking-tight">File not found</h2>
        <p className="text-muted-foreground text-center max-w-sm">
          The file you are looking for does not exist or you do not have
          permission to view it.
        </p>
        <Button variant="outline" asChild>
          <Link to={`/app/${organizationId}/storage`}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Back to Storage
          </Link>
        </Button>
      </div>
    );
  }

  const fileUrl = signedURL?.url ?? "";
  const fileName = data.originalName || data.key.split("/").pop() || data.key;

  const isImage = data.type === "image";
  const isVideo = data.type === "video";

  return (
    <div className="container mx-auto py-6 max-w-6xl space-y-6 animate-in fade-in slide-in-from-bottom-4 duration-500">
      <FileHeader
        organizationId={organizationId}
        fileUrl={fileUrl}
        fileName={fileName}
        file={data}
        onCopyLink={handleCopyLink}
        onDownload={handleDownload}
      />

      <Separator />

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 space-y-6">
          <FileViewer file={data} fileUrl={fileUrl} fileName={fileName} />
        </div>

        <div className="space-y-6">
          <Card className="border shadow-none">
            <CardHeader>
              <CardTitle className="text-lg">File Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-1">
                <span className="text-sm font-medium text-muted-foreground">
                  Filename
                </span>
                <p className="text-sm break-all font-mono bg-muted/50 p-2 rounded-md">
                  {fileName}
                </p>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <span className="text-sm font-medium text-muted-foreground">
                    Type
                  </span>
                  <div className="flex items-center gap-2">
                    {isImage ? (
                      <FileImage className="h-4 w-4 text-blue-500" />
                    ) : isVideo ? (
                      <FileVideo className="h-4 w-4 text-purple-500" />
                    ) : (
                      <FileIcon className="h-4 w-4 text-gray-500" />
                    )}
                    <p className="text-sm">
                      {data.type
                        ? data.type.split("/")[1]?.toUpperCase() || "UNKNOWN"
                        : "UNKNOWN"}
                    </p>
                  </div>
                </div>
                <div className="space-y-1">
                  <span className="text-sm font-medium text-muted-foreground">
                    Size
                  </span>
                  <p className="text-sm font-medium">
                    {formatBytes(data.size)}
                  </p>
                </div>
              </div>

              <Separator />

              <div className="space-y-1">
                <span className="text-sm font-medium text-muted-foreground">
                  Full Path
                </span>
                <p className="text-xs text-muted-foreground break-all">
                  {data.key}
                </p>
              </div>

              <div className="space-y-1">
                <span className="text-sm font-medium text-muted-foreground">
                  Asset ID
                </span>
                <p className="text-xs font-mono text-muted-foreground break-all">
                  {data.id}
                </p>
              </div>
            </CardContent>
          </Card>

          {(isImage || isVideo) && (
            <Card>
              <CardHeader className="pb-3">
                <CardTitle className="text-sm font-medium">
                  Properties
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="text-sm text-muted-foreground">
                  <p className="italic">No additional metadata available.</p>
                </div>
              </CardContent>
            </Card>
          )}

          {canWrite && (
            <Button variant="destructive" className="w-full">
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </Button>
          )}
        </div>
      </div>
    </div>
  );
}
