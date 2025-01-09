import { FC } from "react";
import { Tabs } from "@canonical/react-components";
import { useNotify } from "@canonical/react-components";
import { useNavigate } from "react-router-dom";
import { TabLink } from "@canonical/react-components/dist/components/Tabs/Tabs";

interface Props {
  tabs: (string | TabLink)[];
  activeTab?: string;
  tabUrl: string;
}

const TabLinks: FC<Props> = ({ tabs, activeTab, tabUrl }) => {
  const notify = useNotify();
  const navigate = useNavigate();

  const slugify = (name: string): string => {
    return name.replace(" ", "-").toLowerCase();
  };

  return (
    <Tabs
      links={tabs.map((tab) => {
        if (typeof tab !== "string") {
          return tab;
        }

        const tabPath = slugify(tab);
        const href = tab === tabs[0] ? tabUrl : `${tabUrl}/${tabPath}`;

        return {
          label: tab,
          id: tabPath,
          active: tabPath === activeTab || (tab === tabs[0] && !activeTab),
          onClick: (e) => {
            e.preventDefault();
            notify.clear();
            void navigate(href);
          },
          href,
        };
      })}
    />
  );
};

export default TabLinks;
