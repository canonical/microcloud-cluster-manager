import {
  ActionButton,
  Button,
  CheckboxInput,
  Form,
  Icon,
  Input,
  RadioInput,
  ScrollableContainer,
  SidePanel,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import usePanelParams from "context/usePanelParams";
import {
  createClusterLink,
  fetchCluster,
  fetchClusterLinks,
  fetchClusters,
} from "api/clusters";
import { panels } from "context/usePanelParams";
import { getLinkedCluster } from "util/clusterLinks";
import * as Yup from "yup";
import ClusterSelector from "pages/clusters/ClusterSelector";

type LinkType = "bidirectional" | "unidirectional" | "public";

const LINK_TYPES: { value: LinkType; label: string; help: string }[] = [
  {
    value: "bidirectional",
    label: "Bidirectional",
    help: "Connect two LXD clusters with access in both directions. Required for replication.",
  },
  {
    value: "unidirectional",
    label: "Unidirectional",
    help: "Access another LXD cluster without granting access in return. Often used to fetch images.",
  },
  {
    value: "public",
    label: "Public",
    help: "Access another LXD cluster without authentication or return access. Used for public image fetching.",
  },
];

const CreateClusterLinkPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const clusterName = panelParams.cluster ?? "";

  const { data: cluster } = useQuery({
    queryKey: [queryKeys.clusters, clusterName],
    queryFn: async () => fetchCluster(clusterName),
  });

  const isEdit = panelParams.panel === panels.editClusterLink;

  const { data: clusterLinks = [] } = useQuery({
    queryKey: [queryKeys.clusters, clusterName, queryKeys.links],
    queryFn: async () => fetchClusterLinks(clusterName),
    enabled: isEdit,
  });

  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
    enabled: isEdit,
  });

  const existingLink = isEdit
    ? clusterLinks.find((link) => link.name === panelParams.link)
    : undefined;
  const existingTarget = existingLink
    ? getLinkedCluster(existingLink, clusters)
    : undefined;

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface CreateClusterLinkFormValues {
    targetCluster: string;
    linkType: LinkType;
    name: string;
    description: string;
    applyNameToBoth: boolean;
    applyDescriptionToBoth: boolean;
  }

  const handleSubmit = (values: CreateClusterLinkFormValues) => {
    const payload = {
      target_cluster: values.targetCluster,
      name: values.name.trim(),
      auth_groups: [],
      description: values.description,
      type: values.linkType,
    };

    createClusterLink(clusterName, JSON.stringify(payload))
      .then(() => {
        toastNotification.success(<>Created cluster link.</>);
        closePanel();
      })
      .catch((e: Error) => {
        notify.failure(`Failure creating cluster link.`, e);
      })
      .finally(() => {
        queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters, cluster?.name],
        });
        queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters],
        });
        formik.setSubmitting(false);
      });
  };

  const ConfigurationSchema = Yup.object().shape({
    targetCluster: Yup.string().required("Target cluster is required"),
    name: Yup.string().trim().required("Link name is required"),
  });

  const formik = useFormik<CreateClusterLinkFormValues>({
    initialValues: {
      targetCluster: existingTarget?.name ?? "",
      linkType:
        (existingLink?.type.toLowerCase() as LinkType) ?? "bidirectional",
      name: existingLink?.name ?? "",
      description: existingLink?.description ?? "",
      applyNameToBoth: false,
      applyDescriptionToBoth: false,
    },
    validationSchema: ConfigurationSchema,
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

  if (!cluster || (isEdit && !existingLink)) {
    return null;
  }

  const selectedType =
    LINK_TYPES.find((type) => type.value === formik.values.linkType) ??
    LINK_TYPES[0];

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>
            {isEdit ? "Edit cluster link" : "Create cluster link"}
          </SidePanel.HeaderTitle>
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
              <h2 className="p-muted-heading">Link type</h2>
              <div className="cluster-link-types">
                {isEdit ? (
                  <>
                    <p className="u-no-margin--bottom">{selectedType.label}</p>
                    <p className="p-form-help-text cluster-link-types__help">
                      {selectedType.help}
                    </p>
                  </>
                ) : (
                  LINK_TYPES.map((type) => (
                    <RadioInput
                      key={type.value}
                      name="linkType"
                      label={type.label}
                      help={type.help}
                      checked={formik.values.linkType === type.value}
                      onChange={() => {
                        void formik.setFieldValue("linkType", type.value);
                      }}
                    />
                  ))
                )}
              </div>
              <hr />
              <h2 className="p-muted-heading">Details</h2>
              <div className="cluster-link-endpoints">
                <Input
                  id="source-cluster"
                  type="text"
                  label={
                    <>
                      Source <Icon name="union" />
                    </>
                  }
                  value={cluster.name}
                  disabled
                />
                <span
                  className="cluster-link-endpoints__arrow"
                  aria-hidden="true"
                >
                  {formik.values.linkType === "bidirectional" ? "⇄" : "→"}
                </span>
                {isEdit ? (
                  <Input
                    id="target-cluster"
                    type="text"
                    label={
                      <>
                        Target <Icon name="union" />
                      </>
                    }
                    value={formik.values.targetCluster}
                    disabled
                  />
                ) : (
                  <div>
                    <label htmlFor="target-cluster">Target</label>
                    <ClusterSelector
                      value={formik.values.targetCluster}
                      setValue={(value) => {
                        void formik.setFieldValue("targetCluster", value);
                      }}
                      ignoreOptions={[cluster.name]}
                      selectProps={{ id: "target-cluster", name: "target" }}
                    />
                  </div>
                )}
              </div>
              <Input
                id="name"
                type="text"
                label="Link name"
                required
                placeholder={`${cluster.name}-target-clusterlink`}
                {...formik.getFieldProps("name")}
                error={formik.touched.name ? formik.errors.name : null}
              />
              {isEdit && (
                <CheckboxInput
                  label="Apply change in both source and target clusters"
                  checked={formik.values.applyNameToBoth}
                  onChange={() => {
                    void formik.setFieldValue(
                      "applyNameToBoth",
                      !formik.values.applyNameToBoth,
                    );
                  }}
                />
              )}
              <Input
                id="description"
                type="text"
                label="Description"
                {...formik.getFieldProps("description")}
              />
              {isEdit && (
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
            disabled={
              isEdit || // saving changes needs a backend update endpoint
              !formik.isValid ||
              !formik.values.targetCluster ||
              !formik.values.name.trim()
            }
            onClick={() => void formik.submitForm()}
          >
            Save changes
          </ActionButton>
        </SidePanel.Footer>
      </SidePanel>
    </>
  );
};

export default CreateClusterLinkPanel;
