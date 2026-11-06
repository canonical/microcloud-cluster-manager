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
import { type FC, useState } from "react";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import usePanelParams from "context/usePanelParams";
import { useCluster } from "context/useClusters";
import { createClusterLink } from "api/clusters";
import type { ClusterLink, ClusterLinkPost } from "types/cluster";
import * as Yup from "yup";
import ClusterSelector from "pages/clusters/ClusterSelector";
import { LINK_TYPES } from "pages/clusters/clusterLinkTypes";
import editOffIcon from "icons/edit_off.svg";

const CreateClusterLinkPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const clusterName = panelParams.cluster ?? "";

  const { data: cluster } = useCluster(clusterName);
  const [nameManuallyEdited, setNameManuallyEdited] = useState(false);

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface CreateClusterLinkFormValues {
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
      auth_groups: [],
      description: values.description,
      type: values.type,
    };
    if (values.type === "public" && values.useDifferentTargetAddress) {
      payload.remote_address = values.targetAddress;
    }

    return createClusterLink(clusterName, JSON.stringify(payload))
      .then(() => {
        toastNotification.success(
          <>
            Created cluster link <strong>{payload.name}</strong>.
          </>,
        );
        closePanel();
      })
      .catch((error: Error) => {
        notify.failure("Failure creating cluster link.", error);
      })
      .finally(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters],
        });
      });
  };

  const ConfigurationSchema = Yup.object().shape({
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
    name: Yup.string().trim().required("Link name is required"),
  });

  const formik = useFormik<CreateClusterLinkFormValues>({
    initialValues: {
      targetCluster: "",
      type: "bidirectional",
      useDifferentTargetAddress: false,
      targetAddress: "",
      name: "",
      description: "",
    },
    validationSchema: ConfigurationSchema,
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
    !formik.values.targetCluster ||
    !formik.values.name.trim() ||
    isAlternateTargetAddressInvalid;

  if (!cluster) {
    return null;
  }

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>{`Create cluster link from ${cluster.name}`}</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <NotificationRow className="u-no-padding" />
        <SidePanel.Content className="u-no-padding">
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
                <h2 className="p-muted-heading">Link type</h2>
                <div className="cluster-link-types">
                  {LINK_TYPES.map((type) => (
                    <RadioInput
                      key={type.value}
                      name="type"
                      label={type.label}
                      help={type.help}
                      checked={formik.values.type === type.value}
                      onChange={() => {
                        void formik.setFieldValue("type", type.value);
                        resetTargetAddress();
                      }}
                    />
                  ))}
                </div>
                <hr />
                <h2 className="p-muted-heading">Details</h2>
                <div className="cluster-link-endpoints">
                  <Input
                    id="source-cluster"
                    type="text"
                    label={
                      <>
                        Source{" "}
                        <img src={editOffIcon} alt="" width={16} height={16} />
                      </>
                    }
                    value={cluster.name}
                    readOnly
                  />
                  <span
                    className="cluster-link-endpoints__arrow"
                    aria-hidden="true"
                  >
                    {formik.values.type === "bidirectional" ? "⇄" : "→"}
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
                          ...(nameManuallyEdited
                            ? {}
                            : {
                                name: value
                                  ? `${cluster.name}-${value}-clusterlink`
                                  : "",
                              }),
                        });
                      }}
                      ignoreOptions={[cluster.name]}
                      selectProps={{
                        id: "target-cluster",
                        name: "target",
                      }}
                    />
                  </div>
                </div>
                {formik.values.type === "public" && (
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
                    {formik.values.useDifferentTargetAddress && (
                      <Input
                        name="targetAddress"
                        type="url"
                        label="Alternate LXD URL"
                        placeholder="https://cluster.example:8443"
                        value={formik.values.targetAddress}
                        onBlur={formik.handleBlur}
                        onChange={(e) => {
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
                    )}
                  </>
                )}
                <Input
                  id="name"
                  type="text"
                  label="Link name"
                  required
                  placeholder={`${cluster.name}-target-clusterlink`}
                  {...formik.getFieldProps("name")}
                  onChange={(e) => {
                    setNameManuallyEdited(true);
                    formik.handleChange(e);
                  }}
                  error={formik.touched.name ? formik.errors.name : null}
                />
                <Input
                  id="description"
                  type="text"
                  label="Description"
                  {...formik.getFieldProps("description")}
                />
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
