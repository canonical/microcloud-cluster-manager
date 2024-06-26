import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { fetchSites } from "api/sites";
import { useQuery } from "@tanstack/react-query";
import { Link, MainTable } from "@canonical/react-components";
import DeleteSiteButton from "./DeleteSiteButton";

const SiteList: FC = () => {
  const { data: sites = [] } = useQuery({
    queryKey: [queryKeys.sites],
    queryFn: fetchSites,
  });

  return (
    <div>
      <h1>Sites</h1>
      <MainTable
        headers={[
          { content: "Name" },
          { content: "Status" },
          { content: "JoinedAt" },
          { content: "Instance Count" },
          { content: "Actions" },
        ]}
        rows={(sites || []).map((site) => {
          return {
            columns: [
              { content: site.name },
              { content: site.status },
              { content: site.joined_at },
              { content: site.instance_count },
              {
                content: <DeleteSiteButton siteName={site.name} />,
              },
            ],
          };
        })}
      />

      <Link href="/oidc/logout" className="p-button">
        Logout
      </Link>
    </div>
  );
};

export default SiteList;
