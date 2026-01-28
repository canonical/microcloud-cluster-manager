import {
  ActionButton,
  Button,
  Form,
  Input,
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
import { fetchCluster, updateCluster } from "api/clusters";
import * as Yup from "yup";

export const FIELD_DISK_THRESHOLD = "diskThreshold";
export const FIELD_MEMORY_THRESHOLD = "memoryThreshold";
export const FIELD_DESCRIPTION = "description";

const ConfigureClusterPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const clusterName = panelParams.cluster ?? "";
  const focusField = panelParams.focusField ?? FIELD_DISK_THRESHOLD;

  const { data: cluster } = useQuery({
    queryKey: [queryKeys.clusters, clusterName],
    queryFn: async () => fetchCluster(clusterName),
  });

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface ConfigureClusterFormValues {
    description: string;
    diskThreshold: number;
    memoryThreshold: number;
  }

  const handleSubmit = (values: ConfigureClusterFormValues) => {
    const payload = {
      description: values.description,
      disk_threshold: values.diskThreshold,
      memory_threshold: values.memoryThreshold,
    };

    updateCluster(clusterName, JSON.stringify(payload))
      .then(() => {
        toastNotification.success(
          <>
            Updated cluster <strong>{clusterName}</strong>.
          </>,
        );
        closePanel();
      })
      .catch((e: Error) => {
        notify.failure(`Failure updating cluster ${clusterName}.`, e);
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
    description: Yup.string(),
    diskThreshold: Yup.number().min(0).max(100),
    memoryThreshold: Yup.number().min(0).max(100),
  });

  const formik = useFormik<ConfigureClusterFormValues>({
    initialValues: {
      description: cluster?.description ?? "",
      diskThreshold: cluster?.disk_threshold ?? 80,
      memoryThreshold: cluster?.memory_threshold ?? 80,
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
            Configure cluster {cluster.name}
          </SidePanel.HeaderTitle>
        </SidePanel.Header>
        <NotificationRow className="u-no-padding" />
        <SidePanel.Content className="u-no-padding">
          <ScrollableContainer
            dependencies={[notify.notification]}
            belowIds={["panel-footer"]}
          >
            <Form onSubmit={() => void formik.submitForm()} className="form">
              <Input
                name={FIELD_DISK_THRESHOLD}
                type="number"
                label="Disk threshold percentage"
                placeholder="Enter value"
                min={1}
                max={100}
                autoFocus={focusField === FIELD_DISK_THRESHOLD}
                onBlur={formik.handleBlur}
                onChange={formik.handleChange}
                value={formik.values.diskThreshold}
                error={
                  formik.touched.diskThreshold && formik.errors.diskThreshold
                }
              />
              <Input
                name={FIELD_MEMORY_THRESHOLD}
                type="number"
                label="Memory threshold percentage"
                placeholder="Enter value"
                min={1}
                max={100}
                autoFocus={focusField === FIELD_MEMORY_THRESHOLD}
                onBlur={formik.handleBlur}
                onChange={formik.handleChange}
                value={formik.values.memoryThreshold}
                error={
                  formik.touched.memoryThreshold &&
                  formik.errors.memoryThreshold
                }
              />
              <Input
                name={FIELD_DESCRIPTION}
                type="text"
                label="Description"
                placeholder="Enter description"
                autoFocus={focusField === FIELD_DESCRIPTION}
                onBlur={formik.handleBlur}
                onChange={formik.handleChange}
                value={formik.values.description}
                error={formik.touched.description && formik.errors.description}
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
            disabled={!formik.isValid}
            onClick={() => void formik.submitForm()}
          >
            Save changes
          </ActionButton>
        </SidePanel.Footer>
      </SidePanel>
    </>
  );
};

export default ConfigureClusterPanel;
