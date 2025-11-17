import { List, Notification, Row, Strip } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchCluster } from "api/clusters";
import BaseLayout from "components/BaseLayout";
import type { FC } from "react";
import { Link, useParams } from "react-router-dom";
import { queryKeys } from "util/queryKeys";
import ClusterDetailInstanceGraph from "./metrics/ClusterDetailInstanceGraph";
import ClusterDetailMemberGraph from "./metrics/ClusterDetailMemberGraph";
import ClusterTimer from "./metrics/ClusterTimer";
import ClusterDetailMetrics from "./metrics/ClusterDetailMetrics";
import BreadCrumbHeader from "components/BreadcrumbHeader";
import { ClusterWarningList } from "pages/clusters/metrics/ClusterWarningList";
import usePanelParams, { panels } from "context/usePanelParams";
import ConfigureClusterPanel from "pages/clusters/ConfigureClusterPanel";
import ClusterUiButton from "pages/clusters/actions/ClusterUiButton";
import ClusterMetricsButton from "pages/clusters/actions/ClusterMetricsButton";
import RemoveClusterButton from "pages/clusters/actions/RemoveClusterButton";
import ConfigureClusterButton from "pages/clusters/actions/ConfigureClusterButton";
import ClusterEditDescriptionBtn from "pages/clusters/actions/ClusterEditDescriptionBtn";

const ClusterDetail: FC = () => {
  const panelParams = usePanelParams();
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
    queryFn: async () => fetchCluster(name),
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
    <>
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
              <ConfigureClusterButton
                cluster={cluster}
                className="p-segmented-control__button"
              />
              <ClusterUiButton
                uiUrl={cluster.ui_url}
                className="p-segmented-control__button"
              />
              <ClusterMetricsButton
                clusterName={cluster.name}
                className="p-segmented-control__button"
              />
              <RemoveClusterButton
                clusterName={cluster.name}
                className="p-segmented-control__button"
              />
            </div>
          </div>
        }
      >
        {!cluster && !error && <>Loading cluster failed</>}
        {error && (
          <Strip>
            <Notification severity="negative" title="Error">
              {error.message}
            </Notification>
          </Strip>
        )}
        {cluster && (
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
            <ClusterWarningList cluster={cluster} />
          </Row>
        )}
      </BaseLayout>
      {panelParams.panel === panels.configureCluster && (
        <ConfigureClusterPanel />
      )}
    </>
  );
};

export default ClusterDetail;
