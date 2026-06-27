import { fetchApi } from "@/utils/http-util";
import { useQuery } from "@tanstack/react-query";

export interface OrganizationStats {
  totalFiles: number;
  totalSize: number;
  totalLogs: number;
  totalTokens: number;
  totalDataset: number;
}

export function useOrganizationStats(organizationId: string | undefined) {
  return useQuery({
    queryKey: ["organization-stats", organizationId],
    enabled: !!organizationId,
    queryFn: () =>
      fetchApi<OrganizationStats>(
        `/api/dash/organization/${organizationId}/stats`,
      ),
  });
}
