import {
  ActionButton,
  Button,
  Form,
  Input,
  useNotify,
} from "@canonical/react-components";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import ScrollableContainer from "components/ScrollableContainer";
import SidePanel from "components/SidePanel";
import usePanelParams from "context/usePanelParams";
import { fetchCluster, updateCluster } from "api/clusters";

const ConfigureClusterPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();

  const clusterName = panelParams.cluster ?? "";

  const { data: cluster } = useQuery({
    queryKey: [queryKeys.clusters, clusterName],
    queryFn: () => fetchCluster(clusterName),
  });

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface ConfigureClusterFormValues {
    diskThreshold: number;
    memoryThreshold: number;
  }

  const handleSubmit = (values: ConfigureClusterFormValues) => {
    const payload = {
      disk_threshold: values.diskThreshold,
      memory_threshold: values.memoryThreshold,
    };

    updateCluster(clusterName, JSON.stringify(payload))
      .then(() => {
        notify.success(`Successfully updated cluster ${clusterName}.`);
        closePanel();
      })
      .catch((e: Error) => {
        notify.failure(`Failure updating cluster ${clusterName}.`, e);
      })
      .finally(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters, cluster?.name],
        });
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters],
        });
        formik.setSubmitting(false);
      });
  };

  const formik = useFormik<ConfigureClusterFormValues>({
    initialValues: {
      diskThreshold: cluster?.disk_threshold ?? 80,
      memoryThreshold: cluster?.memory_threshold ?? 80,
    },
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

  if (!cluster) {
    return null;
  }

  return (
    <>
      <SidePanel isOverlay loading={false} hasError={false}>
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
                name="diskThreshold"
                type="number"
                label="Disk threshold"
                placeholder="Enter value"
                min={1}
                max={100}
                autoFocus
                onBlur={formik.handleBlur}
                onChange={formik.handleChange}
                value={formik.values.diskThreshold}
              />
              <Input
                name="memoryThreshold"
                type="number"
                label="Memory threshold"
                placeholder="Enter value"
                min={1}
                max={100}
                onBlur={formik.handleBlur}
                onChange={formik.handleChange}
                value={formik.values.memoryThreshold}
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
