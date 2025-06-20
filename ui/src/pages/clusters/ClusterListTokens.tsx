import {
  TablePagination,
  useNotify,
  EmptyState,
  Icon,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchTokens } from "api/tokens";
import Loader from "components/Loader";
import React, { FC, useEffect } from "react";
import { isoTimeToString } from "util/helpers";
import { queryKeys } from "util/queryKeys";
import RevokeTokenButton from "./RevokeTokenButton";
import ScrollableTable from "components/ScrollableTable";
import EnrolClusterButton from "pages/clusters/EnrolClusterButton";
import SelectedTableNotification from "components/SelectedTableNotification";
import SelectableMainTable from "components/SelectableMainTable";

type Props = {
  processingNames: string[];
  selectedNames: string[];
  setSelectedNames: (names: string[]) => void;
};

const ClusterListTokens: FC<Props> = ({
  processingNames,
  selectedNames,
  setSelectedNames,
}) => {
  const notify = useNotify();

  const { data: tokens = [], isLoading } = useQuery({
    queryKey: [queryKeys.tokens],
    queryFn: fetchTokens,
  });

  useEffect(() => {
    const validNames = new Set(tokens.map((token) => token.cluster_name));
    const validSelections = selectedNames.filter((name) =>
      validNames.has(name),
    );
    if (validSelections.length !== selectedNames.length) {
      setSelectedNames(validSelections);
    }
  }, [tokens]);

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
      content: "",
      "aria-label": "Actions",
      className: "u-align--right",
    },
  ];

  const tableRows = tokens.map((token) => {
    return {
      key: token.cluster_name,
      name: token.cluster_name,
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
          className: "u-align--right",
        },
      ],
      sortData: {
        clusterName: token.cluster_name,
        expiry: token.expiry,
        createdAt: token.created_at,
      },
    };
  });

  const isEmptyState = !tokens.length && !isLoading;

  return (
    <div
      role="tabpanel"
      aria-labelledby="Tokens"
      className="clusterlist-table-container"
    >
      {isEmptyState ? (
        <div className="u-no-margin--top">
          <EmptyState
            className="empty-state"
            image={<Icon name="cluster-host" className="empty-state-icon" />}
            title="No tokens found"
          >
            <p>There are no join tokens. Enroll a cluster to create one.</p>
            <EnrolClusterButton />
          </EmptyState>
        </div>
      ) : (
        <ScrollableTable
          dependencies={[tokens, notify.notification]}
          tableId="cluster-token-table"
          belowIds={["status-bar"]}
        >
          <TablePagination
            data={tableRows}
            id="pagination"
            itemName="token"
            className="u-no-margin--top"
            aria-label="Table pagination control"
            description={
              selectedNames.length > 0 && (
                <SelectedTableNotification
                  totalCount={tokens.length ?? 0}
                  itemName="token"
                  parentName=""
                  selectedNames={selectedNames}
                  setSelectedNames={setSelectedNames}
                  filteredNames={selectedNames}
                />
              )
            }
          >
            <SelectableMainTable
              id="cluster-token-table"
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
              selectedNames={selectedNames}
              setSelectedNames={setSelectedNames}
              itemName="token"
              parentName=""
              filteredNames={tokens.map((item) => item.cluster_name)}
              disabledNames={processingNames}
            />
          </TablePagination>
        </ScrollableTable>
      )}
    </div>
  );
};

export default ClusterListTokens;
