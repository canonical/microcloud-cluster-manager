import {
  ActionButton,
  Button,
  CheckboxInput,
  Input,
  ScrollableContainer,
  SidePanel,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { useFormik } from "formik";
import NotificationRow from "components/NotificationRow";
import usePanelParams from "context/usePanelParams";
import { useCluster } from "context/useClusters";
import { createClusterLink } from "api/clusters";
import type { ClusterLink, ClusterLinkPost } from "types/cluster";
import * as Yup from "yup";
import ClusterSelector from "components/ClusterSelector";
import ClusterLinkForm from "components/forms/ClusterLinkForm";
import { useAuthGroups } from "context/useAuthGroups";
import ClusterLinkAuthGroupsSection from "pages/clusters/ClusterLinkAuthGroupsSection";
import ClusterLinkRichChip from "pages/clusters/ClusterLinkRichChip";
import {
  invalidateClusterLinksAndIdentities,
  isBidirectional,
  isPublic,
  isUnidirectional,
} from "util/clusterLinks";

interface CreateClusterLinkFormValues {
  sourceCluster: string;
  targetCluster: string;
  type: ClusterLink["type"];
  useDifferentTargetAddress: boolean;
  targetAddress: string;
  name: string;
  description: string;
  sourceAuthGroups: string[];
  targetAuthGroups: string[];
}

const CreateClusterLinkPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const clusterName = panelParams.cluster ?? "";

  const { data: cluster } = useCluster(clusterName);

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  const handleSubmit = async (
    values: CreateClusterLinkFormValues,
  ): Promise<void> => {
    if (!areAuthGroupsReady) {
      notify.failure(
        "Failure creating cluster link.",
        new Error("Failed to load auth groups."),
      );
      return;
    }

    const payload: ClusterLinkPost = {
      target_cluster: values.targetCluster,
      name: values.name.trim(),
      description: values.description,
      type: values.type,
      source_auth_groups: isBidirectional(values)
        ? values.sourceAuthGroups
        : [],
    };
    if (!isPublic(values)) {
      payload.target_auth_groups = values.targetAuthGroups;
    }
    if (isPublic(values) && values.useDifferentTargetAddress) {
      payload.remote_address = values.targetAddress;
    }

    return createClusterLink(values.sourceCluster, JSON.stringify(payload))
      .then(() => {
        toastNotification.success(
          <>
            Created cluster link{" "}
            <ClusterLinkRichChip
              clusterLinkName={payload.name}
              clusterLinkType={payload.type}
              cluster={values.sourceCluster}
            />
            .
          </>,
        );
        closePanel();
      })
      .catch((error: Error) => {
        notify.failure("Failure creating cluster link.", error);
      })
      .finally(() => {
        void invalidateClusterLinksAndIdentities(
          queryClient,
          values.sourceCluster,
          payload,
          values.targetCluster,
        );
      });
  };

  const ConfigurationSchema = Yup.object().shape({
    sourceCluster: Yup.string().required("Source cluster is required"),
    targetCluster: Yup.string().required("Target cluster is required"),
    useDifferentTargetAddress: Yup.boolean(),
    targetAddress: Yup.string().when("useDifferentTargetAddress", {
      is: true,
      then: (schema) =>
        schema
          .required("Target address is required")
          .url("Enter a valid HTTPS URL")
          .test(
            "https-url",
            "Target address must use HTTPS",
            (value) => !value || value.startsWith("https://"),
          ),
    }),
  });

  const formik = useFormik<CreateClusterLinkFormValues>({
    initialValues: {
      sourceCluster: clusterName,
      targetCluster: "",
      type: "bidirectional",
      useDifferentTargetAddress: false,
      targetAddress: "",
      name: "",
      description: "",
      sourceAuthGroups: [],
      targetAuthGroups: [],
    },
    validationSchema: ConfigurationSchema,
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

  const targetAuthGroupsQuery = useAuthGroups(
    formik.values.targetCluster,
    Boolean(formik.values.targetCluster) && !isPublic(formik.values),
  );
  const sourceAuthGroupsQuery = useAuthGroups(
    formik.values.sourceCluster,
    Boolean(formik.values.sourceCluster) && isBidirectional(formik.values),
  );

  const areAuthGroupsReady =
    isPublic(formik.values) ||
    (targetAuthGroupsQuery.isSuccess &&
      (!isBidirectional(formik.values) || sourceAuthGroupsQuery.isSuccess));

  const resetTargetAddress = (resetAddressToggle = true) => {
    if (resetAddressToggle) {
      void formik.setFieldValue("useDifferentTargetAddress", false);
    }
    void formik.setFieldValue("targetAddress", "");
  };

  const isAlternateTargetAddressInvalid =
    formik.values.useDifferentTargetAddress &&
    (!formik.values.targetAddress || !!formik.errors.targetAddress);
  const isSubmitDisabled =
    formik.isSubmitting ||
    !formik.values.sourceCluster ||
    !formik.values.targetCluster ||
    !formik.values.name ||
    !areAuthGroupsReady ||
    isAlternateTargetAddressInvalid;

  if (!cluster) {
    return null;
  }

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>Create cluster link</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <NotificationRow className="u-no-padding" />
        <SidePanel.Content className="u-no-padding create-cluster-link-panel">
          <ScrollableContainer
            dependencies={[notify.notification]}
            belowIds={["panel-footer"]}
          >
            <ClusterLinkForm
              mode="create"
              sourceClusterName={cluster.name}
              type={formik.values.type}
              name={formik.values.name}
              descriptionField={formik.getFieldProps("description")}
              disabled={formik.isSubmitting}
              onTypeChange={(type) => {
                void formik.setValues({
                  ...formik.values,
                  type,
                });
                resetTargetAddress();
              }}
              onSubmit={(e) => {
                e.preventDefault();
                void formik.submitForm();
              }}
              targetSelector={
                <div>
                  <label htmlFor="target-cluster">Target</label>
                  <ClusterSelector
                    value={formik.values.targetCluster}
                    setValue={(value) => {
                      void formik.setValues({
                        ...formik.values,
                        targetCluster: value,
                        useDifferentTargetAddress: false,
                        targetAddress: "",
                        targetAuthGroups: [],
                        name: value ? `${value}-link` : "",
                      });
                    }}
                    ignoreOptions={[formik.values.sourceCluster]}
                    selectProps={{
                      id: "target-cluster",
                      name: "target",
                      disabled: formik.isSubmitting,
                    }}
                  />
                </div>
              }
              addressControls={
                isPublic(formik.values) && (
                  <>
                    <CheckboxInput
                      id="use-different-target-address"
                      className="cluster-link-address-checkbox"
                      label="Use a different address than the registered LXD URL"
                      checked={formik.values.useDifferentTargetAddress}
                      disabled={formik.isSubmitting}
                      onChange={() => {
                        void formik.setFieldValue(
                          "useDifferentTargetAddress",
                          !formik.values.useDifferentTargetAddress,
                        );
                        resetTargetAddress(false);
                      }}
                    />
                    <Input
                      name="targetAddress"
                      type="url"
                      aria-label="Alternate LXD URL"
                      disabled={
                        formik.isSubmitting || !formik.values.targetCluster
                      }
                      placeholder="https://cluster.example:8443"
                      value={formik.values.targetAddress}
                      onBlur={formik.handleBlur}
                      onChange={(e) => {
                        void formik.setFieldValue(
                          "useDifferentTargetAddress",
                          true,
                        );
                        void formik.setFieldValue(
                          "targetAddress",
                          e.target.value,
                        );
                      }}
                      error={
                        formik.touched.targetAddress &&
                        formik.errors.targetAddress
                      }
                    />
                  </>
                )
              }
            >
              {(isBidirectional(formik.values) ||
                isUnidirectional(formik.values)) && (
                <ClusterLinkAuthGroupsSection
                  target={{
                    clusterName: formik.values.targetCluster,
                    query: targetAuthGroupsQuery,
                    selectedGroups: formik.values.targetAuthGroups,
                    onChange: (groups) => {
                      void formik.setFieldValue("targetAuthGroups", groups);
                    },
                  }}
                  disabled={formik.isSubmitting}
                  {...(isBidirectional(formik.values)
                    ? {
                        isBidirectional: true,
                        source: {
                          clusterName: formik.values.sourceCluster,
                          query: sourceAuthGroupsQuery,
                          selectedGroups: formik.values.sourceAuthGroups,
                          onChange: (groups: string[]) => {
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
        </SidePanel.Content>
        <SidePanel.Footer className="u-align--right">
          <Button
            appearance="base"
            className="u-no-margin--bottom"
            onClick={closePanel}
            disabled={formik.isSubmitting}
          >
            Cancel
          </Button>
          <ActionButton
            appearance="positive"
            className="u-no-margin--bottom"
            loading={formik.isSubmitting}
            disabled={isSubmitDisabled}
            onClick={() => void formik.submitForm()}
          >
            Create
          </ActionButton>
        </SidePanel.Footer>
      </SidePanel>
    </>
  );
};

export default CreateClusterLinkPanel;
