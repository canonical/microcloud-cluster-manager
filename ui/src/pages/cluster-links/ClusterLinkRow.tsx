import type { FC } from "react";
import {
  CheckboxInput,
  Chip,
  Icon,
  TableCell,
  TableRow,
  Tooltip,
} from "@canonical/react-components";
import { Link } from "react-router-dom";
import classnames from "classnames";
import type { ClusterConnection } from "util/clusterLinks";
import { getLinkAddresses, linkTypeIcons } from "util/clusterLinks";
import { capitalizeFirstLetter } from "util/helpers";
import ActionsList from "components/ActionsList";
import DeleteClusterLinkBtn from "pages/clusters/actions/DeleteClusterLinkBtn";

interface Props {
  entry: ClusterConnection;
  groupBySource: boolean;
  isSelected: boolean;
  isDisabled: boolean;
  onSelect: () => void;
}

const ClusterLinkRow: FC<Props> = ({
  entry,
  groupBySource,
  isSelected,
  isDisabled,
  onSelect,
}) => {
  const addresses = getLinkAddresses(entry.link);

  return (
    <TableRow
      className={classnames("cluster-link-row", {
        "disabled-row": isDisabled,
      })}
    >
      <TableCell className="select" />
      {groupBySource && <TableCell className="cluster" />}
      <TableCell role="rowheader" className="name" title={entry.link.name}>
        <CheckboxInput
          inline
          label={<div className="u-off-screen">Select {entry.link.name}</div>}
          labelClassName="u-no-margin--bottom"
          checked={isSelected}
          disabled={isDisabled}
          onChange={onSelect}
        />
        <span className="u-truncate">{entry.link.name}</span>
      </TableCell>
      <TableCell className="type" title={entry.link.type}>
        <Chip
          isReadOnly
          isDense
          isInline
          iconName={linkTypeIcons[entry.link.type.toLowerCase()]}
          value={capitalizeFirstLetter(entry.link.type)}
        />
      </TableCell>
      <TableCell className="target" title={entry.target?.name}>
        {entry.target ? (
          <Link to={`/ui/cluster/${entry.target.name}`}>
            {entry.target.name}
          </Link>
        ) : (
          <>
            unknown
            <Tooltip
              message="The linked cluster might not be enrolled in this cluster manager."
              position="right"
            >
              <Icon
                name="help"
                className="u-text--muted cluster-link-unknown-target-icon"
              />
            </Tooltip>
          </>
        )}
      </TableCell>
      <TableCell
        className="description u-truncate"
        title={entry.link.description}
      >
        {entry.link.description}
      </TableCell>
      <TableCell className="used-by">
        {entry.link.used_by?.length ?? 0}
      </TableCell>
      <TableCell className="addresses">
        {addresses.length === 0 ? (
          "unknown"
        ) : (
          <>
            <span className="u-truncate" title={addresses[0]}>
              {addresses[0]}
            </span>
            {addresses.length > 1 && (
              <Tooltip
                message={
                  <ul className="u-no-margin--bottom u-no-padding--left cluster-link-addresses-tooltip">
                    {addresses.map((address) => (
                      <li key={address}>{address}</li>
                    ))}
                  </ul>
                }
                position="btm-center"
              >
                <span className="u-text--muted cluster-link-addresses-more">
                  +{addresses.length - 1}
                </span>
              </Tooltip>
            )}
          </>
        )}
      </TableCell>
      <TableCell className="u-align--right actions" hasOverflow>
        <ActionsList
          items={[
            <DeleteClusterLinkBtn
              cluster={entry.source}
              link={entry.link}
              targetCluster={entry.target}
              hasLabel={false}
              disabled={isDisabled}
              key="delete"
            />,
          ]}
        />
      </TableCell>
    </TableRow>
  );
};

export default ClusterLinkRow;
