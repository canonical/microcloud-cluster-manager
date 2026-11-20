import { useEffect, useState, type FC, type ReactNode } from "react";
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
import { fetchIdentities, updateIdentity } from "api/auth-identities";
import {
  fetchClusterCertificateFingerprint,
  fetchClusterLinks,
  fetchClusters,
  updateClusterLink,
} from "api/clusters";
import NotificationRow from "components/NotificationRow";
import PanelSection from "components/PanelSection";
import ResourceIcon from "components/ResourceIcon";
import { useCluster, useClusterLinks } from "context/useClusters";
import { useAuthGroups } from "context/useAuthGroups";
import usePanelParams from "context/usePanelParams";
import { useFormik } from "formik";
import ClusterLinkAuthGroupsSection from "pages/clusters/ClusterLinkAuthGroupsSection";
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";
import type { Cluster, ClusterLink } from "types/cluster";
import {
  getLinkedCluster,
  findClusterLinkIdentity,
  getLinkIdentity,
  getReverseLink,
  isBidirectional,
  isUnidirectional,
} from "util/clusterLinks";
import { areSameArrays } from "util/helpers";
import { queryKeys } from "util/queryKeys";

interface EditClusterLinkFormValues {
  description: string;
  applyDescriptionToBoth: boolean;
  sourceAuthGroups?: string[];
  targetAuthGroups?: string[];
}

interface Props {
  cluster: Cluster;
  link: ClusterLink;
  clusters: Cluster[];
}

