import { FC } from "react";
import { queryKeys } from "util/queryKeys";
import { fetchSites } from "api/sites";
import { useQuery } from "@tanstack/react-query";
import { Button, MainTable } from "@canonical/react-components";
import { useNavigate } from "react-router-dom";

const SiteList: FC = () => {
  const navigate = useNavigate();
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
          { content: "Addresses" },
        ]}
        rows={sites.map((site) => {
          return {
            columns: [
              { content: site.name },
              { content: site.status },
              { content: site.addresses.join(" ") },
            ],
          };
        })}
      />

      <Button onClick={() => navigate("/oidc/logout")}>Logout</Button>
    </div>
  );
};

export default SiteList;
