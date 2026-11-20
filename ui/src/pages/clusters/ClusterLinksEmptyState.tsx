import { EmptyState, Icon } from "@canonical/react-components";
import type { FC } from "react";
import CustomIcon from "components/CustomIcon";
import CreateClusterLinkButton from "components/cluster-links/CreateClusterLinkButton";

interface Props {
  cluster: string;
}

const CLUSTER_LINKS_DOC_URL =
  "https://documentation.ubuntu.com/lxd/latest/howto/cluster_links_create/";

const ClusterLinksEmptyState: FC<Props> = ({ cluster }) => {
  return (
    <div className="cluster-links-empty-state">
      <EmptyState
        className="empty-state"
        image={<CustomIcon name="empty" className="empty-state-icon" />}
        title="No cluster links found"
      >
        <p>There are no cluster links on this cluster.</p>
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
        <CreateClusterLinkButton cluster={cluster} />
      </EmptyState>
    </div>
  );
};

export default ClusterLinksEmptyState;
