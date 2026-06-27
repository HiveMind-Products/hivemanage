import { CreateDatasetSchema, Dataset } from "@/typings/dataset";
import { QueryKeys } from "@/typings/query";
import { fetchApi } from "@/utils/http-util";
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";

export function useListDatasets(organizationId: string | undefined) {
  return useSuspenseQuery({
    queryKey: [QueryKeys.Datasets, organizationId],
    queryFn: ({ signal }) =>
      fetchApi<Dataset[]>(`/api/dash/${organizationId}/dataset`, { signal }),
  });
}

export function useCreateDataset(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateDatasetSchema) =>
      fetchApi<Dataset>(`/api/dash/${organizationId}/dataset`, {
        method: "POST",
        body: JSON.stringify(data),
      }),
    onSuccess: (data) => {
      // not sure if this is the best typing I've done
      queryClient.setQueryData([QueryKeys.Datasets, organizationId], (oldData: Dataset[]) => {
        if (!oldData) return [data];

        return [...oldData, data];
      });
    },
  });
}
