// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go


import { ColumnMetaData } from "features/table/types";
import { ChannelClosed } from "features/channelsClosed/channelsClosedTypes";

export const AllChannelClosedColumns: ColumnMetaData<ChannelClosed>[] = [
	{
		heading: "Peer Alias",
		type: "AliasCell",
		key: "peerAlias",
		valueType: "string",
		locked: true,
	},
	{
		heading: "Short Channel ID",
		type: "LongTextCell",
		key: "shortChannelId",
		valueType: "string",
	},
	{
		heading: "Capacity",
		type: "NumericCell",
		key: "capacity",
		valueType: "number",
	},
	{
		heading: "LND Short Channel ID",
		type: "LongTextCell",
		key: "lndShortChannelId",
		valueType: "string",
	},
	{
		heading: "Funding Transaction",
		type: "LongTextCell",
		key: "fundingTransactionHash",
		valueType: "string",
	},
	{
		heading: "Funding BlockHeight",
		type: "NumericCell",
		key: "fundingBlockHeight",
		valueType: "number",
	},
	{
		heading: "Funding BlockHeight Delta",
		type: "NumericCell",
		key: "fundingBlockHeightDelta",
		valueType: "number",
	},
	{
		heading: "Funding Date",
		type: "DateCell",
		key: "fundedOn",
		valueType: "date",
	},
	{
		heading: "Funding Date Delta (Seconds)",
		type: "DurationCell",
		key: "fundedOnSecondsDelta",
		valueType: "duration",
	},
	{
		heading: "Closing Transaction",
		type: "LongTextCell",
		key: "closingTransactionHash",
		valueType: "string",
	},
	{
		heading: "Closing BlockHeight",
		type: "NumericCell",
		key: "closingBlockHeight",
		valueType: "number",
	},
	{
		heading: "Closing BlockHeight Delta",
		type: "NumericCell",
		key: "closingBlockHeightDelta",
		valueType: "number",
	},
	{
		heading: "Closing Date",
		type: "DateCell",
		key: "closedOn",
		valueType: "date",
	},
	{
		heading: "Closing Date Delta (Seconds)",
		type: "DurationCell",
		key: "closedOnSecondsDelta",
		valueType: "duration",
	},
	{
		heading: "Node Name",
		type: "TextCell",
		key: "nodeName",
		valueType: "string",
	},
	{
		heading: "Public key",
		type: "LongTextCell",
		key: "pubKey",
		valueType: "string",
	},
	{
		heading: "Status",
		type: "TextCell",
		key: "status",
		valueType: "enum",
		selectOptions: [
			{ label: "Opening", value: "Opening" },
			{ label: "Open", value: "Open" },
			{ label: "Closing", value: "Closing" },
			{ label: "Cooperative Closed", value: "Cooperative Closed" },
			{ label: "Local Force Closed", value: "Local Force Closed" },
			{ label: "Remote Force Closed", value: "Remote Force Closed" },
			{ label: "Breach Closed", value: "Breach Closed" },
			{ label: "Funding Cancelled Closed", value: "Funding Cancelled Closed" },
			{ label: "Abandoned Closed", value: "Abandoned Closed" },
		],
	},
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const ChannelsClosedSortableColumns: Array<keyof ChannelClosed> = [
	"peerAlias",
	"shortChannelId",
	"capacity",
	"lndShortChannelId",
	"fundingTransactionHash",
	"fundingBlockHeight",
	"fundingBlockHeightDelta",
	"fundedOn",
	"fundedOnSecondsDelta",
	"closingTransactionHash",
	"closingBlockHeight",
	"closingBlockHeightDelta",
	"closedOn",
	"closedOnSecondsDelta",
	"nodeName",
	"status",
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const ChannelsClosedFilterableColumns: Array<keyof ChannelClosed> = [
	"peerAlias",
	"shortChannelId",
	"capacity",
	"lndShortChannelId",
	"fundingTransactionHash",
	"fundingBlockHeight",
	"fundingBlockHeightDelta",
	"fundedOn",
	"fundedOnSecondsDelta",
	"closingTransactionHash",
	"closingBlockHeight",
	"closingBlockHeightDelta",
	"closedOn",
	"closedOnSecondsDelta",
	"nodeName",
	"status",
];