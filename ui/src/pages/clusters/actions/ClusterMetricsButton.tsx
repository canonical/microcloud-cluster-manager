import { Icon } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { fetchConfigurations } from "api/settings";
import classnames from "classnames";

interface Props {
  clusterName: string;
  className?: string;
  onClose?: () => void;
}

const ClusterMetricsButton: FC<Props> = ({
  clusterName,
  className,
  onClose,
}) => {
  const { data: configurations } = useQuery({
    queryKey: [queryKeys.configuration],
    queryFn: fetchConfigurations,
  });

  const baseUrl = configurations?.grafana_base_url?.value;

  if (!baseUrl) {
    return null;
  }

  return (
    <a
      className={classnames("p-button u-no-margin--bottom has-icon", className)}
      onClick={onClose}
      href={`${baseUrl}/lxd?orgId=1&var-job=${clusterName}`}
      target="_blank"
      rel="noopener noreferrer"
    >
      <Icon name="external-link" />
      <span>Metrics</span>
    </a>
  );
};

export default ClusterMetricsButton;
