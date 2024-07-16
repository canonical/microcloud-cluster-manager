import { TablePagination, MainTable } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchTokens } from "api/tokens";
import Loader from "components/Loader";
import { FC } from "react";
import { isoTimeToString } from "util/helpers";
import { queryKeys } from "util/queryKeys";

const ClusterListTokens: FC = () => {
  const { data: tokens = [], isLoading } = useQuery({
    queryKey: [queryKeys.tokens],
    queryFn: fetchTokens,
  });

  const tableHeaders = [
    {
      content: "Site name",
    },
    {
      content: "Expiry",
    },
    {
      content: "Created at",
    },
  ];

  const tableRows = tokens.map((token) => {
    return {
      columns: [
        {
          content: `${token.site_name}`,
        },
        {
          content: `${isoTimeToString(token.expiry)}`,
        },
        {
          content: `${isoTimeToString(token.created_at)}`,
        },
      ],
    };
  });

  return (
    <div
      role="tabpanel"
      aria-labelledby="Tokens"
      className="clusterlist-table-container"
    >
      <TablePagination
        data={tableRows}
        id="pagination"
        itemName="token"
        className="u-no-margin--top"
        aria-label="Table pagination control"
      >
        <MainTable
          className={"clusterlist-table"}
          responsive={true}
          emptyStateMsg={
            isLoading ? (
              <Loader text="Loading Tokens..." />
            ) : (
              <>No tokens found.</>
            )
          }
          headers={tableHeaders}
          rows={tableRows}
        />
      </TablePagination>
    </div>
  );
};

export default ClusterListTokens;
