import type { FC } from "react";
import { memo } from "react";
import { SearchAndFilter } from "@canonical/react-components";
import type {
  SearchAndFilterChip,
  SearchAndFilterData,
} from "@canonical/react-components/dist/components/SearchAndFilter/types";
import { useSearchParams } from "react-router-dom";
import { paramsFromSearchData } from "util/searchAndFilter";

export const QUERY = "query";
export const LINK_TYPE = "link-type";

const QUERY_PARAMS = [QUERY, LINK_TYPE];

export const linkTypes = ["bidirectional", "unidirectional", "public"];

const ClusterLinkSearchFilter: FC = () => {
  const [searchParams, setSearchParams] = useSearchParams();

  const searchAndFilterData: SearchAndFilterData[] = [
    {
      id: 1,
      heading: "Link type",
      chips: linkTypes.map((type) => {
        return { lead: LINK_TYPE, value: type };
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

export default memo(ClusterLinkSearchFilter);
