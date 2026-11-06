import { useQuery } from "@tanstack/react-query";
import type { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { fetchConfigurations } from "api/settings";
import { getClusterMetricsUrl } from "util/clusterMetrics";
import ExternalLinkButton from "components/ExternalLinkButton";

interface Props {
  clusterName: string;
  className?: string;
  onClose?: () => void;
  hasLabel?: boolean;
}

const ClusterMetricsButton: FC<Props> = ({
  clusterName,
  className,
  onClose,
  hasLabel = true,
}) => {
  const { data: configurations } = useQuery({
    queryKey: [queryKeys.configuration],
    queryFn: fetchConfigurations,
  });

  const baseUrl = configurations?.grafana_base_url?.value;

  return (
    <ExternalLinkButton
      href={baseUrl ? getClusterMetricsUrl(baseUrl, clusterName) : undefined}
      label="Metrics"
      iconName="statistics"
      unavailableReason="Metrics are unavailable because GRAFANA_BASE_URL is not configured"
      className={className}
      onClose={onClose}
      hasLabel={hasLabel}
    />
  );
};

export default ClusterMetricsButton;
