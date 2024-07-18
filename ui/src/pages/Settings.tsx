import { FC } from "react";
import { MainTable, Row } from "@canonical/react-components";
import BaseLayout from "components/BaseLayout";
import { queryKeys } from "util/queryKeys";
import {
  fetchManagerConfigOptions,
  fetchMemberConfigOptions,
} from "api/settings";
import { useQuery } from "@tanstack/react-query";
import SettingForm from "./settings/SettingForm";
import { MainTableRow } from "@canonical/react-components/dist/components/MainTable/MainTable";
import { ManagerOptions, MemberOptions } from "types/config";

const Settings: FC = () => {
  const { data: managerConfigOptions } = useQuery({
    queryKey: [queryKeys.managerConfigOptions],
    queryFn: fetchManagerConfigOptions,
  });

  const { data: memberConfigOptions = [] } = useQuery({
    queryKey: [queryKeys.memberConfigOptions],
    queryFn: fetchMemberConfigOptions,
  });

  const defaultManagerConfigs: ManagerOptions["config"] = {
    "oidc.issuer": "",
    "oidc.client.id": "",
    "oidc.audience": "",
    "global.address": "",
  };

  const defaultMemberConfigs: MemberOptions["config"] = {
    https_address: "",
    external_address: "",
  };

  const headers = [
    { content: "Scope", className: "scope" },
    { content: "Key", className: "key" },
    { content: "Value" },
  ];

  const generateManagerConfigRows = () => {
    const configKeys = Object.keys(defaultManagerConfigs);

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
                  managerConfigOptions?.config[key] ||
                  defaultManagerConfigs[key]
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

    return rows;
  };

  const generateMemberConfigRows = () => {
    const allMemberConfigRows: MainTableRow[] = [];

    for (const memberConfig of memberConfigOptions) {
      const configKeys = Object.keys(defaultMemberConfigs) as Array<
        keyof MemberOptions["config"]
      >;

      const currentMemberConfigRows = configKeys.map((key, index) => {
        const memberConfigScope =
          index === 0 ? `Member (${memberConfig.target})` : "";
        const memberConfigKey = `${memberConfig.target}.${key}`;

        return {
          columns: [
            {
              content: (
                <h2
                  className="p-heading--5 u-truncate"
                  title={memberConfigScope}
                >
                  {memberConfigScope}
                </h2>
              ),
              role: "cell",
              className: "scope",
              "aria-label": "Scope",
            },
            {
              content: (
                <div className="key-cell u-truncate" title={memberConfigKey}>
                  {memberConfigKey}
                </div>
              ),
              role: "cell",
              className: "key",
              "aria-label": "Key",
            },
            {
              content: (
                <SettingForm
                  configField={key}
                  value={memberConfig.config[key] || defaultMemberConfigs[key]}
                  isLast={index === length - 1}
                  member={memberConfig.target}
                />
              ),
              role: "cell",
              "aria-label": "Value",
              className: "u-vertical-align-middle",
            },
          ],
        };
      });

      allMemberConfigRows.push(...currentMemberConfigRows);
    }

    return allMemberConfigRows;
  };

  const rows = [...generateManagerConfigRows(), ...generateMemberConfigRows()];

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
