import { Button } from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import type { FC } from "react";
import classnames from "classnames";
import { queryKeys } from "util/queryKeys";
import { fetchConfigurations } from "api/settings";
import { getClusterMetricsUrl } from "util/clusterMetrics";
import DsIcon from "components/DsIcon";

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
  const href = baseUrl && getClusterMetricsUrl(baseUrl, clusterName);
  const isAvailable = !!href;
  let title: string | undefined;

  if (!isAvailable) {
    title =
      "Metrics are unavailable because GRAFANA_BASE_URL is not configured";
  } else if (!hasLabel) {
    title = "Metrics";
  }

  return (
    <Button
      element={isAvailable ? "a" : "button"}
      appearance={hasLabel ? undefined : "base"}
      hasIcon
      className={classnames("u-no-margin--bottom", className)}
      disabled={!isAvailable}
      href={href}
      target={isAvailable ? "_blank" : undefined}
      rel={isAvailable ? "noopener noreferrer" : undefined}
      onClick={isAvailable ? onClose : undefined}
      aria-label={
        hasLabel ? undefined : `Metrics${isAvailable ? "" : " (unavailable)"}`
      }
      title={title}
    >
      <DsIcon icon="statistics" />
      {hasLabel && <span>Metrics</span>}
    </Button>
  );
};

export default ClusterMetricsButton;
