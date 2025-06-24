import {
  ActionButton,
  Button,
  Form,
  Input,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import ScrollableContainer from "components/ScrollableContainer";
import SidePanel from "components/SidePanel";
import usePanelParams from "context/usePanelParams";
import { fetchClusters, updateClusterBulk } from "api/clusters";
import { pluralize } from "util/helpers";
import BulkConfigurePanelInput from "components/forms/BulkConfigurePanelInput";

const BulkConfigureClusterPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const { data: allClusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });
  const clusterNames = (panelParams.clusters ?? "").split(",");
  const clusters = allClusters.filter((cluster) => {
    return clusterNames.includes(cluster.name);
  });

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface ConfigureClusterFormValues {
    diskThreshold?: number;
    memoryThreshold?: number;
  }

  const handleSubmit = (values: ConfigureClusterFormValues) => {
    const payload = {
      disk_threshold: values.diskThreshold,
      memory_threshold: values.memoryThreshold,
    };

    updateClusterBulk(clusterNames, JSON.stringify(payload))
      .then(() => {
        toastNotification.success(
          <>
            Updated{" "}
            <strong>
              {clusterNames.length} {pluralize("cluster", clusterNames.length)}
            </strong>
          </>,
        );
        closePanel();
      })
      .catch((e: Error) => {
        notify.failure("Failed to update clusters.", e);
      })
      .finally(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters],
        });
        formik.setSubmitting(false);
      });
  };

  const areDiskThresholdsEqual =
    clusters.length > 0 &&
    clusters.every(
      (cluster) => cluster.disk_threshold === clusters?.[0]?.disk_threshold,
    );

  const areMemoryThresholdsEqual =
    clusters.length > 0 &&
    clusters.every(
      (cluster) => cluster.memory_threshold === clusters?.[0]?.memory_threshold,
    );

  const formik = useFormik<ConfigureClusterFormValues>({
    initialValues: {
      diskThreshold: undefined,
      memoryThreshold: undefined,
    },
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

  return (
    <>
      <SidePanel isOverlay loading={false} hasError={false}>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>
            Configure {clusterNames.length}{" "}
            {pluralize("cluster", clusterNames.length)}
          </SidePanel.HeaderTitle>
        </SidePanel.Header>
        <NotificationRow className="u-no-padding" />
        <SidePanel.Content className="u-no-padding">
          <ScrollableContainer
            dependencies={[notify.notification]}
            belowIds={["panel-footer"]}
          >
            <Form onSubmit={() => void formik.submitForm()} className="form">
              <BulkConfigurePanelInput
                areAllValuesEqual={areDiskThresholdsEqual}
                setValue={(value) => {
                  void formik.setFieldValue("diskThreshold", value);
                }}
                firstValue={clusters[0]?.disk_threshold}
                defaultValue={80}
                label="Disk threshold"
                labelForId="diskThreshold"
                value={formik.values.diskThreshold}
              >
                <Input
                  name="diskThreshold"
                  id="diskThreshold"
                  type="number"
                  placeholder="Enter value"
                  min={1}
                  max={100}
                  onBlur={formik.handleBlur}
                  onChange={formik.handleChange}
                  value={formik.values.diskThreshold}
                />
              </BulkConfigurePanelInput>
              <BulkConfigurePanelInput
                areAllValuesEqual={areMemoryThresholdsEqual}
                setValue={(value) => {
                  void formik.setFieldValue("memoryThreshold", value);
                }}
                firstValue={clusters[0]?.memory_threshold}
                defaultValue={80}
                label="Memory threshold"
                labelForId="memoryThreshold"
                value={formik.values.memoryThreshold}
              >
                <Input
                  name="memoryThreshold"
                  id="memoryThreshold"
                  type="number"
                  placeholder="Enter value"
                  min={1}
                  max={100}
                  onBlur={formik.handleBlur}
                  onChange={formik.handleChange}
                  value={formik.values.memoryThreshold}
                />
              </BulkConfigurePanelInput>
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

export default BulkConfigureClusterPanel;
