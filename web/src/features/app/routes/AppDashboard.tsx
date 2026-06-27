import { useParams } from "react-router";
import { useOrganizationStats } from "@/features/organizations/api/useOrganizationStats";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { HardDrive, Activity, Key, FileText } from "lucide-react";
import { Skeleton } from "@/components/ui/skeleton";

export const AppDashboard: React.FC = () => {
  const { organizationId } = useParams();
  const { data: stats, isLoading } = useOrganizationStats(organizationId);

  if (isLoading) {
    return (
      <div className="flex flex-col gap-8">
        <div>
          <Skeleton className="h-9 w-[150px] mb-2" />
          <Skeleton className="h-5 w-[250px]" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {[...Array(4)].map((_, i) => (
            <Card key={i}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <Skeleton className="h-4 w-[100px]" />
                <Skeleton className="h-4 w-4 rounded-full" />
              </CardHeader>
              <CardContent>
                <Skeleton className="h-8 w-[60px] mb-1" />
                <Skeleton className="h-3 w-[120px]" />
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    );
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 Bytes";
    const k = 1024;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const statCards = [
    {
      title: "Total Logs",
      value: stats?.totalLogs.toLocaleString() ?? "0",
      icon: Activity,
      description: "Aggregated logs from all datasets",
    },
    {
      title: "Files Uploaded",
      value: stats?.totalFiles.toLocaleString() ?? "0",
      icon: FileText,
      description: "Total assets and media files",
    },
    {
      title: "Storage Used",
      value: formatBytes(stats?.totalSize || 0),
      icon: HardDrive,
      description: "Total disk space consumed",
    },
    {
      title: "Active Tokens",
      value: stats?.totalTokens.toLocaleString() ?? "0",
      icon: Key,
      description: "Access tokens for API usage",
    },
  ];

  return (
    <div className="flex flex-col gap-8">
      <div>
        <h2 className="text-3xl font-bold tracking-tight">Dashboard</h2>
        <p className="text-muted-foreground">
          Welcome to your organization overview.
        </p>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {statCards.map((stat, i) => (
          <Card key={i}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {stat.title}
              </CardTitle>
              <stat.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground mt-1">
                {stat.description}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
};
