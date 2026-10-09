import { type FC, type OptionHTMLAttributes } from "react";
import { Select, Spinner, useNotify } from "@canonical/react-components";
import type { Props as SelectProps } from "@canonical/react-components/dist/components/Select/Select";
import { useQuery } from "@tanstack/react-query";
import { queryKeys } from "util/queryKeys";
import { fetchClusters } from "api/clusters";

interface Props {
  value: string;
  setValue: (value: string) => void;
  selectProps?: SelectProps;
  ignoreOptions?: string[];
}

const ClusterSelector: FC<Props> = ({
  value,
  setValue,
  ignoreOptions = [],
  selectProps,
}) => {
  const notify = useNotify();
  const {
    data: clusters = [],
    error,
    isLoading,
  } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  if (isLoading) {
    return <Spinner className="u-loader" text="Loading clusters..." />;
  }

  if (error) {
    notify.failure("Loading clusters failed", error);
  }

  const clustersWithTunnel = clusters.filter(
    (cluster) =>
      cluster.tunnel_registered && !ignoreOptions.includes(cluster.name),
  );

  const getClusterOptions = () => {
    const options: OptionHTMLAttributes<HTMLOptionElement>[] = [];
    if (clustersWithTunnel) {
      clustersWithTunnel.forEach((cluster) => {
        options.push({
          label: cluster.name,
          value: cluster.name,
        });
      });
      if (clustersWithTunnel.length === 0) {
        options.unshift({
          label: "No clusters available",
          value: "",
          disabled: true,
        });
      } else {
        options.unshift({
          label: "Select a cluster",
          value: "",
          disabled: true,
        });
      }
    }

    return options;
  };

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setValue(e.target.value);
  };

  return (
    <Select
      name="pool"
      aria-label="Target cluster"
      {...selectProps}
      options={getClusterOptions()}
      onChange={handleChange}
      value={value}
    />
  );
};

export default ClusterSelector;
