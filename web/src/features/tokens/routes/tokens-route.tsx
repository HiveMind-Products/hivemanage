import { Suspense } from "react";
import { CreateTokenDialog } from "../components/CreateTokenDialog";
import { TokensTable } from "../components/TokensTable";
import { Skeleton } from "@/components/ui/skeleton";
import { useParams } from "react-router";
import { Params } from "@/typings/router";
import { usePermission } from "@/features/auth/hooks/use-permission";

function TokensTableSkeleton() {
  return (
    <div className="overflow-hidden rounded-lg border bg-card">
      <div className="flex h-10 items-center gap-4 border-b bg-muted/40 px-4">
        <Skeleton className="h-3.5 w-24" />
        <Skeleton className="h-3.5 w-16" />
      </div>
      {[...Array(3)].map((_, i) => (
        <div
          key={i}
          className="flex h-12 items-center gap-4 border-t border-border/60 px-4 first:border-t-0"
        >
          <Skeleton className="h-4 w-48" />
          <Skeleton className="h-4 w-14" />
          <Skeleton className="ml-auto size-8 rounded-md" />
        </div>
      ))}
    </div>
  );
}

export default function TokensRoute() {
  const params = useParams<Params>();
  const canWrite = usePermission("tokens", "write");

  if (!params.organizationId) {
    return <div>Organization ID is required</div>;
  }

  return (
    <div className="mx-auto w-full max-w-5xl space-y-6">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold tracking-tight">API tokens</h1>
          <p className="text-sm text-muted-foreground">
            Authenticate your servers and scripts with the platform's API.
          </p>
        </div>
        {canWrite && <CreateTokenDialog />}
      </div>
      <Suspense fallback={<TokensTableSkeleton />}>
        <TokensTable organizationId={params.organizationId} canWrite={canWrite} />
      </Suspense>
    </div>
  );
}
