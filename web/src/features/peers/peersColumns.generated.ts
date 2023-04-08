// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go


import { ColumnMetaData } from "features/table/types";
import { Peer } from "features/peers/peersTypes";

export const AllPeersColumns: ColumnMetaData<Peer>[] = [
	{
		heading: "Peer Alias",
		type: "AliasCell",
		key: "peerAlias",
		valueType: "string",
		locked: true,
	},
	{
		heading: "Public key",
		type: "LongTextCell",
		key: "pubKey",
		valueType: "string",
	},
	{
		heading: "Torq Alias",
		type: "AliasCell",
		key: "torqNodeAlias",
		valueType: "string",
	},
	{
		heading: "Reconnect",
		type: "TextCell",
		key: "setting",
		valueType: "enum",
		selectOptions: [
			{ label: "Always Reconnect", value: "AlwaysReconnect" },
			{ label: "Disable Reconnect", value: "DisableReconnect" },
		],
	},
	{
		heading: "Status",
		type: "TextCell",
		key: "connectionStatus",
		valueType: "enum",
		selectOptions: [
			{ label: "Disconnected", value: "NodeConnectionStatusDisconnected" },
			{ label: "Connected", value: "NodeConnectionStatusConnected" },
		],
	},
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const PeersSortableColumns: Array<keyof Peer> = [
	"peerAlias",
	"torqNodeAlias",
	"setting",
	"connectionStatus",
];


// DO NOT EDIT THIS FILE...
// This File is generated by go:generate
// For more information look at cmd/torq/gen.go

export const PeersFilterableColumns: Array<keyof Peer> = [
	"peerAlias",
	"torqNodeAlias",
	"setting",
	"connectionStatus",
];