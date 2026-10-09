import {
  ActionButton,
  Button,
  Form,
  Input,
  ScrollableContainer,
  Select,
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
import { createClusterLink, fetchCluster } from "api/clusters";
import * as Yup from "yup";
import ClusterSelector from "pages/clusters/ClusterSelector";

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

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface CreateClusterLinkFormValues {
    targetCluster: string;
    direction: "bidirectional" | "unidirectional";
  }

  const handleSubmit = (values: CreateClusterLinkFormValues) => {
    const payload = {
      target_cluster: values.targetCluster,
      name: values.targetCluster,
      auth_groups: [],
      description: "",
      type: values.direction,
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
  });

  const formik = useFormik<CreateClusterLinkFormValues>({
    initialValues: {
      targetCluster: "",
      direction: "bidirectional",
    },
    validationSchema: ConfigurationSchema,
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

  if (!cluster) {
    return null;
  }

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>
            Create cluster link from {cluster.name}
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
              <ClusterSelector
                value={formik.values.targetCluster}
                setValue={(value) => {
                  formik.setFieldValue("targetCluster", value);
                }}
                ignoreOptions={[cluster.name]}
              />
              <Select
                id="direction"
                name="direction"
                label="Direction"
                value={formik.values.direction}
                onChange={(e) => {
                  formik.setFieldValue("direction", e.target.value);
                }}
                options={[
                  { value: "bidirectional", label: "Bidirectional" },
                  { value: "unidirectional", label: "Unidirectional" },
                ]}
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
            loading={formik.isSubmitting}
            disabled={!formik.isValid || !formik.values.targetCluster}
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
