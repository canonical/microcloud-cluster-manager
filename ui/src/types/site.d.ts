export interface SiteMemberStatus {
  address: string;
  architecture: string;
  role: string;
  status: string;
  usage_cpu: number;
  usage_memory: number;
  usage_disk: number;
}

export interface Site {
  name: string;
  status: string;
  site_certificate: string;
  joined_at: string;
  instance_count: number;
  instance_statuses: string;
  created_at: string;
  last_status_update_at: string;
  member_statuses: SiteMemberStatus[];
}
