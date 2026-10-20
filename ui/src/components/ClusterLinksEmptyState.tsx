import { EmptyState, Icon } from "@canonical/react-components";
import type { FC } from "react";
import emptyIcon from "icons/empty.svg";
import type { Cluster } from "types/cluster";
import CreateClusterLinkBtn from "pages/clusters/actions/CreateClusterLinkBtn";

const CLUSTER_LINKS_DOC_URL =
  "https://documentation.ubuntu.com/lxd/latest/howto/cluster_links_create/";

interface Props {
  cluster: Cluster;
}

const ClusterLinksEmptyState: FC<Props> = ({ cluster }) => {
  return (
    <div className="cluster-links-empty-state">
      <EmptyState
        className="empty-state"
        image={<img src={emptyIcon} alt="" className="empty-state-icon" />}
        title="No cluster links found"
      >
        <p>There are no cluster links on this server.</p>
        <p>
          <a
            href={CLUSTER_LINKS_DOC_URL}
            target="_blank"
            rel="noopener noreferrer"
          >
            Learn more about cluster links
            <Icon
              aria-hidden="true"
              className="external-link-icon"
              name="external-link"
            />
          </a>
        </p>
        <CreateClusterLinkBtn cluster={cluster} />
      </EmptyState>
    </div>
  );
};

export default ClusterLinksEmptyState;