const EditClusterLinkForm: FC<Props> = ({ cluster, link, clusters }) => {
  const panelParams = usePanelParams();
  const notify = useNotify();
  const toastNotify = useToastNotification();
  const [error, setError] = useState<NotificationType | null>(null);
  const [editedAuthGroups, setEditedAuthGroups] = useState({
    source: false,
    target: false,
  });
  const queryClient = useQueryClient();

  const clusterName = cluster.name;
  const linkName = link.name;
  const targetCluster = getLinkedCluster(link, clusters);
  const linkType = LINK_TYPES.find((type) => type.value === link.type);
  const bidirectional = isBidirectional(link);
  const hasAuthGroups = bidirectional || isUnidirectional(link);
  const targetClusterName = targetCluster?.name ?? "";
  const canLoadTarget = hasAuthGroups && !!targetCluster?.tunnel_registered;

  const sourceIdentityKey = [
    queryKeys.identities,
    clusterName,
    "cluster-link",
    linkName,
  ];
  const sourceIdentityQuery = useQuery({
    queryKey: sourceIdentityKey,
    queryFn: async () => {
      const identity = getLinkIdentity(
        await fetchIdentities(clusterName),
        linkName,
      );
      if (
        !identity ||
        identity.authentication_method !== "tls" ||
        identity.type !== "Cluster link certificate"
      ) {
        throw new Error(
          `No active TLS identity found for cluster link "${linkName}" on ${clusterName}.`,
        );
      }
      return identity;
    },
    enabled: bidirectional,
  });
  const targetIdentityKey = [
    queryKeys.identities,
    targetClusterName,
    "cluster-link-peer",
    clusterName,
  ];
  const targetIdentityQuery = useQuery({
    queryKey: targetIdentityKey,
    queryFn: async () => {
      const [identities, fingerprint] = await Promise.all([
        fetchIdentities(targetClusterName),
        fetchClusterCertificateFingerprint(clusterName),
      ]);
      const identity = findClusterLinkIdentity(identities, fingerprint);
      if (!identity) {
        throw new Error(
          `No active TLS identity for ${clusterName} found on ${targetClusterName}.`,
        );
      }
      return identity;
    },
    enabled: canLoadTarget,
  });
  const sourceAuthGroupsQuery = useAuthGroups(targetClusterName, canLoadTarget);
  const targetAuthGroupsQuery = useAuthGroups(clusterName, bidirectional);

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  const formik = useFormik<EditClusterLinkFormValues>({
    initialValues: {
      description: link.description ?? "",
      applyDescriptionToBoth: false,
      sourceAuthGroups: targetIdentityQuery.data
        ? (targetIdentityQuery.data.groups ?? [])
        : undefined,
      targetAuthGroups: sourceIdentityQuery.data
        ? (sourceIdentityQuery.data.groups ?? [])
        : undefined,
    },
    onSubmit: async (values) => {
      setError(null);
      const savedUpdates: string[] = [];
      let operation = "cluster link update";
      const updateBoth = hasAuthGroups && values.applyDescriptionToBoth;
      const sourceGroupsChanged =
        hasAuthGroups &&
        editedAuthGroups.source &&
        values.sourceAuthGroups !== undefined &&
        !areSameArrays(
          values.sourceAuthGroups,
          targetIdentityQuery.data?.groups ?? [],
        );
      const targetGroupsChanged =
        bidirectional &&
        editedAuthGroups.target &&
        values.targetAuthGroups !== undefined &&
        !areSameArrays(
          values.targetAuthGroups,
          sourceIdentityQuery.data?.groups ?? [],
        );
      const recordSuccess = (label: string, message: ReactNode): void => {
        savedUpdates.push(label);
        toastNotify.success(message);
        if (savedUpdates.length === 2 && !toastNotify.isListView) {
          toastNotify.toggleListView();
        }
      };

      try {
        if (
          sourceGroupsChanged &&
          (!canLoadTarget ||
            !targetIdentityQuery.isSuccess ||
            !sourceAuthGroupsQuery.isSuccess)
        ) {
          throw new Error(
            "Load the target identity and auth groups before changing source permissions.",
          );
        }
        if (
          targetGroupsChanged &&
          (!sourceIdentityQuery.isSuccess || !targetAuthGroupsQuery.isSuccess)
        ) {
          throw new Error(
            "Load the source identity and auth groups before changing target permissions.",
          );
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

        operation = "source description update";
        await updateClusterLink(
          clusterName,
          linkName,
          JSON.stringify({
            description: values.description,
            config: link.config,
          }),
        );
        // TODO: replace by ClusterLinkRichChip
        recordSuccess(
          "Source description",
          <>
            Cluster link <b>{link.name}</b> on <b>{clusterName}</b> saved.
          </>,
        );

        if (reverseLink && targetCluster) {
          operation = "target description update";
          await updateClusterLink(
            targetCluster.name,
            reverseLink.name,
            JSON.stringify({
              description: values.description,
              config: reverseLink.config,
            }),
          );
          recordSuccess(
            "Target description",
            <>
              Cluster link <b>{reverseLink.name}</b> on{" "}
              <b>{targetCluster.name}</b> saved.
            </>,
          );
        }

        if (
          sourceGroupsChanged &&
          targetIdentityQuery.data &&
          values.sourceAuthGroups !== undefined
        ) {
          operation = "source auth groups update";
          const identity = {
            ...targetIdentityQuery.data,
            groups: values.sourceAuthGroups,
          };
          await updateIdentity(targetClusterName, identity);
          queryClient.setQueryData(targetIdentityKey, identity);
          recordSuccess(
            "Source auth groups",
            <>
              Source auth groups for cluster link <b>{linkName}</b> on{" "}
              <b>{targetClusterName}</b> saved.
            </>,
          );
        }
        if (
          targetGroupsChanged &&
          sourceIdentityQuery.data &&
          values.targetAuthGroups !== undefined
        ) {
          operation = "target auth groups update";
          const identity = {
            ...sourceIdentityQuery.data,
            groups: values.targetAuthGroups,
          };
          await updateIdentity(clusterName, identity);
          queryClient.setQueryData(sourceIdentityKey, identity);
          recordSuccess(
            "Target auth groups",
            <>
              Target auth groups for cluster link <b>{linkName}</b> on{" "}
              <b>{clusterName}</b> saved.
            </>,
          );
        }
        closePanel();
      } catch (e) {
        setError(
          failure(
            savedUpdates.length
              ? `${savedUpdates.join(", ")} saved, but ${operation} failed`
              : "Cluster link update failed",
            e,
          ),
        );
      } finally {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters, clusterName, queryKeys.links],
        });
        if (updateBoth && targetCluster) {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.clusters, targetCluster.name, queryKeys.links],
          });
        }
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.identities, clusterName],
        });
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.authGroups, clusterName],
        });
        if (hasAuthGroups && targetClusterName) {
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.identities, targetClusterName],
          });
          void queryClient.invalidateQueries({
            queryKey: [queryKeys.authGroups, targetClusterName],
          });
        }
      }
    },
  });

  useEffect(() => {
    if (
      formik.values.sourceAuthGroups === undefined &&
      targetIdentityQuery.isSuccess
    ) {
      void formik.setFieldValue(
        "sourceAuthGroups",
        targetIdentityQuery.data.groups ?? [],
        false,
      );
    }
    if (
      formik.values.targetAuthGroups === undefined &&
      sourceIdentityQuery.isSuccess
    ) {
      void formik.setFieldValue(
        "targetAuthGroups",
        sourceIdentityQuery.data.groups ?? [],
        false,
      );
    }
  }, [
    formik.values.sourceAuthGroups,
    formik.values.targetAuthGroups,
    formik.setFieldValue,
    sourceIdentityQuery.isSuccess,
    sourceIdentityQuery.data,
    targetIdentityQuery.isSuccess,
    targetIdentityQuery.data,
  ]);

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
          dependencies={[
            notify.notification,
            error,
            sourceIdentityQuery.data,
            targetIdentityQuery.data,
            sourceAuthGroupsQuery.data,
            targetAuthGroupsQuery.data,
            formik.values.sourceAuthGroups,
            formik.values.targetAuthGroups,
          ]}
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
                  <p className="p-form-help-text">{linkType.help}</p>
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
              <Input type="text" label="Link name" value={link.name} readOnly />
              <Input
                id="description"
                type="text"
                label="Description"
                className="u-no-margin--bottom"
                disabled={formik.isSubmitting}
                {...formik.getFieldProps("description")}
              />
              {isBidirectional(link) && (
                <CheckboxInput
                  className="cluster-link-address-checkbox"
                  label="Apply change in both source and target clusters"
                  checked={formik.values.applyDescriptionToBoth}
                  disabled={formik.isSubmitting}
                  onChange={() => {
                    void formik.setFieldValue(
                      "applyDescriptionToBoth",
                      !formik.values.applyDescriptionToBoth,
                    );
                  }}
                />
              )}
            </PanelSection>
            {hasAuthGroups && (
              <ClusterLinkAuthGroupsSection
                source={{
                  clusterName: targetClusterName,
                  query: sourceAuthGroupsQuery,
                  identityQuery: targetIdentityQuery,
                  unavailableReason: canLoadTarget
                    ? undefined
                    : "The target cluster must be managed and have a registered tunnel to edit source auth groups.",
                  selectedGroups:
                    formik.values.sourceAuthGroups ??
                    targetIdentityQuery.data?.groups ??
                    [],
                  onChange: (groups) => {
                    setEditedAuthGroups((current) => ({
                      ...current,
                      source: true,
                    }));
                    void formik.setFieldValue("sourceAuthGroups", groups);
                  },
                }}
                disabled={formik.isSubmitting}
                {...(bidirectional
                  ? {
                      isBidirectional: true,
                      target: {
                        clusterName,
                        query: targetAuthGroupsQuery,
                        identityQuery: sourceIdentityQuery,
                        selectedGroups:
                          formik.values.targetAuthGroups ??
                          sourceIdentityQuery.data?.groups ??
                          [],
                        onChange: (groups: string[]) => {
                          setEditedAuthGroups((current) => ({
                            ...current,
                            target: true,
                          }));
                          void formik.setFieldValue("targetAuthGroups", groups);
                        },
                      },
                    }
                  : { isBidirectional: false })}
              />
            )}
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
  } = useClusterLinks(clusterName);
  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });
  const link = clusterLinks.find((item) => item.name === linkName);
  const closePanel = (): void => {
    panelParams.clear();
    notify.clear();
  };

  if (!cluster) {
    return null;
  }

  if (link) {
    return (
      <EditClusterLinkForm
        key={`${clusterName}/${linkName}`}
        cluster={cluster}
        link={link}
        clusters={clusters}
      />
    );
  }

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
};

export default EditClusterLinkPanel;
