export type ConfigPair = Record<string, string | undefined>;

interface ManagerOptions {
  config: {
    [key: string]: string;
  };
}

interface MemberOptions {
  target?: string;
  config: {
    https_address: string;
    external_address?: string;
  };
}
