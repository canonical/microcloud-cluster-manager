export type LxdConfigPair = Record<string, string | undefined>;

interface ManagerOptions {
  config: {
    [key: string]: string;
  };
}
