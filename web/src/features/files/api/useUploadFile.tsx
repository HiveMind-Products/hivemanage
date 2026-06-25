import { Params } from "@/typings/router";
import { fetchApi } from "@/utils/http-util";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams } from "react-router";

export function useUploadFile() {
  const params = useParams<Params>();
  const queryClient = useQueryClient();

  const { mutate, mutateAsync, isPending } = useMutation({
    mutationFn: async (file: File) => {
      const formData = new FormData();
      formData.append("file", file);

      return fetchApi<string>(`/api/dash/storage/${params.organizationId}/upload`, {
        method: "POST",
        body: formData,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["storage", params.organizationId] });
    },
  });

  return { mutate, mutateAsync, isPending };
}