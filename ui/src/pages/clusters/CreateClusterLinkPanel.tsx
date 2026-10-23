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
import { useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import usePanelParams from "context/usePanelParams";
import { createClusterLink } from "api/clusters";
import * as Yup from "yup";
import ClusterSelector from "pages/clusters/ClusterSelector";

interface Props {
  sourceCluster?: string;
}

const CreateClusterLinkPanel: FC<Props> = ({ sourceCluster = "" }) => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface CreateClusterLinkFormValues {
    sourceCluster: string;
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

    createClusterLink(values.sourceCluster, JSON.stringify(payload))
      .then(() => {
        toastNotification.success(<>Created cluster link.</>);
        closePanel();
      })
      .catch((e: Error) => {
        notify.failure(`Failure creating cluster link.`, e);
      })
      .finally(() => {
        queryClient.invalidateQueries({
          queryKey: [queryKeys.clusters],
        });
        formik.setSubmitting(false);
      });
  };

  const ConfigurationSchema = Yup.object().shape({
    sourceCluster: Yup.string().required("Source cluster is required"),
    targetCluster: Yup.string().required("Target cluster is required"),
  });

  const formik = useFormik<CreateClusterLinkFormValues>({
    initialValues: {
      sourceCluster,
      targetCluster: "",
      direction: "bidirectional",
    },
    validationSchema: ConfigurationSchema,
    onSubmit: handleSubmit,
  });

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>
            {sourceCluster
              ? `Create cluster link from ${sourceCluster}`
              : "Create cluster link"}
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
              {!sourceCluster && (
                <ClusterSelector
                  value={formik.values.sourceCluster}
                  setValue={(value) => {
                    formik.setFieldValue("sourceCluster", value);
                  }}
                  ignoreOptions={[formik.values.targetCluster]}
                  selectProps={{
                    id: "sourceCluster",
                    label: "Source cluster",
                    "aria-label": "Source cluster",
                  }}
                />
              )}
              <ClusterSelector
                value={formik.values.targetCluster}
                setValue={(value) => {
                  formik.setFieldValue("targetCluster", value);
                }}
                ignoreOptions={[formik.values.sourceCluster]}
                selectProps={{
                  id: "targetCluster",
                  label: "Target cluster",
                  "aria-label": "Target cluster",
                }}
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

export default CreateClusterLinkPanel;
