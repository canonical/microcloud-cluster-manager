import type { FC } from "react";
import {
  EmptyState,
  Icon,
  ScrollableTable,
  Spinner,
  TablePagination,
  useNotify,
} from "@canonical/react-components";
import { Link } from "react-router-dom";
import { ClusterInstances } from "./metrics/ClusterInstances";
import { ClusterMembers } from "./metrics/ClusterMembers";
import ClusterHeartbeat from "./metrics/ClusterHeartbeat";
import ClusterStatus from "./metrics/ClusterStatus";
import type { Cluster } from "types/cluster";
import { ClusterWarningCount } from "pages/clusters/metrics/ClusterWarningCount";
import SelectableMainTable from "components/SelectableMainTable";
import SelectedTableNotification from "components/SelectedTableNotification";
import ClusterActions from "pages/clusters/ClusterActions";
import EnrollClusterButton from "pages/clusters/actions/EnrollClusterButton";
import usePanelParams from "context/usePanelParams";

interface Props {
  clusters: Cluster[];
  isEmptyState: boolean;
  isLoading: boolean;
  processingNames: string[];
  selectedNames: string[];
  setSelectedNames: (names: string[]) => void;
}

const ClusterListActive: FC<Props> = ({
  clusters,
  isEmptyState,
  isLoading,
  processingNames,
  selectedNames,
  setSelectedNames,
}) => {
  const notify = useNotify();
  const panelParams = usePanelParams();

  const tableHeaders = [
    {
      content: "Cluster Name",
      sortKey: "name",
      className: "name",
    },
    {
      content: "Description",
      sortKey: "description",
      className: "description",
    },
    {
      content: "Status",
      sortKey: "status",
      className: "status",
    },
    {
      content: "Members",
      className: "members",
    },
    {
      content: "Running instances",
      className: "instances",
    },
    {
      content: "Warnings",
      className: "warnings",
    },
    {
      content: "",
      "aria-label": "Actions",
      className: "actions",
    },
  ];

  const tableRows = clusters.map((cluster) => {
    return {
      key: cluster.name,
      name: cluster.name,
      columns: [
        {
          content: (
            <Link to={`/ui/cluster/${cluster.name}`}>{cluster.name}</Link>
          ),
          className: "name",
          role: "rowheader",
        },
        {
          content: cluster.description,
          className: "description",
          role: "cell",
        },
        {
          content: (
            <>
              <ClusterStatus cluster={cluster} />
              <ClusterHeartbeat cluster={cluster} />
            </>
          ),
          className: "status",
          role: "cell",
        },
        {
          content: <ClusterMembers cluster={cluster} />,
          className: "members",
          role: "cell",
        },
        {
          content: <ClusterInstances cluster={cluster} />,
          className: "instances",
          role: "cell",
        },
        {
          content: <ClusterWarningCount cluster={cluster} />,
          className: "warnings",
          role: "cell",
        },
        {
          content: <ClusterActions cluster={cluster} />,
          className: "actions",
          role: "cell",
        },
      ],
      sortData: {
        name: cluster.name,
        description: cluster.description.toLowerCase(),
        status: cluster.last_status_update_at,
        lastHeartbeat: cluster.last_status_update_at,
      },
    };
  });

  return (
    <div
      role="tabpanel"
      aria-labelledby="Active"
      className="clusterlist-table-container"
    >
      {isEmptyState ? (
        <div className="u-no-margin--top">
          <EmptyState
            className="empty-state"
            image={<Icon name="cluster-host" className="empty-state-icon" />}
            title="No active clusters found"
          >
            <p>There are no active clusters. Enroll your first cluster!</p>
            <EnrollClusterButton />
          </EmptyState>
        </div>
      ) : (
        <ScrollableTable
          dependencies={[clusters, notify.notification]}
          tableId="clusterlist-table"
          belowIds={["status-bar"]}
        >
          <TablePagination
            data={tableRows}
            id="pagination"
            itemName=" active cluster"
            className="u-no-margin--top"
            aria-label="Table pagination control"
            description={
              selectedNames.length > 0 && (
                <SelectedTableNotification
                  totalCount={clusters.length ?? 0}
                  itemName="cluster"
                  parentName=""
                  selectedNames={selectedNames}
                  setSelectedNames={setSelectedNames}
                  filteredNames={selectedNames}
                />
              )
            }
          >
            <SelectableMainTable
              id="clusterlist-table"
              className="clusterlist-table"
              responsive
              emptyStateMsg={
                isLoading ? (
                  <Spinner className="u-loader" text="Loading Clusters..." />
                ) : (
                  <>No clusters found matching this search.</>
                )
              }
              headers={tableHeaders}
              rows={tableRows}
              sortable
              selectedNames={selectedNames}
              setSelectedNames={setSelectedNames}
              itemName="cluster"
              parentName=""
              filteredNames={clusters.map((item) => item.name)}
              disabledNames={processingNames}
              disableSelect={!!panelParams.panel}
            />
          </TablePagination>
        </ScrollableTable>
      )}
    </div>
  );
};

export default ClusterListActive;
