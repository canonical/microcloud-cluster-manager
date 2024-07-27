export interface Token {
  created_at: string;
  expiry: string;
  cluster_name: string;
}

export interface TokenPostResponse {
  token: string;
}
