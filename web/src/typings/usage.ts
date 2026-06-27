export interface StorageCategory {
  category: string;
  count: number;
  size: number;
}

export interface LargestFile {
  id: string;
  name: string;
  type: string;
  size: number;
}

export interface DatasetLogCount {
  datasetId: string;
  name: string;
  count: number;
}

export interface LogDayCount {
  date: string;
  count: number;
}

export interface OrganizationUsage {
  totalFiles: number;
  totalSize: number;
  storageByCategory: StorageCategory[];
  largestFiles: LargestFile[];
  totalLogs: number;
  logsByDataset: DatasetLogCount[];
  logsTimeseries: LogDayCount[];
  loggingAvailable: boolean;
}
