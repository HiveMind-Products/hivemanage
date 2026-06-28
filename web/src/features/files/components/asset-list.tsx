import type { ColumnDef, RowSelectionState } from "@tanstack/react-table";
import { useMemo, useState } from "react";
import { DataTable } from "@/components/data/DataTable";
import { Asset } from "@/typings/asset";
import { useParams, useSearchParams } from "react-router";
import { Params } from "@/typings/router";
import { assetColumns } from "./asset-columns";
import { useStorageFiles } from "../api/useStorageFiles";
import { EmptyState } from "@/components/ui/empty-state";
import { FileX2, FolderOpen } from "lucide-react";

import { AssetGridView } from "./asset-grid-view";

export function AssetList() {
  const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
  const [showDeleteModal, setShowDeleteModal] = useState<boolean>(false);

  const columns = useMemo<ColumnDef<Asset>[]>(() => assetColumns(), []);

  const [searchParams] = useSearchParams();
  const params = useParams<Params>();

  const view = searchParams?.get("view") ?? "list";
  const assetType = searchParams?.get("type") ?? "all";
  const search = searchParams.get("search");
  const page = Number(searchParams.get("page") ?? "0");

  const { data } = useStorageFiles({
    organizationId: params.organizationId as string,
    search: search ?? undefined,
    type: assetType,
    page: page,
    pageSize: 20,
  });

  if (data && data.files.length === 0) {
    return search ? (
      <EmptyState
        icon={FileX2}
        title="No files match your search"
        description={`Nothing found for “${search}”. Try a different term or clear the filter.`}
      />
    ) : (
      <EmptyState
        icon={FolderOpen}
        title="No files yet"
        description="Upload your first asset and it will show up here."
      />
    );
  }

  return (
    <div>
      <div>
        {data && view === "grid" && (
          <AssetGridView assets={data.files} totalCount={data.totalCount} />
        )}
      </div>
      <div>
        {data && view === "list" && (
          <DataTable
            columns={columns}
            rowSelection={rowSelection}
            setRowSelection={setRowSelection}
            showDeleteModal={showDeleteModal}
            setShowDeleteModal={setShowDeleteModal}
            isLoading={false}
            data={data.files}
            totalCount={data.totalCount}
            onDelete={async () => {}}
          />
        )}
      </div>
    </div>
  );
}
