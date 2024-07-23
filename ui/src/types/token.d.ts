export interface Token {
  created_at: string;
  expiry: string;
  site_name: string;
}

export interface TokenPostResponse {
  token: string;
}
