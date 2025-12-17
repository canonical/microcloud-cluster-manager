import type { FC } from "react";
import { memo } from "react";
import { SearchAndFilter } from "@canonical/react-components";
import type {
  SearchAndFilterChip,
  SearchAndFilterData,
} from "@canonical/react-components/dist/components/SearchAndFilter/types";
import { useSearchParams } from "react-router-dom";
import { paramsFromSearchData } from "util/searchAndFilter";
import {
  instanceStatuses,
  nodeStatuses,
  usagePercentiles,
} from "util/clusterFilter";

export const QUERY = "query";
export const NAME = "name";
export const INSTANCE_STATUS = "instance-status";
export const NODE_STATUS = "node-status";
export const MEMORY_USAGE = "memory-usage";
export const POOL_USAGE = "pool-usage";

const QUERY_PARAMS = [
  QUERY,
  NAME,
  INSTANCE_STATUS,
  NODE_STATUS,
  MEMORY_USAGE,
  POOL_USAGE,
];

const ClusterSearchFilter: FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const searchAndFilterData: SearchAndFilterData[] = [
    {
      id: 1,
      heading: "Instance status",
      chips: instanceStatuses.map((status) => {
        return { lead: INSTANCE_STATUS, value: status };
      }),
    },
    {
      id: 2,
      heading: "Node status",
      chips: nodeStatuses.map((status) => {
        return { lead: NODE_STATUS, value: status };
      }),
    },
    {
      id: 3,
      heading: "Memory usage",
      chips: usagePercentiles.map((percentile) => {
        return { lead: MEMORY_USAGE, value: percentile };
      }),
    },
    {
      id: 4,
      heading: "Storage pool usage",
      chips: usagePercentiles.map((percentile) => {
        return { lead: POOL_USAGE, value: percentile };
      }),
    },
  ];

  const onSearchDataChange = (searchData: SearchAndFilterChip[]) => {
    const newParams = paramsFromSearchData(
      searchData,
      searchParams,
      QUERY_PARAMS,
    );

    if (newParams.toString() !== searchParams.toString()) {
      setSearchParams(newParams);
    }
  };

  return (
    <>
      <h2 className="u-off-screen">Search and filter</h2>
      <SearchAndFilter
        filterPanelData={searchAndFilterData}
        returnSearchData={onSearchDataChange}
      />
    </>
  );
};

export default memo(ClusterSearchFilter);
