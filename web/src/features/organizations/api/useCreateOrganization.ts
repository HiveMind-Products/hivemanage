import { QueryKeys } from "@/typings/query";
import { Organization } from "@/typings/organizations";
import { fetchApi } from "@/utils/http-util";
import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  const { mutate, isPending } = useMutation({
    mutationFn: (params: Omit<Organization, "id">) =>
      fetchApi("/api/dash/organization", {
        method: "POST",
        body: JSON.stringify(params),
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: [QueryKeys.Organizations] });
    },
  });

  return { mutate, isPending };
}
