export const queryKeys = {
  clusters: "clusters",
  links: "links",
  tokens: "tokens",
  server: "server",
  configuration: "configuration",
  cluster: "cluster",
  identities: "identities",
  state: "state",
};

export const allClusterLinksQueryKey = [
  queryKeys.clusters,
  { resource: queryKeys.links },
  "all",
] as const;
