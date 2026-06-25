import { QueryKeys } from "@/typings/query";
import { Token } from "@/typings/token";
import { ApiError, fetchApi } from "@/utils/http-util";
import { useSuspenseQuery } from "@tanstack/react-query";

export function useTokens(organizationId: string) {
  const result = useSuspenseQuery({
    queryKey: [QueryKeys.Tokens, organizationId],
    queryFn: async () => {
      try {
        return fetchApi<Token[]>(`/api/dash/${organizationId}/token`);
      } catch (err) {
        if (err instanceof ApiError) {
          throw new Error(err.message);
        }
      }
    },
  });

  return result.data;
}
