import { Notification, List, Row, Strip } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchCluster } from "api/clusters";
import BaseLayout from "components/BaseLayout";
import React, { FC } from "react";
import { Link, useParams } from "react-router-dom";
import { queryKeys } from "util/queryKeys";
import ClusterDetailInstanceGraph from "./metrics/ClusterDetailInstanceGraph";
import ClusterDetailMemberGraph from "./metrics/ClusterDetailMemberGraph";
import ClusterTimer from "./metrics/ClusterTimer";
import ClusterDetailMetrics from "./metrics/ClusterDetailMetrics";
import Loader from "components/Loader";
import BreadCrumbHeader from "components/BreadcrumbHeader";
import DeleteClusterButton from "pages/clusters/DeleteClusterButton";
import ClusterMetricsButton from "pages/clusters/ClusterMetricsButton";
import ClusterUiButton from "pages/clusters/ClusterUiButton";
import { ClusterWarningList } from "pages/clusters/metrics/ClusterWarningList";

const ClusterDetail: FC = () => {
  const { name } = useParams<{ name: string }>();
  if (!name) {
    return <>Missing name</>;
  }

  const {
    data: cluster,
    error,
    isLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters, name],
    queryFn: () => fetchCluster(name),
    refetchInterval: 60000,
    refetchIntervalInBackground: true,
  });

  if (isLoading) {
    return <></>;
  }

  if (!cluster) {
    return <>Unable to get details</>;
  }

  return (
    <BaseLayout
      title={
        <BreadCrumbHeader
          name={cluster.name}
          parentItems={[
            <Link to="/ui/clusters" key="clusters">
              Clusters
            </Link>,
          ]}
        />
      }
      controls={
        <div className="p-segmented-control">
          <div className="p-segmented-control__list">
            <ClusterUiButton uiUrl={cluster.ui_url} />
            <ClusterMetricsButton clusterName={cluster.name} />
            <DeleteClusterButton clusterName={cluster.name} />
          </div>
        </div>
      }
    >
      {isLoading && <Loader text="Loading cluster details..." />}
      {!isLoading && !cluster && !error && <>Loading cluster failed</>}
      {error && (
        <Strip>
          <Notification severity="negative" title="Error">
            {error.message}
          </Notification>
        </Strip>
      )}
      {!isLoading && cluster && (
        <Row>
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
                key="cluster-node-graph"
              />,
              <ClusterDetailInstanceGraph
                cluster={cluster}
                key="cluster-instance-graph"
              />,
            ]}
          />
          <ClusterWarningList cluster={cluster} />
        </Row>
      )}
    </BaseLayout>
  );
};

export default ClusterDetail;
