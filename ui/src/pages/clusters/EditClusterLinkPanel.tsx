import {
  ActionButton,
  Button,
  CheckboxInput,
  Form,
  Input,
  Notification,
  ScrollableContainer,
  SidePanel,
  Spinner,
  useNotify,
} from "@canonical/react-components";
import { useQuery } from "@tanstack/react-query";
import type { FC } from "react";
import { useFormik } from "formik";
import NotificationRow from "components/NotificationRow";
import usePanelParams from "context/usePanelParams";
import { useCluster, useClusterLinks } from "context/useClusters";
import { fetchClusters } from "api/clusters";
import { queryKeys } from "util/queryKeys";
import { getLinkedCluster } from "util/clusterLinks";
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";
import ReadOnlyInput from "components/forms/ReadOnlyInput";
import ResourceIcon from "components/ResourceIcon";

interface EditClusterLinkFormValues {
  description: string;
  applyDescriptionToBoth: boolean;
}

const EditClusterLinkPanel: FC = () => {
  const panelParams = usePanelParams();
  const notify = useNotify();

  const clusterName = panelParams.cluster ?? "";
  const linkName = panelParams.clusterLink ?? "";

  const { data: cluster } = useCluster(clusterName);
  const {
    data: clusterLinks = [],
    isLoading: isLinksLoading,
    error: linksError,
    refetch: refetchLinks,
  } = useClusterLinks(clusterName);
  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const link = clusterLinks.find((item) => item.name === linkName);
  const targetCluster = link ? getLinkedCluster(link, clusters) : undefined;
  const linkType = LINK_TYPES.find((type) => type.value === link?.type);

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  const formik = useFormik<EditClusterLinkFormValues>({
    initialValues: {
      description: link?.description ?? "",
      applyDescriptionToBoth: false,
    },
    enableReinitialize: true,
    // saving edits needs a backend update endpoint
    onSubmit: () => undefined,
  });

  if (!cluster) {
    return null;
  }

  if (!link) {
    return (
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>Edit cluster link</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <SidePanel.Content>
          {isLinksLoading ? (
            <Spinner text="Loading cluster link..." />
          ) : linksError ? (
            <Notification
              severity="negative"
              title="Failed to load cluster links"
              actions={[
                {
                  label: "Retry",
                  onClick: () => void refetchLinks(),
                },
              ]}
            >
              {linksError.message}
            </Notification>
          ) : (
            <Notification severity="caution" title="Cluster link not found">
              {`"${linkName}" no longer exists on ${cluster.name}.`}
            </Notification>
          )}
        </SidePanel.Content>
        <SidePanel.Footer className="u-align--right">
          <Button className="u-no-margin--bottom" onClick={closePanel}>
            Close
          </Button>
        </SidePanel.Footer>
      </SidePanel>
    );
  }

  return (
    <SidePanel>
      <SidePanel.Header>
        <SidePanel.HeaderTitle>Edit cluster link</SidePanel.HeaderTitle>
      </SidePanel.Header>
      <NotificationRow className="u-no-padding" />
      <SidePanel.Content className="u-no-padding">
        <ScrollableContainer
          dependencies={[notify.notification]}
          belowIds={["panel-footer"]}
        >
          {/* submission is disabled until the update endpoint exists */}
          <Form
            onSubmit={(e) => {
              e.preventDefault();
            }}
            className="form"
          >
            <h2 className="p-muted-heading">Link type</h2>
            <div className="cluster-link-types">
              <p>{linkType?.label ?? link.type}</p>
              {linkType && (
                <p className="p-form-help-text cluster-link-types__help">
                  {linkType.help}
                </p>
              )}
            </div>
            <hr />
            <h2 className="p-muted-heading">Details</h2>
            <div className="cluster-link-endpoints">
              <ReadOnlyInput
                id="source-cluster"
                label="Source"
                value={cluster.name}
              />
              <span
                className="cluster-link-endpoints__arrow"
                aria-hidden="true"
              >
                <ResourceIcon type={getLinkIconType(link.type)} />
              </span>
              <ReadOnlyInput
                id="target-cluster"
                label="Target"
                value={targetCluster?.name ?? ""}
              />
            </div>
            <ReadOnlyInput id="name" label="Link name" value={link.name} />
            <Input
              id="description"
              type="text"
              label="Description"
              {...formik.getFieldProps("description")}
            />
            <CheckboxInput
              label="Apply change in both source and target clusters"
              checked={formik.values.applyDescriptionToBoth}
              onChange={() => {
                void formik.setFieldValue(
                  "applyDescriptionToBoth",
                  !formik.values.applyDescriptionToBoth,
                );
              }}
            />
          </Form>
        </ScrollableContainer>
      </SidePanel.Content>
      <SidePanel.Footer className="u-align--right">
        <Button
          appearance="base"
          className="u-no-margin--bottom"
          onClick={closePanel}
        >
          Cancel
        </Button>
        <ActionButton
          appearance="positive"
          className="u-no-margin--bottom"
          disabled
        >
          Save changes
        </ActionButton>
      </SidePanel.Footer>
    </SidePanel>
  );
};

export default EditClusterLinkPanel;
