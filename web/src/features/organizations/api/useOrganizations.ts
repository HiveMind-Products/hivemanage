import { QueryKeys } from "@/typings/query";
import { Organization } from "@/typings/organizations";
import { fetchApi } from "@/utils/http-util";
import { useQuery } from "@tanstack/react-query";

export function useOrganizations() {
  const { data, isLoading } = useQuery({
    queryKey: [QueryKeys.Organizations],
    queryFn: () => fetchApi<Organization[]>("/api/dash/organization"),
  });

  return { data, isLoading };
}
