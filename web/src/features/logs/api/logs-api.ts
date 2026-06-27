import { Field, Log, QueryLogResponse } from "@/typings/logs";
import { ListLogsSchema } from "@/typings/logs";
import { QueryKeys } from "@/typings/query";
import { fetchApi } from "@/utils/http-util";
import { useMutation, useQuery } from "@tanstack/react-query";

export function useListFields(
  organizationId: string | undefined,
  datasetId: string | undefined,
) {
  return useQuery({
    queryKey: [QueryKeys.DatasetFields, organizationId, datasetId],
    queryFn: ({ signal }) =>
      fetchApi<Field[]>(
        `/api/dash/${organizationId}/dataset/${datasetId}/fields`,
        { signal },
      ),
  });
}

export function useQueryLogs(
  organizationId: string | undefined,
  datasetId: string | undefined,
) {
  return useMutation({
    mutationKey: [QueryKeys.Logs, organizationId, datasetId],
    mutationFn: (params: ListLogsSchema) =>
      fetchApi<QueryLogResponse>(
        `/api/dash/${organizationId}/dataset/${datasetId}/logs`,
        {
          method: "POST",
          body: JSON.stringify(params),
        },
      ),
  });
}

export function useLog(
  organizationId: string | undefined,
  datasetId: string | undefined,
  logId: string | null,
) {
  return useQuery({
    queryKey: [QueryKeys.Logs, organizationId, datasetId, logId],
    enabled: !!logId,
    queryFn: ({ signal }) =>
      fetchApi<Log>(
        `/api/dash/${organizationId}/dataset/${datasetId}/logs/${logId}`,
        { signal },
      ),
  });
}
