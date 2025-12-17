export interface StatusDistribution {
  status: string;
  count: number;
}

export interface StoragePoolUsage {
  name: string;
  member?: string;
  total: number;
  usage: number;
}

export interface Cluster {
  name: string;
  description: string;
  disk_threshold: number;
  memory_threshold: number;
  cpu_total_count: number;
  cpu_load_1: string;
  cpu_load_5: string;
  cpu_load_15: string;
  created_at: string;
  instance_count: number;
  instance_statuses: StatusDistribution[];
  joined_at: string;
  last_status_update_at: string;
  member_count: number;
  member_statuses: StatusDistribution[];
  memory_total_amount: number;
  memory_usage: number;
  storage_pool_usages: StoragePoolUsage[];
  cluster_certificate: string;
  status: string;
  ui_url: string;
}

export type ClusterInstanceStatus = "Running" | "Frozen" | "Error" | "Stopped";

export type ClusterNodeStatus = "Online" | "Blocked" | "Offline" | "Evacuated";

export type ClusterPercentiles = 0.5 | 0.75 | 0.8 | 0.9;
