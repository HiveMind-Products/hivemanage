"use client";

import { Sheet, SheetContent } from "@/components/ui/sheet";
import { useState } from "react";
import { Prism as SyntaxHighlighter } from "react-syntax-highlighter";
import { materialDark } from "react-syntax-highlighter/dist/esm/styles/prism";

import { Button } from "@/components/ui/button";
import {
  CustomTabs,
  CustomTabsContent,
  CustomTabsList,
  CustomTabsTrigger,
} from "@/components/ui/custom-tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";
import dayjs from "dayjs";
import timezone from "dayjs/plugin/timezone";
import utc from "dayjs/plugin/utc";
import {
  Check,
  ChevronDown,
  ChevronUp,
  ChevronsUpDown,
  Clipboard,
} from "lucide-react";
import { useProfileTimezone } from "@/hooks/use-profile-timezone";
import { LOG_LEVELS } from "../../utils/constants";
import { useDataTable } from "../../providers/logs-table-provider";
import { useParams, useSearchParams } from "react-router";
import { useLog } from "../../api/logs-api";

dayjs.extend(utc);
dayjs.extend(timezone);

type Params = {
  organizationId: string;
  datasetId: string;
};

type SortDirection = "asc" | "desc" | null;

function setNestedValue(
  obj: Record<string, unknown>,
  keys: string[],
  value: unknown,
): void {
  let current = obj;

  for (let i = 0; i < keys.length - 1; i++) {
    const key = keys[i]!;
    const nextKey = keys[i + 1]!;

    if (current[key] === undefined) {
      current[key] = /^\d+$/.test(nextKey) ? [] : {};
    }

    current = current[key] as Record<string, unknown>;
  }

  const lastKey = keys[keys.length - 1]!;
  current[lastKey] = value;
}

function convertArraysRecursively(obj: unknown): unknown {
  if (obj === null || typeof obj !== "object") {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(convertArraysRecursively);
  }

  const record = obj as Record<string, unknown>;
  const keys = Object.keys(record);

  const isArrayLike = keys.length > 0 && keys.every((key) => /^\d+$/.test(key));

  if (isArrayLike) {
    const maxIndex = Math.max(...keys.map(Number));
    const arr: unknown[] = [];
    for (let i = 0; i <= maxIndex; i++) {
      arr[i] = convertArraysRecursively(record[String(i)]);
    }
    return arr;
  }

  const result: Record<string, unknown> = {};
  for (const key of keys) {
    result[key] = convertArraysRecursively(record[key]);
  }
  return result;
}

function sortObjectKeysRecursively(obj: unknown): unknown {
  if (obj === null || typeof obj !== "object") {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(sortObjectKeysRecursively);
  }

  const sorted: Record<string, unknown> = {};
  const keys = Object.keys(obj as Record<string, unknown>).sort((a, b) =>
    a.localeCompare(b),
  );

  for (const key of keys) {
    sorted[key] = sortObjectKeysRecursively(
      (obj as Record<string, unknown>)[key],
    );
  }

  return sorted;
}

function formatRawJson(metadata: Record<string, string>) {
  const entries = Object.entries(metadata).filter(
    ([key]) => !key.startsWith("_"),
  );

  const result: Record<string, unknown> = {};

  for (const [key, value] of entries) {
    if (key.includes(".")) {
      const keys = key.split(".");
      setNestedValue(result, keys, value);
    } else {
      result[key] = value;
    }
  }

  const formatted = convertArraysRecursively(result);
  const sorted = sortObjectKeysRecursively(formatted);

  return JSON.stringify(sorted, null, 2);
}

