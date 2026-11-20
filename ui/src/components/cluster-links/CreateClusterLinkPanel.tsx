import {
  ActionButton,
  Button,
  CheckboxInput,
  Form,
  Input,
  RadioInput,
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
import { getLinkIconType, LINK_TYPES } from "util/clusterLinkTypes";
import ResourceIcon from "components/ResourceIcon";
import {
  invalidateClusterLinksAndIdentities,
  isPublic,
} from "util/clusterLinks";
import PanelSection from "components/PanelSection";
import ClusterLinkName from "pages/clusters/ClusterLinkName";
import ClusterLinkRichChip from "pages/clusters/ClusterLinkRichChip";

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

  interface CreateClusterLinkFormValues {
    sourceCluster: string;
    targetCluster: string;
    type: ClusterLink["type"];
    useDifferentTargetAddress: boolean;
    targetAddress: string;
    name: string;
    description: string;
  }

  const handleSubmit = async (
    values: CreateClusterLinkFormValues,
  ): Promise<void> => {
    const payload: ClusterLinkPost = {
      target_cluster: values.targetCluster,
      name: values.name.trim(),
      source_auth_groups: [],
      description: values.description,
      type: values.type,
    };
    if (isPublic(payload) && values.useDifferentTargetAddress) {
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
    },
    validationSchema: ConfigurationSchema,
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

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
            <Form
              onSubmit={(e) => {
                e.preventDefault();
                void formik.submitForm();
              }}
              className="form"
            >
              {/* hidden submit to enable enter key in inputs */}
              <Input type="submit" hidden value="Hidden input" />
              <>
                <PanelSection header="Link type" isFirst>
                  <div className="cluster-link-types cluster-link-types--radio-buttons">
                    {LINK_TYPES.map((type) => (
                      <RadioInput
                        key={type.value}
                        name="type"
                        label={type.label}
                        help={type.description}
                        checked={formik.values.type === type.value}
                        onChange={() => {
                          void formik.setFieldValue("type", type.value);
                          resetTargetAddress();
                        }}
                      />
                    ))}
                  </div>
                </PanelSection>
                <PanelSection header="Details">
                  <div className="cluster-link-endpoints">
                    <Input
                      type="text"
                      label="Source"
                      value={cluster?.name ?? clusterName}
                      readOnly
                    />
                    <span
                      className="cluster-link-endpoints__arrow"
                      aria-hidden="true"
                    >
                      <ResourceIcon
                        type={getLinkIconType(formik.values.type)}
                      />
                    </span>
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
                            name: value ? `${value}-link` : "",
                          });
                        }}
                        ignoreOptions={[formik.values.sourceCluster]}
                        selectProps={{
                          id: "target-cluster",
                          name: "target",
                        }}
                      />
                    </div>
                  </div>
                  {isPublic(formik.values) && (
                    <>
                      <CheckboxInput
                        id="use-different-target-address"
                        className="cluster-link-address-checkbox"
                        label="Use a different address than the registered LXD URL"
                        checked={formik.values.useDifferentTargetAddress}
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
                        disabled={!formik.values.targetCluster}
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
                  )}
                  <ClusterLinkName name={formik.values.name} />
                  <Input
                    id="description"
                    type="text"
                    label="Description"
                    {...formik.getFieldProps("description")}
                  />
                </PanelSection>
              </>
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
