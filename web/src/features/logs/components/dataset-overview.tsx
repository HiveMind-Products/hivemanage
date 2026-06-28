import { lazy } from "react";
import { List } from "@/components/ui/list";
import { EmptyState } from "@/components/ui/empty-state";
import { Layers } from "lucide-react";
import { useNavigate } from "react-router";
import { useListDatasets } from "../api/dataset-api";
import { usePermission } from "@/features/auth/hooks/use-permission";
const DatasetSheet = lazy(() => import("./dataset-sheet/dataset-sheet"));

interface DatasetActionsProps {
  organizationId: string;
}

export default function DatasetOverview({
  organizationId,
}: DatasetActionsProps) {
  const { data: datasets } = useListDatasets(organizationId);
  const navigate = useNavigate();
  const canWrite = usePermission("logs", "write");

  function navigateToDataset(id: string) {
    navigate(`${id}`);
  }

  const isEmpty = datasets && datasets.length === 0;

  return (
    <div className="mx-auto w-full max-w-4xl space-y-5">
      <div className="flex items-end justify-between gap-4">
        <div>
          <h1 className="text-lg font-semibold tracking-tight">Datasets</h1>
          <p className="text-sm text-muted-foreground">
            Streams of logs ingested from your servers and scripts.
          </p>
        </div>
        {canWrite && !isEmpty && <DatasetSheet organizationId={organizationId} />}
      </div>

      {isEmpty ? (
        <EmptyState
          icon={Layers}
          title="No datasets yet"
          description="Create a dataset to start ingesting and querying logs."
          action={canWrite ? <DatasetSheet organizationId={organizationId} /> : undefined}
        />
      ) : (
        <List>
          {datasets &&
            datasets.map((dataset) => (
              <List.Item
                key={dataset.id}
                title={dataset.name}
                subtitle={dataset.description}
                onClick={() => navigateToDataset(dataset.id)}
              />
            ))}
        </List>
      )}
    </div>
  );
}
