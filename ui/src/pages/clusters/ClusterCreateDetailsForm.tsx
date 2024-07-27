import { FC } from "react";
import { Col, Input, Row } from "@canonical/react-components";
import { FormikProps } from "formik/dist/types";
import ScrollableForm from "components/ScrollableForm";
import { convertToISOFormat } from "util/helpers";

export interface CreateClusterFormValues {
  cluster_name?: string;
  expiry?: string;
}

export const newTokenPayload = (values: CreateClusterFormValues) => {
  const payload: Record<string, string | undefined> = {
    cluster_name: values.cluster_name,
    expiry: convertToISOFormat(values.expiry ?? ""),
  };

  return payload;
};

interface Props {
  formik: FormikProps<CreateClusterFormValues>;
}

const ClusterCreateDetailsForm: FC<Props> = ({ formik }) => {
  return (
    <ScrollableForm>
      <Row>
        <Col size={12}>
          <Input
            id="name"
            name="cluster_name"
            type="text"
            label="Cluster name"
            placeholder="Enter Name"
            onBlur={formik.handleBlur}
            onChange={formik.handleChange}
            value={formik.values.cluster_name}
            error={
              formik.touched.cluster_name ? formik.errors.cluster_name : null
            }
          />
          <Input
            id="expiry"
            name="expiry"
            type="datetime-local"
            label="Expiry Date"
            placeholder="Enter Date"
            onBlur={formik.handleBlur}
            onChange={formik.handleChange}
            value={formik.values.expiry}
            error={formik.touched.expiry ? formik.errors.expiry : null}
          />
        </Col>
      </Row>
    </ScrollableForm>
  );
};

export default ClusterCreateDetailsForm;
