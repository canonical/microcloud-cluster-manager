import type { FC } from "react";
import type { Cluster } from "types/cluster";
import { humanFileSize } from "util/helpers";
import { Icon, MainTable, Modal } from "@canonical/react-components";

interface Props {
  cluster: Cluster;
  onClose: () => void;
}

export const ClusterDiskModal: FC<Props> = ({ cluster, onClose }: Props) => {
  const headers = [
    {
      content: "Storage Pool",
      sortKey: "name",
    },
    {
      content: "Location",
      sortKey: "member",
    },
    {
      content: "Percent Used",
      sortKey: "percent_used",
    },
    {
      content: "Total Used",
      sortKey: "total_used",
    },
    {
      content: "Total Size",
      sortKey: "total_size",
    },
  ];

  const poolDetails = cluster.storage_pool_usages.map((pool) => {
    const poolPercent = Math.ceil((100 / pool.total) * pool.usage || 0);
    const isWarning = poolPercent >= cluster.disk_threshold;

    return {
      key: pool.name + pool.member,
      columns: [
        {
          content: (
            <>
              {pool.name}
              {isWarning ? (
                <>
                  {" "}
                  <Icon
                    name="warning"
                    title="High disk usage. Above disk threshold for this cluster"
                  />
                </>
              ) : null}
            </>
          ),
          role: "rowheader",
        },
        {
          content: pool.member || "cluster wide",
          role: "cell",
          sortKey: pool.name,
        },
        {
          content: <>{poolPercent}%</>,
          role: "cell",
        },
        {
          content: humanFileSize(pool.usage),
          role: "cell",
        },
        {
          content: humanFileSize(pool.total),
          role: "cell",
        },
      ],
      sortData: {
        name: pool.name,
        member: pool.member ?? "cluster wide",
        percent_used: poolPercent,
        total_used: pool.usage,
        total_size: pool.total,
      },
    };
  });

  return (
    <Modal close={onClose} title="Storage pool details">
      <MainTable
        headers={headers}
        rows={poolDetails}
        sortable
        defaultSort="name"
        defaultSortDirection="ascending"
      />
    </Modal>
  );
};
