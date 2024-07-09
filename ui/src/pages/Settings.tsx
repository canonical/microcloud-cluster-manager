import { FC } from "react";
import { MainTable, Row } from "@canonical/react-components";
import BaseLayout from "components/BaseLayout";
import { queryKeys } from "util/queryKeys";
import { fetchConfigOptions } from "api/settings";
import { useQuery } from "@tanstack/react-query";
import SettingForm from "./settings/SettingForm";

const Settings: FC = () => {
  const { data: configOptions } = useQuery({
    queryKey: [queryKeys.configOptions],
    queryFn: fetchConfigOptions,
  });

  const defaultConfig = {
    "oidc.issuer": "",
    "oidc.client.id": "",
    "oidc.audience": "",
    "global.address": "",
  };

  const headers = [
    { content: "Scope", className: "scope" },
    { content: "Key", className: "key" },
    { content: "Value" },
  ];

  const configKeys = Object.keys(defaultConfig);
  // const configKeys = Object.keys(defaultConfig);

  const rows = configKeys.map((key, index) => {
    return {
      columns: [
        {
          content: (
            <h2 className="p-heading--5">{index === 0 ? "Cluster" : ""}</h2>
          ),
          role: "cell",
          className: "scope",
          "aria-label": "Scope",
        },
        {
          content: <div className="key-cell">{key}</div>,
          role: "cell",
          className: "key",
          "aria-label": "Key",
        },
        {
          content: (
            <SettingForm
              configField={key}
              value={
                configOptions?.config[key] ||
                defaultConfig[key as keyof typeof defaultConfig]
              }
              isLast={index === length - 1}
            />
          ),
          role: "cell",
          "aria-label": "Value",
          className: "u-vertical-align-middle",
        },
      ],
    };
  });

  return (
    <BaseLayout title="Settings">
      <Row>
        <div className="settings">
          <MainTable
            id="settings-table"
            headers={headers}
            rows={rows}
            emptyStateMsg="No settings to display"
          />
        </div>
      </Row>
    </BaseLayout>
  );
};

export default Settings;
