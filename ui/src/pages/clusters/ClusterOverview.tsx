import { List, Notification } from "@canonical/react-components";
import type { FC } from "react";
import type { Cluster } from "types/cluster";
import ClusterDetailInstanceGraph from "./metrics/ClusterDetailInstanceGraph";
import ClusterDetailMemberGraph from "./metrics/ClusterDetailMemberGraph";
import ClusterTimer from "./metrics/ClusterTimer";
import ClusterDetailMetrics from "./metrics/ClusterDetailMetrics";
import { ClusterWarningList } from "pages/clusters/metrics/ClusterWarningList";
import ClusterEditDescriptionBtn from "pages/clusters/actions/ClusterEditDescriptionBtn";

interface Props {
  cluster: Cluster;
}

const ClusterOverview: FC<Props> = ({ cluster }) => {
  return (
    <>
      <List
        className="cluster-detail-graphs"
        inline
        items={[
          <ClusterTimer cluster={cluster} key="cluster-timer" />,
          <ClusterDetailMetrics
            cluster={cluster}
            key="cluster-detail-metrics"
          />,
          <ClusterDetailMemberGraph
            cluster={cluster}
            key="cluster-member-graph"
          />,
          <ClusterDetailInstanceGraph
            cluster={cluster}
            key="cluster-instance-graph"
          />,
        ]}
      />
      {cluster.description ? (
        <>
          <h2 className="p-heading--4">Description</h2>
          <p>
            {cluster.description}{" "}
            <ClusterEditDescriptionBtn
              appearance="base"
              clusterName={cluster.name}
            />
          </p>
        </>
      ) : (
        <div>
          <ClusterEditDescriptionBtn
            appearance=""
            label="Add description"
            clusterName={cluster.name}
          />
        </div>
      )}
      <Notification
        severity="information"
        title={
          cluster.tunnel_registered
            ? "Tunnel registered"
            : "Tunnel not available"
        }
      />
      <ClusterWarningList cluster={cluster} />
    </>
  );
};

export default ClusterOverview;
