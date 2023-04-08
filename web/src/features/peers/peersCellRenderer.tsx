import { ColumnMetaData } from "features/table/types";
import { ConnectionStatus, NodeConnectionSetting, Peer } from "features/peers/peersTypes";
import DefaultCellRenderer from "features/table/DefaultCellRenderer";
import useTranslations from "services/i18n/useTranslations";
import TextCell from "components/table/cells/text/TextCell";
import PeersAliasCell from "components/table/cells/peersCell/PeersAliasCell";

export default function peerCellRenderer(
  row: Peer,
  rowIndex: number,
  column: ColumnMetaData<Peer>,
  columnIndex: number,
  isTotalsRow?: boolean,
  maxRow?: Peer
): JSX.Element {
  const { t } = useTranslations();

  if (column.key === "connectionStatus") {
    return (
      <TextCell
        current={t.peersPage[ConnectionStatus[row.connectionStatus]]}
        key={column.key.toString() + rowIndex}
        totalCell={isTotalsRow}
      />
    );
  }

  if (column.key === "setting") {
    return (
      <TextCell
        current={t.peersPage[NodeConnectionSetting[row.setting]]}
        key={column.key.toString() + rowIndex}
        totalCell={isTotalsRow}
      />
    );
  }

  if (column.key === "peerAlias") {
    return (
      <PeersAliasCell
        alias={row.peerAlias}
        peerNodeId={row.nodeId}
        torqNodeId={row.torqNodeId}
        connectionStatus={row.connectionStatus}
      />
    );
  }

  return DefaultCellRenderer(row, rowIndex, column, columnIndex, false, maxRow);
}
