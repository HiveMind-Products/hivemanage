import { Organization } from "@/typings/organizations";
import { ApiError, fetchApi } from "@/utils/http-util";
import { useMutation, useQueryClient } from "@tanstack/react-query";

export function useCreateOrganization() {
  const queryClient = useQueryClient();
  const { mutate, isPending } = useMutation({
    mutationFn: async (params: Omit<Organization, "id">) => {
      try {
        return await fetchApi("/api/dash/organization", {
          method: "POST",
          body: JSON.stringify(params),
        });
      } catch (err) {
        if (err instanceof ApiError) {
          throw new Error(err.message);
        }
      }
    },
    onSuccess: () => { queryClient.invalidateQueries({ queryKey: ["organizations"] }); },
  });

  return { mutate, isPending };
}
