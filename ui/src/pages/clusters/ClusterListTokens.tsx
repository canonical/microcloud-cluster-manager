import { TablePagination, MainTable } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchTokens } from "api/tokens";
import Loader from "components/Loader";
import { FC } from "react";
import { isoTimeToString } from "util/helpers";
import { queryKeys } from "util/queryKeys";
import RevokeTokenButton from "./RevokeTokenButton";

const ClusterListTokens: FC = () => {
  const { data: tokens = [], isLoading } = useQuery({
    queryKey: [queryKeys.tokens],
    queryFn: fetchTokens,
  });

  const tableHeaders = [
    {
      content: "Cluster name",
      sortKey: "clusterName",
    },
    {
      content: "Expiry",
      sortKey: "expiry",
    },
    {
      content: "Created at",
      sortKey: "createdAt",
    },
    {
      content: "Actions",
    },
  ];

  const tableRows = tokens.map((token) => {
    return {
      columns: [
        {
          content: `${token.cluster_name}`,
        },
        {
          content: `${isoTimeToString(token.expiry)}`,
        },
        {
          content: `${isoTimeToString(token.created_at)}`,
        },
        {
          content: <RevokeTokenButton token={token} />,
        },
      ],
      sortData: {
        clusterName: token.cluster_name,
        expiry: token.expiry,
        createdAt: token.created_at,
      },
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
          sortable
          headers={tableHeaders}
          rows={tableRows}
          defaultSort="createdAt"
        />
      </TablePagination>
    </div>
  );
};

export default ClusterListTokens;
