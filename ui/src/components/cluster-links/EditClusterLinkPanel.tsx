import { useEffect, useRef, useState, type FC, type ReactNode } from "react";
import {
  ActionButton,
  Button,
  failure,
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
import { useAuthGroups } from "context/useAuthGroups";
import { useClusterLinks } from "context/useClusterLinks";
import { useCluster } from "context/useClusters";
import usePanelParams from "context/usePanelParams";
import { queryKeys } from "util/queryKeys";
import ClusterLinkForm from "components/forms/ClusterLinkForm";
import NotificationRow from "components/NotificationRow";
import ClusterLinkAuthGroupsSection from "pages/clusters/ClusterLinkAuthGroupsSection";
import ClusterLinkRichChip from "pages/clusters/ClusterLinkRichChip";
import { useFormik } from "formik";
import type { ClusterLink } from "types/cluster";
import {
  findClusterLinkIdentity,
  getLinkedCluster,
  getLinkIdentity,
  getReverseLink,
  invalidateClusterLinkLists,
  isBidirectional,
  isUnidirectional,
} from "util/clusterLinks";
import { areSameArrays } from "util/helpers";

interface EditClusterLinkFormValues {
  description: string;
  applyDescriptionToBoth: boolean;
  sourceAuthGroups?: string[];
  targetAuthGroups?: string[];
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
  } = useClusterLinks(clusterName);
  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });
  const link = clusterLinks.find((item) => item.name === linkName);
  const toastNotify = useToastNotification();
  const [error, setError] = useState<NotificationType | null>(null);
  const [editedAuthGroups, setEditedAuthGroups] = useState({
    source: false,
    target: false,
  });
  const queryClient = useQueryClient();
  const initializedLink = useRef<string | null>(null);
  const targetCluster = link ? getLinkedCluster(link, clusters) : undefined;
  const bidirectional = !!cluster && !!link && isBidirectional(link);
  const hasAuthGroups =
    !!cluster && !!link && (bidirectional || isUnidirectional(link));
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

  const sourceAuthGroupsQuery = useAuthGroups(clusterName, bidirectional);
  const targetAuthGroupsQuery = useAuthGroups(targetClusterName, canLoadTarget);

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  const formik = useFormik<EditClusterLinkFormValues>({
    initialValues: {
      description: link?.description ?? "",
      applyDescriptionToBoth: false,
      sourceAuthGroups: sourceIdentityQuery.data
        ? (sourceIdentityQuery.data.groups ?? [])
        : undefined,
      targetAuthGroups: targetIdentityQuery.data
        ? (targetIdentityQuery.data.groups ?? [])
        : undefined,
    },
    onSubmit: async (values) => {
      setError(null);
      const attemptedClusters: string[] = [];
      const savedUpdates: string[] = [];
      let operation = "cluster link update";
      const updateBoth = hasAuthGroups && values.applyDescriptionToBoth;
      const sourceGroupsChanged =
        bidirectional &&
        editedAuthGroups.source &&
        values.sourceAuthGroups !== undefined &&
        !areSameArrays(
          values.sourceAuthGroups,
          sourceIdentityQuery.data?.groups ?? [],
        );
      const targetGroupsChanged =
        hasAuthGroups &&
        editedAuthGroups.target &&
        values.targetAuthGroups !== undefined &&
        !areSameArrays(
          values.targetAuthGroups,
          targetIdentityQuery.data?.groups ?? [],
        );
      const recordSuccess = (label: string, message: ReactNode): void => {
        savedUpdates.push(label);
        toastNotify.success(message);
        if (savedUpdates.length === 2 && !toastNotify.isListView) {
          toastNotify.toggleListView();
        }
      };

      try {
        if (!cluster || !link) {
          throw new Error("Load the cluster link before saving changes.");
        }
        if (
          sourceGroupsChanged &&
          (!sourceIdentityQuery.isSuccess || !sourceAuthGroupsQuery.isSuccess)
        ) {
          throw new Error(
            "Load the source identity and auth groups before changing source auth groups.",
          );
        }
        if (
          targetGroupsChanged &&
          (!canLoadTarget ||
            !targetIdentityQuery.isSuccess ||
            !targetAuthGroupsQuery.isSuccess)
        ) {
          throw new Error(
            "Load the target identity and auth groups before changing target auth groups.",
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
        attemptedClusters.push(clusterName);
        await updateClusterLink(
          clusterName,
          linkName,
          JSON.stringify({
            description: values.description,
            config: link.config,
          }),
        );
        recordSuccess(
          "Source description",
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
          operation = "target description update";
          attemptedClusters.push(targetCluster.name);
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
          sourceIdentityQuery.data &&
          values.sourceAuthGroups !== undefined
        ) {
          operation = "source auth groups update";
          const identity = {
            ...sourceIdentityQuery.data,
            groups: values.sourceAuthGroups,
          };
          await updateIdentity(clusterName, identity);
          queryClient.setQueryData(sourceIdentityKey, identity);
          recordSuccess(
            "Source auth groups",
            <>
              Source auth groups for cluster link <b>{linkName}</b> on{" "}
              <b>{clusterName}</b> saved.
            </>,
          );
        }
        if (
          targetGroupsChanged &&
          targetIdentityQuery.data &&
          values.targetAuthGroups !== undefined
        ) {
          operation = "target auth groups update";
          const identity = {
            ...targetIdentityQuery.data,
            groups: values.targetAuthGroups,
          };
          await updateIdentity(targetClusterName, identity);
          queryClient.setQueryData(targetIdentityKey, identity);
          recordSuccess(
            "Target auth groups",
            <>
              Target auth groups for cluster link <b>{linkName}</b> on{" "}
              <b>{targetClusterName}</b> saved.
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
        void invalidateClusterLinkLists(queryClient, attemptedClusters);
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
    if (!cluster || !link) {
      if (initializedLink.current !== null) {
        setError(null);
      }
      initializedLink.current = null;
      return;
    }
    const key = `${cluster.name}/${link.name}`;
    if (initializedLink.current === key) {
      return;
    }
    initializedLink.current = key;
    formik.resetForm({
      values: {
        description: link.description ?? "",
        applyDescriptionToBoth: false,
        sourceAuthGroups: sourceIdentityQuery.data
          ? (sourceIdentityQuery.data.groups ?? [])
          : undefined,
        targetAuthGroups: targetIdentityQuery.data
          ? (targetIdentityQuery.data.groups ?? [])
          : undefined,
      },
    });
    setEditedAuthGroups({ source: false, target: false });
    setError(null);
  }, [
    cluster,
    link,
    formik.resetForm,
    sourceIdentityQuery.data,
    targetIdentityQuery.data,
  ]);

  useEffect(() => {
    if (
      formik.values.sourceAuthGroups === undefined &&
      sourceIdentityQuery.isSuccess
    ) {
      void formik.setFieldValue(
        "sourceAuthGroups",
        sourceIdentityQuery.data.groups ?? [],
        false,
      );
    }
    if (
      formik.values.targetAuthGroups === undefined &&
      targetIdentityQuery.isSuccess
    ) {
      void formik.setFieldValue(
        "targetAuthGroups",
        targetIdentityQuery.data.groups ?? [],
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

  if (!cluster) {
    return null;
  }

  return (
    <SidePanel>
      <SidePanel.Header>
        <SidePanel.HeaderTitle>Edit cluster link</SidePanel.HeaderTitle>
      </SidePanel.Header>
      {link && <NotificationRow className="u-no-padding" />}
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
      <SidePanel.Content
        className={link ? "u-no-padding edit-cluster-link-panel" : undefined}
      >
        {link ? (
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
            <ClusterLinkForm
              mode="edit"
              sourceClusterName={cluster.name}
              targetClusterName={targetClusterName}
              type={link.type}
              name={link.name}
              descriptionField={formik.getFieldProps("description")}
              disabled={formik.isSubmitting}
              applyDescriptionToBoth={formik.values.applyDescriptionToBoth}
              onApplyDescriptionToBothChange={() => {
                void formik.setFieldValue(
                  "applyDescriptionToBoth",
                  !formik.values.applyDescriptionToBoth,
                );
              }}
              onSubmit={(e) => {
                e.preventDefault();
                void formik.submitForm();
              }}
            >
              {hasAuthGroups && (
                <ClusterLinkAuthGroupsSection
                  target={{
                    clusterName: targetClusterName,
                    query: targetAuthGroupsQuery,
                    identityQuery: targetIdentityQuery,
                    unavailableReason: canLoadTarget
                      ? undefined
                      : "The target cluster must be managed and have a registered tunnel to edit target auth groups.",
                    selectedGroups:
                      formik.values.targetAuthGroups ??
                      targetIdentityQuery.data?.groups ??
                      [],
                    onChange: (groups) => {
                      setEditedAuthGroups((current) => ({
                        ...current,
                        target: true,
                      }));
                      void formik.setFieldValue("targetAuthGroups", groups);
                    },
                  }}
                  disabled={formik.isSubmitting}
                  {...(bidirectional
                    ? {
                        isBidirectional: true,
                        source: {
                          clusterName,
                          query: sourceAuthGroupsQuery,
                          identityQuery: sourceIdentityQuery,
                          selectedGroups:
                            formik.values.sourceAuthGroups ??
                            sourceIdentityQuery.data?.groups ??
                            [],
                          onChange: (groups: string[]) => {
                            setEditedAuthGroups((current) => ({
                              ...current,
                              source: true,
                            }));
                            void formik.setFieldValue(
                              "sourceAuthGroups",
                              groups,
                            );
                          },
                        },
                      }
                    : { isBidirectional: false })}
                />
              )}
            </ClusterLinkForm>
          </ScrollableContainer>
        ) : isLinksLoading ? (
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
        <Button
          appearance={link ? "base" : undefined}
          className="u-no-margin--bottom"
          onClick={closePanel}
        >
          {link ? "Cancel" : "Close"}
        </Button>
        {link && (
          <ActionButton
            appearance="positive"
            className="u-no-margin--bottom"
            loading={formik.isSubmitting}
            disabled={!formik.isValid || formik.isSubmitting}
            onClick={() => void formik.submitForm()}
          >
            Save
          </ActionButton>
        )}
      </SidePanel.Footer>
    </SidePanel>
  );
};

export default EditClusterLinkPanel;
