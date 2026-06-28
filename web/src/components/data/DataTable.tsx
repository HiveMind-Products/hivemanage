import {
  type ColumnDef,
  type PaginationState,
  type Row,
  type RowSelectionState,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { useEffect, useRef, useState } from "react";

import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useSearchParams } from "react-router";
import { Asset } from "@/typings/asset";
//import { DataDeleteDialog } from "./DataDeleteDialog";

interface DataTableProps<T = Asset> {
  data: T[];
  columns: ColumnDef<T>[];
  totalCount: number;
  isLoading: boolean;
  onDelete: (rows: Row<T>[]) => Promise<void>;
  rowSelection: RowSelectionState;
  setRowSelection: React.Dispatch<React.SetStateAction<RowSelectionState>>;
  showDeleteModal: boolean;
  setShowDeleteModal: React.Dispatch<React.SetStateAction<boolean>>;
}

export function DataTable<T extends Asset>({
  data,
  totalCount,
  columns,
  rowSelection,
  setRowSelection,
}: DataTableProps<T>) {
  const [searchParams, setSearchParams] = useSearchParams();

  const [pagination, setPagination] = useState<PaginationState>({
    pageSize: 20,
    pageIndex: Number(searchParams.get("page")) || 0,
  });

  const tableContainerRef = useRef<HTMLDivElement>(null);

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    onPaginationChange: setPagination,
    onRowSelectionChange: setRowSelection,
    pageCount: Math.ceil(totalCount / 20),
    manualPagination: true,
    state: {
      pagination,
      rowSelection: rowSelection,
    },
  });

  /*const handleDeleteSelection = async () => {
    const rows = table.getFilteredSelectedRowModel().rows;
    await onDelete(rows);
  }; */

  useEffect(() => {
    setSearchParams((current) => {
      const params = new URLSearchParams(current);
      params.set("page", `${pagination.pageIndex}`);
      return params;
    }, { replace: true });
  }, [pagination.pageIndex, setSearchParams]);

  return (
    <div>
      {/*table.getFilteredSelectedRowModel().rows.length ? (
        <DataDeleteDialog<T>
          selectedRows={table.getFilteredSelectedRowModel().rows}
          showModal={showDeleteModal}
          setShowDeleteModal={setShowDeleteModal}
          onDelete={handleDeleteSelection}
          isLoading={isLoading}
        />
      ) : null */}
      <div ref={tableContainerRef} className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader className="bg-muted/40">
            {table.getHeaderGroups()?.map((headerGroup) => (
              <TableRow key={headerGroup.id} className="hover:bg-transparent">
                {headerGroup.headers?.map((header) => {
                  return (
                    <TableHead
                      key={header.id}
                      colSpan={header.colSpan}
                      className="h-9 text-xs font-medium uppercase tracking-wide text-muted-foreground"
                    >
                      {header.isPlaceholder
                        ? null
                        : flexRender(
                            header.column.columnDef.header,
                            header.getContext(),
                          )}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => {
                return (
                  <TableRow key={row.id} className="border-border/60">
                    {row.getVisibleCells().map((cell) => {
                      return (
                        <TableCell key={cell.id} className="py-2.5">
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext(),
                          )}
                        </TableCell>
                      );
                    })}
                  </TableRow>
                );
              })
            ) : (
              <TableRow>
                <TableCell
                  colSpan={columns.length}
                  className="h-24 text-center text-muted-foreground"
                >
                  No results.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <div className="flex items-center justify-between gap-2 py-4">
        <p className="text-sm text-muted-foreground tabular-nums">
          {table.getFilteredSelectedRowModel().rows.length > 0
            ? `${table.getFilteredSelectedRowModel().rows.length} selected`
            : `${totalCount.toLocaleString()} item${totalCount === 1 ? "" : "s"}`}
        </p>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
          >
            Previous
          </Button>
          <span className="px-1 text-sm tabular-nums text-muted-foreground">
            {pagination.pageIndex + 1} / {table.getPageCount() || 1}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
          >
            Next
          </Button>
        </div>
      </div>
    </div>
  );
}
