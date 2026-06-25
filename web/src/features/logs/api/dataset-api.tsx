import { CreateDatasetSchema, Dataset } from "@/typings/dataset";
import { QueryKeys } from "@/typings/query";
import { ApiError, fetchApi } from "@/utils/http-util";
import {
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query";

export function useListDatasets(organizationId: string | undefined) {
  return useSuspenseQuery({
    queryKey: [QueryKeys.Datasets, organizationId],
    queryFn: async ({ signal }) => {
      try {
        return await fetchApi<Dataset[]>(
          `/api/dash/${organizationId}/dataset`,
          {
            signal,
          },
        );
      } catch (err) {
        if (err instanceof ApiError) {
          throw new Error(err.message);
        }
      }
    },
  });
}

export function useCreateDataset(organizationId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: CreateDatasetSchema) => {
      try {
        return await fetchApi<Dataset>(`/api/dash/${organizationId}/dataset`, {
          method: "POST",
          body: JSON.stringify(data),
        });
      } catch (err) {
        if (err instanceof ApiError) {
          throw new Error(err.message);
        }
      }
    },
    onSuccess: (data) => {
      // not sure if this is the best typing I've done
      queryClient.setQueryData([QueryKeys.Datasets, organizationId], (oldData: Dataset[]) => {
        if (!oldData) return [data];

        return [...oldData, data];
      });
    },
  });
}
