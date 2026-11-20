import type { FC } from "react";
import { memo } from "react";
import { SearchAndFilter } from "@canonical/react-components";
import type {
  SearchAndFilterChip,
  SearchAndFilterData,
} from "@canonical/react-components/dist/components/SearchAndFilter/types";
import type { ClusterLinkFilter } from "util/clusterLinks";
import { LINK_TYPES } from "util/clusterLinkTypes";
import { useSearchParams } from "react-router-dom";
import {
  paramsFromSearchData,
  searchParamsToChips,
} from "util/searchAndFilter";

const QUERY = "query";
const LINK_TYPE = "link-type";

const QUERY_PARAMS = [QUERY, LINK_TYPE];

export const getClusterLinkFilters = (
  searchParams: URLSearchParams,
): ClusterLinkFilter => ({
  queries: searchParams.getAll(QUERY),
  types: searchParams.getAll(LINK_TYPE),
});

const searchAndFilterData: SearchAndFilterData[] = [
  {
    id: 1,
    heading: "Link type",
    chips: LINK_TYPES.map(({ value }) => ({ lead: LINK_TYPE, value })),
  },
];

const ClusterLinkSearchFilter: FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();

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
        existingSearchData={searchParamsToChips(searchParams, QUERY_PARAMS)}
        filterPanelData={searchAndFilterData}
        returnSearchData={onSearchDataChange}
      />
    </>
  );
};

export default memo(ClusterLinkSearchFilter);