export function LogDataSheet() {
  const [sortDirection, setSortDirection] = useState<SortDirection>(null);
  const [profileTz] = useProfileTimezone();
  const [isCopied, setIsCopied] = useState(false);

  const { rowSelection, table } = useDataTable();
  const [searchParams, setSearchParams] = useSearchParams();
  const { organizationId, datasetId } = useParams<Params>();

  const traceId = searchParams.get("logId");
  const selectedRowKey = Object.keys(rowSelection)?.[0];

  function deleteTraceParam() {
    const params = new URLSearchParams(searchParams);
    params.delete("logId");

    setSearchParams(params);
  }

  const { data: log } = useLog(organizationId, datasetId, traceId);

  if (!log) {
    return null;
  }

  const userTimezone = profileTz ?? "Etc/UTC";
  const timestamp = dayjs
    .utc(log.Timestamp)
    .tz(userTimezone)
    .format("MMM DD, HH:mm:ss");

  const handleSort = () => {
    if (sortDirection === null) {
      setSortDirection("asc");
    } else if (sortDirection === "asc") {
      setSortDirection("desc");
    } else {
      setSortDirection(null);
    }
  };

  const getSortIcon = () => {
    if (sortDirection === "asc") {
      return <ChevronUp className="h-3 w-3" />;
    } else if (sortDirection === "desc") {
      return <ChevronDown className="h-3 w-3" />;
    } else {
      return <ChevronsUpDown className="h-3 w-3" />;
    }
  };

  const sortedMetadata = () => {
    const entries = Object.entries(log?.Metadata ?? {});

    if (sortDirection === null) {
      return entries;
    }

    return entries.sort(([keyA], [keyB]) => {
      const comparison = keyA.localeCompare(keyB);
      return sortDirection === "asc" ? comparison : -comparison;
    });
  };

  return (
    <Sheet
      open={!!selectedRowKey || !!traceId}
      onOpenChange={() => {
        const el = selectedRowKey
          ? document.getElementById(selectedRowKey)
          : null;
        deleteTraceParam();
        table.resetRowSelection();
        setTimeout(() => el?.focus(), 0);
      }}
    >
      <SheetContent
        style={{ minWidth: "70%" }}
        className="h-full flex-1 overflow-auto p-2"
      >
        <div className="h-full w-full overflow-auto">
          <div className="mt-2 flex items-center space-x-4">
            <div>
              <p className="text-lg font-medium">{timestamp}</p>
            </div>

            <div>
              <span
                className={cn(
                  "text-md inline-flex items-center rounded-md px-2 py-1 font-medium ring-1 ring-inset",
                  LOG_LEVELS[log.Metadata.severity] ??
                    "bg-muted text-muted-foreground ring-border",
                )}
              >
                {log?.Metadata.severity}
              </span>
            </div>
          </div>

          <div className="mt-4 rounded-md border bg-accent">
            <div className="p-4">
              <p className="break-words text-sm text-foreground">{log?.Body}</p>
            </div>
          </div>

          <div className="mt-8">
            <CustomTabs defaultValue="table" className="w-full">
              <CustomTabsList className="w-full justify-start border-b rounded-none border-accent pl-0">
                <CustomTabsTrigger
                  value="table"
                  className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary data-[state=active]:shadow-none px-4 pb-3 pt-2"
                >
                  Table
                </CustomTabsTrigger>
                <CustomTabsTrigger
                  value="json"
                  className="rounded-none border-b-2 data-[state=active]:border-primary data-[state=active]:shadow-none px-4 pb-3 pt-2"
                >
                  Raw JSON
                </CustomTabsTrigger>
              </CustomTabsList>
              <CustomTabsContent value="table" className="mt-4">
                <div className="rounded-md border">
                  <Table className="">
                    <TableHeader className="h-4 bg-accent">
                      <TableRow className="">
                        <TableHead
                          className="h-10 rounded-tl-md font-normal text-xs cursor-pointer hover:bg-accent/80 transition-colors"
                          onClick={handleSort}
                        >
                          <div className="flex items-center gap-1">
                            FIELD
                            {getSortIcon()}
                          </div>
                        </TableHead>
                        <TableHead className="h-10 rounded-tr-md border-l font-normal text-xs">
                          VALUE
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {sortedMetadata().map(([key, value]) => (
                        <TableRow key={key} className="hover:bg-accent text-sm">
                          <TableCell>{key}</TableCell>
                          <TableCell className="break-all border-l whitespace-pre-line">
                            {value.toString()}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </CustomTabsContent>
              <CustomTabsContent value="json" className="mt-4">
                <div className="relative rounded-md border bg-muted/50 text-xs font-mono whitespace-pre-wrap break-all">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="absolute top-2 right-2 h-7 w-7 z-10"
                    onClick={() => {
                      navigator.clipboard.writeText(
                        JSON.stringify(log.Metadata, null, 2),
                      );
                      setIsCopied(true);
                      setTimeout(() => setIsCopied(false), 2000);
                    }}
                  >
                    {isCopied ? (
                      <Check className="h-4 w-4 text-foreground" />
                    ) : (
                      <Clipboard className="h-4 w-4 text-foreground" />
                    )}
                  </Button>
                  <SyntaxHighlighter
                    customStyle={{ margin: 0 }}
                    language="json"
                    style={materialDark}
                    wrapLines
                    wrapLongLines
                  >
                    {formatRawJson(
                      log.Metadata as unknown as Record<string, string>,
                    )}
                  </SyntaxHighlighter>
                </div>
              </CustomTabsContent>
            </CustomTabs>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  );
}
