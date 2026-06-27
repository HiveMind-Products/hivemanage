import { Params } from "@/typings/router";
import { fetchApi } from "@/utils/http-util";
import { useMutation } from "@tanstack/react-query";
import { useParams } from "react-router";

export function useDeleteToken() {
  const params = useParams<Params>();
  const { mutateAsync, isPending } = useMutation({
    mutationFn: (tokenId: number) =>
      fetchApi(`/api/dash/${params.organizationId}/token/${tokenId}`, {
        method: "DELETE",
      }),
  });
  return { mutateAsync, isPending };
}