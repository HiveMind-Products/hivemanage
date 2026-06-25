export interface Asset {
  id: string;
  key: string;
  originalName?: string;
  size: number;
  type: string;
}

export interface AssetResponse {
  files: Asset[];
  totalCount: number;
}

export interface AssetURLResponse {
  url: string;
}
