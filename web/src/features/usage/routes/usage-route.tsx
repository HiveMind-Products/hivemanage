import { useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { Activity, Database, FileText, HardDrive } from "lucide-react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { fetchApi } from "@/utils/http-util";
import { formatBytes, formatNumber } from "@/utils/format";
import { QueryKeys } from "@/typings/query";
import type { OrganizationUsage } from "@/typings/usage";

const PLATFORM_LIMITS = [
  { label: "Max file upload", value: "500 MB" },
  { label: "Max logs per request", value: "1,000" },
  { label: "Max log message size", value: "8,192 chars" },
  { label: "API rate limit", value: "20 req/s" },
];

export default function UsageRoute() {
  const { organizationId } = useParams<{ organizationId: string }>();
  const { data: usage, isLoading } = useQuery({
    queryKey: [QueryKeys.Usage, organizationId],
    enabled: !!organizationId,
    queryFn: () => fetchApi<OrganizationUsage>("/api/dash/organization/" + organizationId + "/usage"),
  });

  if (isLoading) {
    return (
      <div className="container mx-auto max-w-6xl space-y-6 py-2">
        <Skeleton className="h-9 w-40" />
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-28" />
          ))}
        </div>
        <Skeleton className="h-64" />
      </div>
    );
  }

  const summary = [
    { title: "Storage Used", value: formatBytes(usage?.totalSize ?? 0), icon: HardDrive },
    { title: "Files", value: formatNumber(usage?.totalFiles), icon: FileText },
    { title: "Logs", value: usage?.loggingAvailable ? formatNumber(usage?.totalLogs) : "—", icon: Activity },
    { title: "Datasets", value: formatNumber(usage?.logsByDataset?.length ?? 0), icon: Database },
  ];

  return (
    <div className="container mx-auto max-w-6xl space-y-6 py-2">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Usage</h1>
        <p className="text-muted-foreground">Storage and logging breakdown for this organization.</p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {summary.map((card) => (
          <Card key={card.title}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">{card.title}</CardTitle>
              <card.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{card.value}</div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Storage by type</CardTitle>
          </CardHeader>
          <CardContent>
            <StorageBreakdown usage={usage} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Largest files</CardTitle>
          </CardHeader>
          <CardContent>
            {usage && usage.largestFiles.length > 0 ? (
              <ul className="space-y-2">
                {usage.largestFiles.map((file) => (
                  <li key={file.id} className="flex items-center justify-between gap-3 text-sm">
                    <span className="truncate" title={file.name}>{file.name}</span>
                    <span className="shrink-0 text-muted-foreground">{formatBytes(file.size)}</span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-muted-foreground">No files uploaded yet.</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Logs over the last 14 days</CardTitle>
        </CardHeader>
        <CardContent>
          {!usage?.loggingAvailable ? (
            <p className="text-sm text-muted-foreground">Logging is currently unavailable.</p>
          ) : (
            <LogsChart data={usage.logsTimeseries} />
          )}
        </CardContent>
      </Card>

      {usage?.loggingAvailable && usage.logsByDataset.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Logs by dataset</CardTitle>
          </CardHeader>
          <CardContent>
            <ul className="space-y-2">
              {[...usage.logsByDataset]
                .sort((a, b) => b.count - a.count)
                .map((ds) => (
                  <li key={ds.datasetId} className="flex items-center justify-between gap-3 text-sm">
                    <span className="truncate" title={ds.name}>{ds.name}</span>
                    <span className="shrink-0 text-muted-foreground">{formatNumber(ds.count)}</span>
                  </li>
                ))}
            </ul>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Platform limits</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {PLATFORM_LIMITS.map((limit) => (
              <div key={limit.label} className="rounded-md border p-3">
                <div className="text-xs text-muted-foreground">{limit.label}</div>
                <div className="text-lg font-semibold">{limit.value}</div>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}

function StorageBreakdown({ usage }: { usage: OrganizationUsage | undefined }) {
  if (!usage || usage.storageByCategory.length === 0) {
    return <p className="text-sm text-muted-foreground">No files uploaded yet.</p>;
  }
  const total = usage.storageByCategory.reduce((sum, c) => sum + c.size, 0) || 1;
  return (
    <div className="space-y-3">
      {usage.storageByCategory.map((category) => {
        const pct = Math.round((category.size / total) * 100);
        return (
          <div key={category.category} className="space-y-1">
            <div className="flex items-center justify-between text-sm">
              <span className="capitalize">{category.category}</span>
              <span className="text-muted-foreground">
                {formatBytes(category.size)} · {category.count} file{category.count === 1 ? "" : "s"}
              </span>
            </div>
            <div className="h-2 w-full rounded-full bg-muted">
              <div className="h-2 rounded-full bg-primary" style={{ width: `${pct}%` }} />
            </div>
          </div>
        );
      })}
    </div>
  );
}

function LogsChart({ data }: { data: LogDay[] }) {
  if (!data || data.length === 0) {
    return <p className="text-sm text-muted-foreground">No logs in this period.</p>;
  }
  const max = Math.max(...data.map((d) => d.count), 1);
  return (
    <div className="flex h-40 items-end gap-1">
      {data.map((day) => (
        <div key={day.date} className="group flex flex-1 flex-col items-center justify-end gap-1">
          <div
            className="w-full rounded-t bg-primary/80 transition-colors group-hover:bg-primary"
            style={{ height: `${Math.max((day.count / max) * 100, 2)}%` }}
            title={`${day.date}: ${day.count.toLocaleString()} logs`}
          />
          <span className="text-[10px] text-muted-foreground">{day.date.slice(5)}</span>
        </div>
      ))}
    </div>
  );
}

type LogDay = { date: string; count: number };
