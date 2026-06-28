import { useParams } from "react-router";
import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  Database,
  FileText,
  HardDrive,
  Inbox,
  type LucideIcon,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { EmptyState } from "@/components/ui/empty-state";
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
    queryFn: () =>
      fetchApi<OrganizationUsage>(
        "/api/dash/organization/" + organizationId + "/usage",
      ),
  });

  if (isLoading) {
    return (
      <div className="mx-auto w-full max-w-6xl space-y-6">
        <div className="space-y-2">
          <Skeleton className="h-6 w-32" />
          <Skeleton className="h-4 w-72" />
        </div>
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <Skeleton key={i} className="h-[88px] rounded-lg" />
          ))}
        </div>
        <div className="grid gap-4 lg:grid-cols-2">
          <Skeleton className="h-64 rounded-lg" />
          <Skeleton className="h-64 rounded-lg" />
        </div>
        <Skeleton className="h-64 rounded-lg" />
      </div>
    );
  }

  const summary: { title: string; value: string; icon: LucideIcon }[] = [
    {
      title: "Storage used",
      value: formatBytes(usage?.totalSize ?? 0),
      icon: HardDrive,
    },
    { title: "Files", value: formatNumber(usage?.totalFiles), icon: FileText },
    {
      title: "Logs",
      value: usage?.loggingAvailable ? formatNumber(usage?.totalLogs) : "—",
      icon: Activity,
    },
    {
      title: "Datasets",
      value: formatNumber(usage?.logsByDataset?.length ?? 0),
      icon: Database,
    },
  ];

  return (
    <div className="mx-auto w-full max-w-6xl space-y-6">
      <div>
        <h1 className="text-lg font-semibold tracking-tight">Usage</h1>
        <p className="text-sm text-muted-foreground">
          Storage and logging breakdown for this organization.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {summary.map((card) => (
          <Card key={card.title}>
            <CardContent className="flex items-center gap-3 p-4">
              <span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
                <card.icon className="size-[18px]" />
              </span>
              <div className="min-w-0">
                <div className="text-xs font-medium text-muted-foreground">
                  {card.title}
                </div>
                <div className="truncate text-xl font-semibold tracking-tight tabular-nums">
                  {card.value}
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-semibold">Storage by type</CardTitle>
          </CardHeader>
          <CardContent>
            <StorageBreakdown usage={usage} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-semibold">Largest files</CardTitle>
          </CardHeader>
          <CardContent>
            {usage && usage.largestFiles.length > 0 ? (
              <ul className="space-y-2.5">
                {usage.largestFiles.map((file) => (
                  <li
                    key={file.id}
                    className="flex items-center justify-between gap-3 text-sm"
                  >
                    <span className="truncate" title={file.name}>
                      {file.name}
                    </span>
                    <span className="shrink-0 text-muted-foreground tabular-nums">
                      {formatBytes(file.size)}
                    </span>
                  </li>
                ))}
              </ul>
            ) : (
              <p className="py-4 text-sm text-muted-foreground">
                No files uploaded yet.
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-semibold">
            Logs over the last 14 days
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!usage?.loggingAvailable ? (
            <p className="py-4 text-sm text-muted-foreground">
              Logging is currently unavailable.
            </p>
          ) : (
            <LogsChart data={usage.logsTimeseries} />
          )}
        </CardContent>
      </Card>

      {usage?.loggingAvailable && usage.logsByDataset.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-semibold">Logs by dataset</CardTitle>
          </CardHeader>
          <CardContent>
            <DatasetBreakdown datasets={usage.logsByDataset} />
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="text-sm font-semibold">Platform limits</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {PLATFORM_LIMITS.map((limit) => (
              <div key={limit.label} className="rounded-md border bg-muted/30 p-3">
                <div className="text-xs text-muted-foreground">{limit.label}</div>
                <div className="mt-0.5 text-base font-semibold tabular-nums">
                  {limit.value}
                </div>
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
    return (
      <EmptyState
        icon={Inbox}
        title="No storage used"
        description="Upload files to see how your storage breaks down by type."
        className="border-0 py-8"
      />
    );
  }
  const total = usage.storageByCategory.reduce((sum, c) => sum + c.size, 0) || 1;
  return (
    <div className="space-y-3.5">
      {usage.storageByCategory.map((category) => {
        const pct = Math.round((category.size / total) * 100);
        return (
          <div key={category.category} className="space-y-1.5">
            <div className="flex items-center justify-between text-sm">
              <span className="capitalize">{category.category}</span>
              <span className="text-muted-foreground tabular-nums">
                {formatBytes(category.size)} · {category.count} file
                {category.count === 1 ? "" : "s"}
              </span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary transition-[width] duration-500 ease-out"
                style={{ width: `${pct}%` }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}

function DatasetBreakdown({
  datasets,
}: {
  datasets: { datasetId: string; name: string; count: number }[];
}) {
  const sorted = [...datasets].sort((a, b) => b.count - a.count);
  const max = Math.max(...sorted.map((d) => d.count), 1);
  return (
    <div className="space-y-3.5">
      {sorted.map((ds) => {
        const pct = Math.round((ds.count / max) * 100);
        return (
          <div key={ds.datasetId} className="space-y-1.5">
            <div className="flex items-center justify-between text-sm">
              <span className="truncate" title={ds.name}>
                {ds.name}
              </span>
              <span className="shrink-0 text-muted-foreground tabular-nums">
                {formatNumber(ds.count)}
              </span>
            </div>
            <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
              <div
                className="h-full rounded-full bg-primary/70 transition-[width] duration-500 ease-out"
                style={{ width: `${Math.max(pct, 2)}%` }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}

function LogsChart({ data }: { data: LogDay[] }) {
  if (!data || data.length === 0) {
    return (
      <p className="py-4 text-sm text-muted-foreground">
        No logs in this period.
      </p>
    );
  }
  const max = Math.max(...data.map((d) => d.count), 1);
  return (
    <div className="flex h-44 items-end gap-1.5">
      {data.map((day) => (
        <div
          key={day.date}
          className="group flex flex-1 flex-col items-center justify-end gap-1.5"
          title={`${day.date}: ${day.count.toLocaleString()} logs`}
        >
          <div className="flex w-full flex-1 items-end">
            <div
              className="w-full rounded-t-sm bg-primary/70 transition-colors group-hover:bg-primary"
              style={{ height: `${Math.max((day.count / max) * 100, 2)}%` }}
            />
          </div>
          <span className="text-[10px] tabular-nums text-muted-foreground">
            {day.date.slice(5)}
          </span>
        </div>
      ))}
    </div>
  );
}

type LogDay = { date: string; count: number };
