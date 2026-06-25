import { Params } from "@/typings/router";
import { ApiError, fetchApi } from "@/utils/http-util";
import { useMutation } from "@tanstack/react-query";
import { useParams } from "react-router";

export function useDeleteToken() {
  const params = useParams<Params>();
  const { mutateAsync, isPending } = useMutation({
    mutationFn: async (tokenId: number) => {
      try {
        return await fetchApi(`/api/dash/${params.organizationId}/token/${tokenId}`, {
          method: "DELETE",
        });
      } catch (error) {
        if (error instanceof ApiError) throw new Error(error.message);
        throw error;
      }
    },
  });
  return { mutateAsync, isPending };
}