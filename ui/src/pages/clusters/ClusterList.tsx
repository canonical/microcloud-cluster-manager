import { FC } from "react";
import { List, Row } from "@canonical/react-components";
import BaseLayout from "components/BaseLayout";
import { useParams } from "react-router-dom";
import TabLinks from "components/TabLinks";
import ClusterListTokens from "./ClusterListTokens";
import ClusterListActive from "./ClusterListActive";
import ClusterListPending from "./ClusterListPending";
import AddClusterButton from "./AddClusterButton";
import ClusterStatusGraph from "./metrics/ClusterStatusGraph";
import NotificationRow from "components/NotificationRow";

const ClusterList: FC = () => {
  const { activeTab } = useParams<{
    activeTab?: string;
  }>();

  const tabs: string[] = ["Active", "Pending", "Tokens"];

  return (
    <BaseLayout title={"Clusters"} controls={<AddClusterButton />}>
      <Row>
        <List
          inline={true}
          items={[<ClusterStatusGraph key="clusterstatusgraph" />]}
        />
        <TabLinks tabs={tabs} activeTab={activeTab} tabUrl="/ui/sites" />
        <NotificationRow />
        <div>
          {!activeTab && <ClusterListActive />}
          {activeTab === "pending" && <ClusterListPending />}
          {activeTab === "tokens" && <ClusterListTokens />}
        </div>
      </Row>
    </BaseLayout>
  );
};

export default ClusterList;
