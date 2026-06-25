import { Asset, AssetURLResponse } from "@/typings/asset";
import { fetchApi } from "@/utils/http-util";
import { useQuery } from "@tanstack/react-query";

export function useFile(organizationId: string | undefined, fileId: string | undefined) {
  const { data, isPending } = useQuery({ queryKey: ["file", organizationId, fileId], enabled: !!organizationId && !!fileId, queryFn: async () => fetchApi<Asset>(`/api/dash/storage/${organizationId}/file/${fileId}`) });
  return { data, isPending };
}

export function useFileURL(organizationId: string | undefined, fileId: string | undefined) {
  return useQuery({ queryKey: ["file-url", organizationId, fileId], enabled: !!organizationId && !!fileId, queryFn: async () => fetchApi<AssetURLResponse>(`/api/dash/storage/${organizationId}/file/${fileId}/url`) });
}
