import { QueryKeys } from "@/typings/query";
import { Organization } from "@/typings/organizations";
import { fetchApi } from "@/utils/http-util";
import { useQuery } from "@tanstack/react-query";

export function useCurrentOrganization(organizationId: string | undefined) {
  const { data, isLoading } = useQuery({
    queryKey: [QueryKeys.Organization, organizationId],
    queryFn: () =>
      fetchApi<Organization>(`/api/dash/organization/${organizationId}`),
  });

  return { data, isLoading };
}
