import type { FC } from "react";

interface Props {
  name: string;
}

const ClusterLinkName: FC<Props> = ({ name }) => (
  <dl className="cluster-link-name">
    <dt>Link name:</dt>{" "}
    <dd aria-live="polite">
      {name || (
        <span className="u-text--muted">
          Select target cluster to generate link name.
        </span>
      )}
    </dd>
  </dl>
);

export default ClusterLinkName;
