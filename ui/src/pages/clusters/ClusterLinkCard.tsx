import { Button, Card, Col, Icon, Row } from "@canonical/react-components";
import type { FC } from "react";
import type { Cluster, ClusterLink } from "types/cluster";
import ClusterLinkRichChip from "pages/clusters/ClusterLinkRichChip";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";
import usePanelParams from "context/usePanelParams";
import { getLinkedCluster } from "util/clusterLinks";
import ClusterRichChip from "./ClusterRichChip";

interface Props {
  cluster: Cluster;
  clusters: Cluster[];
  link: ClusterLink;
}

const ClusterLinkCard: FC<Props> = ({ cluster, clusters, link }) => {
  const panelParams = usePanelParams();
  const targetCluster = getLinkedCluster(link, clusters);
  const targetClusterName = targetCluster?.name ?? "unknown";
  const targetClusterURL = targetCluster
    ? `/ui/cluster/${targetCluster.name}`
    : "";

  return (
    <Card
      title={
        <ClusterLinkRichChip
          cluster={cluster.name}
          clusterLinkName={link.name}
          clusterLinkType={link.type}
        />
      }
    >
      <Row>
        <Col size={4}>
          <p className="u-no-margin--bottom u-text--muted">Type</p>
          <p className="u-no-margin--bottom">{link.type}</p>
        </Col>
        <Col size={4}>
          <p className="u-no-margin--bottom u-text--muted">Target cluster</p>
          <p className="u-no-margin--bottom">
            {targetClusterURL ? (
              <ClusterRichChip clusterName={targetClusterName} />
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
