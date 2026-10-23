import { Notification, Row, Spinner, Strip } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import { fetchCluster } from "api/clusters";
import BaseLayout from "components/BaseLayout";
import type { FC } from "react";
import { Link, useParams } from "react-router-dom";
import { queryKeys } from "util/queryKeys";
import BreadCrumbHeader from "components/BreadcrumbHeader";
import usePanelParams, { panels } from "context/usePanelParams";
import ConfigureClusterPanel from "pages/clusters/ConfigureClusterPanel";
import ClusterUiButton from "pages/clusters/actions/ClusterUiButton";
import ClusterMetricsButton from "pages/clusters/actions/ClusterMetricsButton";
import RemoveClusterButton from "pages/clusters/actions/RemoveClusterButton";
import ConfigureClusterButton from "pages/clusters/actions/ConfigureClusterButton";
import ClusterOverview from "pages/clusters/ClusterOverview";
import ClusterLinks from "pages/clusters/ClusterLinks";
import TabLinks from "components/TabLinks";
import CreateClusterLinkPanel from "pages/clusters/CreateClusterLinkPanel";

const ClusterDetail: FC = () => {
  const panelParams = usePanelParams();
  const { name, activeTab } = useParams<{ name: string; activeTab?: string }>();
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
    return <Spinner className="u-loader" text="Loading..." isMainComponent />;
  }

  return (
    <>
      <BaseLayout
        title={
          <BreadCrumbHeader
            name={name}
            parentItems={[
              <Link to="/ui/clusters" key="clusters">
                Clusters
              </Link>,
            ]}
          />
        }
        controls={
          cluster && (
            <div className="p-segmented-control">
              <div className="p-segmented-control__list">
                <ConfigureClusterButton
                  cluster={cluster}
                  className="p-segmented-control__button"
                />
                <ClusterUiButton
                  lxdUrl={cluster.lxd_url}
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
          )
        }
      >
        {!cluster && !error && (
          <Strip>
            <Notification severity="negative" title="Error">
              Unable to get cluster details.
            </Notification>
          </Strip>
        )}
        {error && (
          <Strip>
            <Notification severity="negative" title="Error">
              {error.message}
            </Notification>
          </Strip>
        )}
        {cluster && (
          <Row>
            <TabLinks
              tabs={["Overview", "Cluster links"]}
              activeTab={activeTab}
              tabUrl={`/ui/cluster/${cluster.name}`}
            />
            {activeTab === "cluster-links" ? (
              <ClusterLinks cluster={cluster} />
            ) : (
              <ClusterOverview cluster={cluster} />
            )}
          </Row>
        )}
      </BaseLayout>
      {panelParams.panel === panels.configureCluster && (
        <ConfigureClusterPanel />
      )}
      {(panelParams.panel === panels.createClusterLink ||
        panelParams.panel === panels.editClusterLink) && (
        <CreateClusterLinkPanel />
      )}
    </>
  );
};

export default ClusterDetail;
