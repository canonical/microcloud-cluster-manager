import { Button, Card, Col, Icon, Row } from "@canonical/react-components";
import type { FC } from "react";
import type { Cluster, ClusterLink } from "types/cluster";
import { Link } from "react-router-dom";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";
import usePanelParams from "context/usePanelParams";
import { getLinkedCluster, isPendingPublicLink } from "util/clusterLinks";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  link: ClusterLink;
}

const ClusterLinkCard: FC<Props> = ({ cluster, clusters, link }) => {
  const panelParams = usePanelParams();
  const targetCluster = getLinkedCluster(link, clusters);
  const pending = isPendingPublicLink(link);
  const targetClusterName =
    targetCluster?.name ?? (pending ? link.name : "unknown");
  const targetClusterURL = targetCluster
    ? `/ui/cluster/${targetCluster.name}`
    : "";

  return (
    <Card title={link.name}>
      <Row>
        <Col size={4}>
          <p className="u-no-margin--bottom u-text--muted">Type</p>
          <p className="u-no-margin--bottom">{link.type}</p>
        </Col>
        <Col size={4}>
          <p className="u-no-margin--bottom u-text--muted">Target cluster</p>
          <p className="u-no-margin--bottom">
            {targetClusterURL ? (
              <Link to={targetClusterURL}>{targetClusterName}</Link>
            ) : (
              targetClusterName
            )}
          </p>
        </Col>
        <Col size={4}>
          <p className="u-no-margin--bottom u-text--muted">Addresses</p>
          <p className="u-no-margin--bottom">
            {link.config?.["volatile.addresses"] ??
              link.config?.["volatile.pending_address"] ??
              "unknown"}
          </p>
        </Col>
      </Row>
      <Row>
        <Col size={12}>
          {pending && (
            <p className="u-text--muted">Pending certificate confirmation</p>
          )}
          {pending && (
            <Button
              appearance=""
              className="u-no-margin--bottom has-icon"
              onClick={() => {
                panelParams.openCreateClusterLink(cluster.name, link.name);
              }}
              hasIcon
            >
              <Icon name="certificate" />
              <span>Verify certificate</span>
            </Button>
          )}
          <DeleteClusterLinkBtn
            cluster={cluster}
            link={link}
            targetCluster={targetCluster}
          />
        </Col>
      </Row>
    </Card>
  );
};

export default ClusterLinkCard;
