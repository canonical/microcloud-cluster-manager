import { useState, type FC } from "react";
import {
  ActionButton,
  Button,
  CheckboxInput,
  failure,
  Form,
  Input,
  Notification,
  Row,
  ScrollableContainer,
  SidePanel,
  Spinner,
  useNotify,
  useToastNotification,
  type NotificationType,
} from "@canonical/react-components";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchClusterLinks,
  fetchClusters,
  updateClusterLink,
} from "api/clusters";
import NotificationRow from "components/NotificationRow";
import ResourceIcon from "components/ResourceIcon";
import { useCluster } from "context/useClusters";
import { useClusterLinks } from "context/useClusterLinks";
import { useClusterIdentities } from "context/useClusterIdentities";
import usePanelParams from "context/usePanelParams";
import { useFormik } from "formik";
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";
import type { ClusterLink } from "types/cluster";
import {
  getLinkedCluster,
  getLinkIdentity,
  getReverseLink,
  invalidateClusterLinkLists,
  isBidirectional,
} from "util/clusterLinks";
import { queryKeys } from "util/queryKeys";
import PanelSection from "components/PanelSection";
import ClusterLinkName from "pages/clusters/ClusterLinkName";
import ClusterLinkRichChip from "pages/clusters/ClusterLinkRichChip";

interface EditClusterLinkFormValues {
  description: string;
  applyDescriptionToBoth: boolean;
}

const EditClusterLinkPanel: FC = () => {
  const panelParams = usePanelParams();
  const notify = useNotify();
  const toastNotify = useToastNotification();
  const [error, setError] = useState<NotificationType | null>(null);
  const queryClient = useQueryClient();

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
  const { data: identities = [] } = useClusterIdentities(clusterName);

  const link = clusterLinks.find((item) => item.name === linkName);
  const targetCluster = link ? getLinkedCluster(link, clusters) : undefined;
  const linkType = LINK_TYPES.find((type) => type.value === link?.type);
  const identity = getLinkIdentity(identities, panelParams.clusterLink);

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
    onSubmit: async (values) => {
      setError(null);
      const attemptedClusters: string[] = [];
      let sourceUpdated = false;
      const updateBoth =
        link && isBidirectional(link) && values.applyDescriptionToBoth;

      try {
        if (!link || !cluster) {
          throw new Error("Cluster link not found");
        }
        if (isBidirectional(link) && !identity) {
          throw new Error("Identity not found");
        }

        let reverseLink: ClusterLink | undefined;
        if (updateBoth) {
          if (!targetCluster?.tunnel_registered) {
            throw new Error(
              "The target cluster must be managed and have a registered tunnel to update both ends.",
            );
          }
          const targetLinks = await fetchClusterLinks(targetCluster.name);
          reverseLink = getReverseLink(targetLinks, cluster);
          if (!reverseLink) {
            throw new Error(
              `No cluster link pointing back at ${cluster.name} found on cluster ${targetCluster.name}.`,
            );
          }
        }

        attemptedClusters.push(clusterName);
        await updateClusterLink(
          clusterName,
          linkName,
          JSON.stringify({
            description: values.description,
            config: link.config,
          }),
        );
        sourceUpdated = true;
        toastNotify.success(
          <>
            Cluster link{" "}
            <ClusterLinkRichChip
              clusterLinkName={link.name}
              clusterLinkType={link.type}
              cluster={clusterName}
            />{" "}
            on <b>{clusterName}</b> saved.
          </>,
        );

        if (reverseLink && targetCluster) {
          attemptedClusters.push(targetCluster.name);
          await updateClusterLink(
            targetCluster.name,
            reverseLink.name,
            JSON.stringify({
              description: values.description,
              config: reverseLink.config,
            }),
          );
          toastNotify.success(
            <>
              Cluster link <b>{reverseLink.name}</b> on{" "}
              <b>{targetCluster.name}</b> saved.
            </>,
          );
          if (!toastNotify.isListView) {
            toastNotify.toggleListView();
          }
        }
        closePanel();
      } catch (e) {
        setError(
          failure(
            sourceUpdated
              ? "Source description saved, but target description update failed"
              : "Cluster link update failed",
            e,
          ),
        );
      } finally {
        void invalidateClusterLinkLists(queryClient, attemptedClusters);
      }
    },
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
        <Row className="u-no-padding">
          {error && (
            <Notification
              title={error.title}
              severity="negative"
              onDismiss={() => {
                setError(null);
              }}
            >
              {error.message}
            </Notification>
          )}
        </Row>
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
      {error && (
        <Row className="u-no-padding">
          <Notification
            title={error.title}
            severity="negative"
            onDismiss={() => {
              setError(null);
            }}
          >
            {error.message}
          </Notification>
        </Row>
      )}
      <SidePanel.Content className="u-no-padding edit-cluster-link-panel">
        <ScrollableContainer
          dependencies={[notify.notification, error]}
          belowIds={["panel-footer"]}
        >
          <Form
            onSubmit={(e) => {
              e.preventDefault();
              void formik.submitForm();
            }}
            className="form"
          >
            <PanelSection header="Link type" isFirst>
              <div className="cluster-link-type-summary">
                <p className="link-type-label">
                  {linkType?.label ?? link.type}
                </p>
                {linkType && (
                  <p className="p-form-help-text">{linkType.description}</p>
                )}
              </div>
            </PanelSection>
            <PanelSection header="Details">
              <div className="cluster-link-endpoints">
                <Input
                  type="text"
                  label="Source"
                  value={cluster.name}
                  readOnly
                />
                <span
                  className="cluster-link-endpoints__arrow"
                  aria-hidden="true"
                >
                  <ResourceIcon type={getLinkIconType(link.type)} />
                </span>
                <Input
                  type="text"
                  label="Target"
                  value={targetCluster?.name ?? ""}
                  readOnly
                />
              </div>
              <ClusterLinkName name={link.name} />
              <Input
                id="description"
                type="text"
                label="Description"
                className="u-no-margin--bottom"
                {...formik.getFieldProps("description")}
              />
              {isBidirectional(link) && (
                <CheckboxInput
                  className="cluster-link-address-checkbox"
                  label="Apply change in both source and target clusters"
                  checked={formik.values.applyDescriptionToBoth}
                  onChange={() => {
                    void formik.setFieldValue(
                      "applyDescriptionToBoth",
                      !formik.values.applyDescriptionToBoth,
                    );
                  }}
                />
              )}
            </PanelSection>
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
          loading={formik.isSubmitting}
          disabled={!formik.isValid || formik.isSubmitting}
          onClick={() => void formik.submitForm()}
        >
          Save
        </ActionButton>
      </SidePanel.Footer>
    </SidePanel>
  );
};

export default EditClusterLinkPanel;
