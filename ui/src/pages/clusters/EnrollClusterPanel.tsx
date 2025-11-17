import {
  ActionButton,
  Button,
  Form,
  ScrollableContainer,
  SidePanel,
  useNotify,
} from "@canonical/react-components";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import type { FC } from "react";
import * as Yup from "yup";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import type { CreateClusterFormValues } from "pages/clusters/ClusterCreateDetailsForm";
import ClusterCreateDetailsForm, {
  newTokenPayload,
} from "pages/clusters/ClusterCreateDetailsForm";
import { useNavigate } from "react-router-dom";
import { createToken, fetchTokens } from "api/tokens";
import type { TokenState } from "pages/clusters/ClusterList";
import { getDefaultExpiryDate } from "util/createCluster";
import usePanelParams from "context/usePanelParams";
import { fetchClusters } from "api/clusters";

const EnrollClusterPanel: FC = () => {
  const panelParams = usePanelParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const notify = useNotify();

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  const { data: clusters = [] } = useQuery({
    queryKey: [queryKeys.clusters],
    queryFn: fetchClusters,
  });

  const { data: tokens = [] } = useQuery({
    queryKey: [queryKeys.tokens],
    queryFn: fetchTokens,
  });

  const tokenNames = tokens.map((token) => token.cluster_name);
  const clusterNames = clusters.map((cluster) => cluster.name);
  const existingClusterNames = [...tokenNames, ...clusterNames];

  const ClusterSchema = Yup.object().shape({
    clusterName: Yup.string()
      .required("Cluster name is required")
      .notOneOf(existingClusterNames, "A token with this name already exists"),
    expiry: Yup.date()
      .optional()
      .min(new Date(), "Expiry date cannot be in the past"),
  });

  const handleSubmit = (values: CreateClusterFormValues) => {
    const tokenPayload = newTokenPayload(values);

    createToken(JSON.stringify(tokenPayload))
      .then((response) => {
        const state: TokenState = {
          createdCluster: {
            name: values.clusterName,
            token: response.token,
            expiry: values.expiry,
          },
        };

        navigate("/ui/clusters/tokens", { state });
      })
      .catch((e: Error) => {
        notify.failure("Unable to create token.", e);
      })
      .finally(() => {
        void queryClient.invalidateQueries({
          queryKey: [queryKeys.tokens],
        });
        formik.setSubmitting(false);
      });
  };

  const formik = useFormik<CreateClusterFormValues>({
    initialValues: {
      clusterName: "",
      description: "",
      expiry: getDefaultExpiryDate(),
    },
    validationSchema: ClusterSchema,
    onSubmit: handleSubmit,
  });

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>Enroll cluster</SidePanel.HeaderTitle>
        </SidePanel.Header>
        <NotificationRow className="u-no-padding" />
        <SidePanel.Content className="u-no-padding">
          <ScrollableContainer
            dependencies={[notify.notification]}
            belowIds={["panel-footer"]}
          >
            <Form onSubmit={() => void formik.submitForm()} className="form">
              <ClusterCreateDetailsForm formik={formik} />
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
            Create join token
          </ActionButton>
        </SidePanel.Footer>
      </SidePanel>
    </>
  );
};

export default EnrollClusterPanel;
