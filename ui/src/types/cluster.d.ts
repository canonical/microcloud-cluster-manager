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
  cluster_uuid: string;
  description: string;
  disk_threshold: number;
  memory_threshold: number;
  ceph_count: number;
  ceph_statuses: StatusDistribution[];
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
  tunnel_registered: boolean;
  lxd_url: string;
}

export type ClusterInstanceStatus = "Running" | "Frozen" | "Error" | "Stopped";

export type ClusterMemberStatus =
  | "Online"
  | "Blocked"
  | "Offline"
  | "Evacuated";

export type ClusterPercentiles = 0.5 | 0.75 | 0.8 | 0.9;

export interface ClusterLink {
  name: string;
  description?: string;
  type: "bidirectional" | "unidirectional" | "public";
  used_by?: string[];
  config?: {
    "volatile.addresses"?: string;
    "volatile.uuid"?: string;
    "volatile.pending_address"?: string;
    "volatile.pending_certificate"?: string;
  };
}

// A cluster link with the name of its source cluster, the managed cluster it
// is defined on, and of its target cluster, the managed cluster it points at.
// The target cluster is null when the link does not point at a managed cluster.
export interface ResolvedClusterLink {
  link: ClusterLink;
  source_cluster: string;
  target_cluster: string | null;
}

export interface ClusterLinkListError {
  cluster: string;
  error: string;
}

export interface ClusterLinkList {
  cluster_links: ResolvedClusterLink[];
  errors: ClusterLinkListError[];
}

export interface ClusterLinkPost {
  target_cluster: string;
  name: string;
  source_auth_groups: string[];
  target_auth_groups?: string[];
  description: string;
  type: ClusterLink["type"];
  remote_address?: string;
}
