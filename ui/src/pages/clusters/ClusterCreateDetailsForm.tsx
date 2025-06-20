import { FC } from "react";
import { Input } from "@canonical/react-components";
import { FormikProps } from "formik/dist/types";
import { convertToISOFormat } from "util/helpers";

export interface CreateClusterFormValues {
  clusterName: string;
  expiry: string;
}

export const newTokenPayload = (values: CreateClusterFormValues) => {
  const payload: Record<string, string | undefined> = {
    cluster_name: values.clusterName,
    expiry: convertToISOFormat(values.expiry ?? ""),
  };

  return payload;
};

interface Props {
  formik: FormikProps<CreateClusterFormValues>;
}

const ClusterCreateDetailsForm: FC<Props> = ({ formik }) => {
  return (
    <>
      <Input
        id="name"
        name="clusterName"
        type="text"
        label="Cluster name"
        help="Choose a name for the new cluster."
        placeholder="Enter Name"
        autoFocus
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.clusterName}
        error={formik.touched.clusterName ? formik.errors.clusterName : null}
      />
      <Input
        id="expiry"
        name="expiry"
        type="datetime-local"
        label="Expiry date for join token"
        placeholder="Enter Date"
        onBlur={formik.handleBlur}
        onChange={formik.handleChange}
        value={formik.values.expiry}
        error={formik.touched.expiry ? formik.errors.expiry : null}
      />
    </>
  );
};

export default ClusterCreateDetailsForm;
