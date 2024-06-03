import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { fetchSites } from "api/sites";
import { useQuery } from "@tanstack/react-query";
import { MainTable } from "@canonical/react-components";

const SiteList: FC = () => {
  const { data: sites = [], error } = useQuery({
    queryKey: [queryKeys.sites],
    queryFn: () => fetchSites(),
  });

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div>
      <h1>Sites</h1>
      <MainTable
        headers={[
          { content: "Name" },
          { content: "Status" },
          { content: "Addresses" },
        ]}
        rows={sites.map((site) => {
          return {
            columns: [
              { content: site.Name },
              { content: site.Status },
              { content: site.Addresses.join(" ") },
            ],
          };
        })}
      />
    </div>
  );
};

export default SiteList;
