import { Button, Card, Col, Icon, Row } from "@canonical/react-components";
import type { FC } from "react";
import type { Cluster, ClusterLink } from "types/cluster";
import { Link } from "react-router-dom";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";
import { getLinkedCluster } from "util/clusterLinks";
import usePanelParams from "context/usePanelParams";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  link: ClusterLink;
}

const ClusterLinkCard: FC<Props> = ({ cluster, clusters, link }) => {
  const panelParams = usePanelParams();
  const targetCluster = getLinkedCluster(link, clusters);

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
            {targetCluster ? (
              <Link to={`/ui/cluster/${targetCluster.name}`}>
                {targetCluster.name}
              </Link>
            ) : (
              "unknown"
            )}
          </p>
        </Col>
        <Col size={4}>
          <p className="u-no-margin--bottom u-text--muted">Addresses</p>
          <p className="u-no-margin--bottom">
            {link.config?.["volatile.addresses"] ?? "unknown"}
          </p>
        </Col>
      </Row>
      <Row>
        <Col size={12}>
          <p className="u-no-margin--bottom u-text--muted">Description</p>
          <p className="u-no-margin--bottom">{link.description || "-"}</p>
        </Col>
      </Row>
      <Row>
        <Col size={12}>
          <Button
            className="u-no-margin--bottom has-icon"
            onClick={() => {
              panelParams.openEditClusterLink(cluster.name, link.name);
            }}
            hasIcon
          >
            <Icon name="edit" />
            <span>Edit</span>
          </Button>
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
