import {
  ActionButton,
  Button,
  CheckboxInput,
  CodeSnippet,
  Form,
  Input,
  ScrollableContainer,
  Select,
  SidePanel,
  useNotify,
  useToastNotification,
} from "@canonical/react-components";
import { useQueryClient } from "@tanstack/react-query";
import { type FC, useEffect, useState } from "react";
import { useFormik } from "formik";
import { queryKeys } from "util/queryKeys";
import NotificationRow from "components/NotificationRow";
import usePanelParams from "context/usePanelParams";
import { useCluster, useClusterLinks } from "context/useClusters";
import { createClusterLink } from "api/clusters";
import type { ClusterLink, ClusterLinkPost } from "types/cluster";
import * as Yup from "yup";
import ClusterSelector from "pages/clusters/ClusterSelector";
import { getCertificateFingerprint } from "util/clusterLinks";
import CopyToClipboard from "components/CopyToClipboard";

const CreateClusterLinkPanel: FC = () => {
  const panelParams = usePanelParams();
  const queryClient = useQueryClient();
  const notify = useNotify();
  const toastNotification = useToastNotification();

  const clusterName = panelParams.cluster ?? "";

  const { data: cluster } = useCluster(clusterName);
  const clusterLinkName = panelParams.clusterLink ?? "";
  const { data: clusterLinks = [], isLoading: isPendingLinksLoading } =
    useClusterLinks(clusterName, !!clusterLinkName);
  const pendingClusterLink = clusterLinks.find(
    (link) => link.name === clusterLinkName,
  );
  const clusterLinkCertificate =
    pendingClusterLink?.config?.["volatile.pending_certificate"] ?? "";

  const [certificateFingerprint, setCertificateFingerprint] = useState("");
  const [pendingCertificateReadFailed, setPendingCertificateReadFailed] =
    useState(false);

  useEffect(() => {
    if (!clusterLinkName) {
      return;
    }

    if (isPendingLinksLoading) {
      return;
    }

    if (!clusterLinkCertificate) {
      setPendingCertificateReadFailed(false);
      return;
    }

    let cancelled = false;
    setCertificateFingerprint("");
    setPendingCertificateReadFailed(false);
    void getCertificateFingerprint(clusterLinkCertificate)
      .then((fingerprint) => {
        if (!cancelled) {
          setCertificateFingerprint(fingerprint);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setPendingCertificateReadFailed(true);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [clusterLinkCertificate, clusterLinkName, isPendingLinksLoading]);

  const getPendingCertificateMessage = (): string => {
    if (isPendingLinksLoading) {
      return "Loading pending certificate...";
    }

    if (!clusterLinkCertificate) {
      return "The pending link certificate is unavailable.";
    }

    if (pendingCertificateReadFailed) {
      return "The pending link certificate could not be read.";
    }

    return "Loading pending certificate...";
  };

  const closePanel = () => {
    panelParams.clear();
    notify.clear();
  };

  interface CreateClusterLinkFormValues {
    targetCluster: string;
    type: ClusterLink["type"];
    useDifferentTargetAddress: boolean;
    targetAddress: string;
    certificateConfirmed: boolean;
  }

  const isPendingLinkFingerprintMissing =
    !!clusterLinkName && !certificateFingerprint;

  const isAwaitingUserConfirmation = (
    type: CreateClusterLinkFormValues["type"],
    fingerprint: string,
    certificateConfirmed: boolean,
  ): boolean => {
    return type === "public" && !!fingerprint && !certificateConfirmed;
  };

  const handleSubmit = async (
    values: CreateClusterLinkFormValues,
  ): Promise<void> => {
    const checkingCertificate =
      values.type === "public" && !certificateFingerprint;
    const isSubmittedVerificationIncomplete =
      isPendingLinkFingerprintMissing ||
      isAwaitingUserConfirmation(
        values.type,
        certificateFingerprint,
        values.certificateConfirmed,
      );
    if (isSubmittedVerificationIncomplete) {
      return;
    }

    const payload: ClusterLinkPost = {
      target_cluster: values.targetCluster,
      name: values.targetCluster,
      auth_groups: [],
      description: "",
      type: values.type,
    };
    if (values.type === "public") {
      if (values.useDifferentTargetAddress) {
        payload.remote_address = values.targetAddress;
      }

      if (certificateFingerprint) {
        payload.fingerprint = certificateFingerprint;
      }
    }

    return createClusterLink(clusterName, JSON.stringify(payload))
      .then((fingerprint) => {
        if (checkingCertificate) {
          if (typeof fingerprint !== "string" || !fingerprint) {
            throw new Error("The target certificate fingerprint was missing");
          }

          setCertificateFingerprint(fingerprint);
          void formik.setFieldValue("certificateConfirmed", false);
          return;
        }

        toastNotification.success(<>Created cluster link.</>);
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
  });

  const formik = useFormik<CreateClusterLinkFormValues>({
    initialValues: {
      targetCluster: clusterLinkName,
      type: clusterLinkName ? "public" : "bidirectional",
      useDifferentTargetAddress: false,
      targetAddress: "",
      certificateConfirmed: false,
    },
    validationSchema: ConfigurationSchema,
    enableReinitialize: true,
    onSubmit: handleSubmit,
  });

  const resetCertificateVerification = () => {
    setCertificateFingerprint("");
    void formik.setFieldValue("certificateConfirmed", false);
  };

  const resetTargetAddress = (resetAddressToggle = true) => {
    if (resetAddressToggle) {
      void formik.setFieldValue("useDifferentTargetAddress", false);
    }
    void formik.setFieldValue("targetAddress", "");
    resetCertificateVerification();
  };

  const isCertificateVerificationIncomplete =
    isPendingLinkFingerprintMissing ||
    isAwaitingUserConfirmation(
      formik.values.type,
      certificateFingerprint,
      formik.values.certificateConfirmed,
    );
  const isAlternateTargetAddressInvalid =
    formik.values.useDifferentTargetAddress &&
    (!formik.values.targetAddress || !!formik.errors.targetAddress);
  const isSubmitDisabled =
    formik.isSubmitting ||
    !formik.values.targetCluster ||
    isAlternateTargetAddressInvalid ||
    isCertificateVerificationIncomplete;

  if (!cluster) {
    return null;
  }

  return (
    <>
      <SidePanel>
        <SidePanel.Header>
          <SidePanel.HeaderTitle>
            {certificateFingerprint
              ? `Verify certificate for ${formik.values.targetCluster}`
              : `Create cluster link from ${cluster.name}`}
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
              {!certificateFingerprint && !clusterLinkName && (
                <>
                  <ClusterSelector
                    value={formik.values.targetCluster}
                    setValue={(value) => {
                      void formik.setFieldValue("targetCluster", value);
                      resetTargetAddress();
                    }}
                    ignoreOptions={[cluster.name]}
                  />
                  <Select
                    id="type"
                    name="type"
                    label="Type"
                    value={formik.values.type}
                    onChange={(e) => {
                      void formik.setFieldValue("type", e.target.value);
                      resetTargetAddress();
                    }}
                    options={[
                      { value: "bidirectional", label: "Bidirectional" },
                      { value: "unidirectional", label: "Unidirectional" },
                      { value: "public", label: "Public" },
                    ]}
                  />
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
                            resetCertificateVerification();
                          }}
                          error={
                            formik.touched.targetAddress &&
                            formik.errors.targetAddress
                          }
                        />
                      )}
                    </>
                  )}
                </>
              )}
              {clusterLinkName && !certificateFingerprint && (
                <p role="status">{getPendingCertificateMessage()}</p>
              )}
              {certificateFingerprint && (
                <>
                  <p>
                    Compare this SHA-256 fingerprint with the certificate on the
                    target cluster before continuing.
                  </p>
                  <CodeSnippet
                    className="cluster-link-fingerprint"
                    blocks={[
                      {
                        title: "Certificate fingerprint",
                        wrapLines: true,
                        code: (
                          <div className="command-wrapper">
                            <span
                              id="certificate-fingerprint"
                              className="command"
                            >
                              {certificateFingerprint}
                            </span>
                            <CopyToClipboard
                              value={certificateFingerprint}
                              tooltipMessage="Copy fingerprint"
                            />
                          </div>
                        ),
                      },
                    ]}
                  />
                  <CheckboxInput
                    id="confirm-target-certificate"
                    className="cluster-link-address-checkbox"
                    label="Confirm the certificate fingerprint before creating the link"
                    checked={formik.values.certificateConfirmed}
                    onChange={() => {
                      void formik.setFieldValue(
                        "certificateConfirmed",
                        !formik.values.certificateConfirmed,
                      );
                    }}
                  />
                </>
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
            disabled={isSubmitDisabled}
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
