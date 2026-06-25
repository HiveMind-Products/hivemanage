import { Suspense, useEffect, useState } from "react";
import { DataTableSkeleton } from "@/components/data/DataTableSkeleton";
import { UploadDialog } from "../components/upload-dialog";
import { AssetList } from "../components/asset-list";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useSearchParams } from "react-router";
import { useDebounce } from "@/hooks/use-debounce";
import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { Grid2X2, List, Search } from "lucide-react";
import { usePermission } from "@/features/auth/hooks/use-permission";

import { AssetGridSkeleton } from "../components/asset-grid-skeleton";

export default function StorageRoute() {
  const canWrite = usePermission("storage", "write");
  const [searchParams, setSearchParams] = useSearchParams();
  const [search, setSearch] = useState(searchParams.get("search") ?? "");
  const debouncedSearch = useDebounce(search, 300);

  const type = searchParams.get("type") ?? "all";
  const view = searchParams.get("view") ?? "list";

  useEffect(() => {
    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      if (debouncedSearch) {
        params.set("search", debouncedSearch);
      } else {
        params.delete("search");
      }
      return params;
    }, { replace: true });
  }, [debouncedSearch, setSearchParams]);

  const handleTypeChange = (value: string) => {
    const params = new URLSearchParams(searchParams);
    if (value === "all") {
      params.delete("type");
    } else {
      params.set("type", value);
    }
    setSearchParams(params, { replace: true });
  };

  const handleViewChange = (value: string) => {
    if (!value) return;
    const params = new URLSearchParams(searchParams);
    params.set("view", value);
    setSearchParams(params, { replace: true });
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2 flex-1">
          <div className="relative w-full max-w-2xl">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search files..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="pl-8"
            />
          </div>
          <Select value={type} onValueChange={handleTypeChange}>
            <SelectTrigger className="w-[140px]">
              <SelectValue placeholder="All types" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All types</SelectItem>
              <SelectItem value="image">Images</SelectItem>
              <SelectItem value="video">Videos</SelectItem>
              <SelectItem value="audio">Audio</SelectItem>
            </SelectContent>
          </Select>
          <ToggleGroup
            type="single"
            value={view}
            onValueChange={handleViewChange}
            variant="outline"
          >
            <ToggleGroupItem value="list" aria-label="List view">
              <List className="h-4 w-4" />
            </ToggleGroupItem>
            <ToggleGroupItem value="grid" aria-label="Grid view">
              <Grid2X2 className="h-4 w-4" />
            </ToggleGroupItem>
          </ToggleGroup>
        </div>
        {canWrite && <UploadDialog />}
      </div>

      <div className="mt-4">
        <Suspense
          fallback={
            view === "list" ? (
              <DataTableSkeleton columnCount={10} />
            ) : (
              <AssetGridSkeleton rowCount={10} />
            )
          }
        >
          <AssetList />
        </Suspense>
      </div>
    </div>
  );
}
