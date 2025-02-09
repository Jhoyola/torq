package lnd

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/internal/services_helpers"
	"github.com/lncapital/torq/proto/lnrpc/zpay32"

	"github.com/lncapital/torq/proto/lnrpc"

	"github.com/lncapital/torq/internal/cache"
	"github.com/lncapital/torq/internal/core"
)

const streamLndMaxInvoices = 1000

type invoicesClient interface {
	SubscribeInvoices(ctx context.Context, in *lnrpc.InvoiceSubscription,
		opts ...grpc.CallOption) (lnrpc.Lightning_SubscribeInvoicesClient, error)
	ListInvoices(ctx context.Context, in *lnrpc.ListInvoiceRequest,
		opts ...grpc.CallOption) (*lnrpc.ListInvoiceResponse, error)
}

type Invoice struct {
	InvoiceId int `db:"invoice_id" json:"invoiceId"`

	/*
	   An optional memo to attach along with the invoice. Used for record keeping
	   purposes for the invoice's creator, and will also be set in the description
	   field of the encoded payment request if the description_hash field is not
	   being used.
	*/
	Memo string `db:"memo" json:"memo"`

	/*
	   The hex-encoded preimage (32 byte) which will allow settling an incoming
	   HTLC payable to this preimage.
	*/
	RPreimage string `db:"r_preimage" json:"r_preimage"`

	/*
	   The hash of the preimage.
	*/
	RHash string `db:"r_hash" json:"r_hash"`

	// The value of the invoice
	ValueMsat int64 `db:"value_msat" json:"value_msat"`

	// When this invoice was created
	CreationDate time.Time `db:"creation_date" json:"creation_date"`

	// When this invoice was settled
	SettleDate time.Time `db:"settle_date" json:"settle_date"`

	/*
	   A bare-bones invoice for a payment within the Lightning Network. With the
	   details of the invoice, the sender has all the data necessary to send a
	   payment to the recipient.
	*/
	PaymentRequest string `db:"payment_request" json:"payment_request"`

	/*
	   A bare-bones invoice for a payment within the Lightning Network. With the
	   details of the invoice, the sender has all the data necessary to send a
	   payment to the recipient.
	*/
	Destination string `db:"destination_pub_key" json:"destination_pub_key"`

	/*
	   Hash (SHA-256) of a description of the payment. Used if the description of
	   payment (memo) is too long to naturally fit within the description field
	   of an encoded payment request.
	*/
	DescriptionHash []byte `db:"description_hash" json:"description_hash"`

	// Payment request expiry time in seconds. Default is 3600 (1 hour).
	Expiry int64 `db:"expiry" json:"expiry"`

	// Fallback on-chain address.
	FallbackAddr string `db:"fallback_addr" json:"fallback_addr"`

	// Delta to use for the time-lock of the CLTV extended to the final hop.
	CltvExpiry uint64 `db:"cltv_expiry" json:"cltv_expiry"`

	/*
	   Route hints that can each be individually used to assist in reaching the
	   invoice's destination.
	*/
	//repeated RouteHint route_hints = 14;
	RouteHints []byte `db:"route_hints" json:"route_hints"`

	// Whether this invoice should include routing hints for private channels.
	Private bool `db:"private" json:"private"`

	/*
	   The "add" index of this invoice. Each newly created invoice will increment
	   this index making it monotonically increasing. Callers to the
	   SubscribeInvoices call can use this to instantly get notified of all added
	   invoices with an add_index greater than this one.
	*/
	AddIndex uint64 `db:"add_index" json:"add_index"`

	/*
	   The "settle" index of this invoice. Each newly settled invoice will
	   increment this index making it monotonically increasing. Callers to the
	   SubscribeInvoices call can use this to instantly get notified of all
	   settled invoices with an settle_index greater than this one.
	*/
	SettleIndex uint64 `db:"settle_index" json:"settle_index"`

	/*
	   The amount that was accepted for this invoice, in satoshis. This will ONLY
	   be set if this invoice has been settled. We provide this field as if the
	   invoice was created with a zero value, then we need to record what amount
	   was ultimately accepted. Additionally, it's possible that the sender paid
	   MORE that was specified in the original invoice. So we'll record that here
	   as well.
	*/
	AmtPaidSat int64 `db:"amt_paid_sat" json:"amt_paid_sat"`

	/*
	   The amount that was accepted for this invoice, in millisatoshis. This will
	   ONLY be set if this invoice has been settled. We provide this field as if
	   the invoice was created with a zero value, then we need to record what
	   amount was ultimately accepted. Additionally, it's possible that the sender
	   paid MORE that was specified in the original invoice. So we'll record that
	   here as well.
	*/
	AmtPaidMsat int64 `db:"amt_paid_msat" json:"amt_paid_msat"`

	InvoiceState string `db:"invoice_state" json:"invoice_state"`
	//OPEN = 0;
	//SETTLED = 1;
	//CANCELED = 2;
	//ACCEPTED = 3;

	// List of HTLCs paying to this invoice [EXPERIMENTAL].
	Htlcs []byte `db:"htlcs" json:"htlcs"`
	//repeated InvoiceHTLC htlcs = 22;

	// List of features advertised on the invoice.
	//map<uint32, Feature> features = 24;
	// features []*lnrpc.Feature
	Features []byte `db:"features" json:"features"`

	/*
	   Indicates if this invoice was a spontaneous payment that arrived via keysend
	   [EXPERIMENTAL].
	*/
	IsKeysend bool `db:"is_keysend" json:"is_keysend"`

	/*
	   The payment address of this invoice. This value will be used in MPP
	   payments, and also for newer invoices that always require the MPP payload
	   for added end-to-end security.
	*/
	PaymentAddr string `db:"payment_addr" json:"payment_addr"`

	/*
	   Signals whether this is an AMP invoice.
	*/
	IsAmp bool `db:"is_amp" json:"is_amp"`

	/*
	   [EXPERIMENTAL]:
	   Maps a 32-byte hex-encoded set ID to the sub-invoice AMP state for the
	   given set ID. This field is always populated for AMP invoices, and can be
	   used alongside LookupInvoice to obtain the HTLC information related to a
	   given sub-invoice.
	*/
	//map<string, AMPInvoiceState> amp_invoice_state = 28;
	AmpInvoiceState   []byte    `db:"amp_invoice_state" json:"amp_invoice_state"`
	DestinationNodeId *int      `db:"destination_node_id" json:"destinationNodeId"`
	NodeId            int       `db:"node_id" json:"nodeId"`
	ChannelId         *int      `db:"channel_id" json:"channelId"`
	CreatedOn         time.Time `db:"created_on" json:"created_on"`
	UpdatedOn         time.Time `db:"updated_on" json:"updated_on"`
}

func fetchLastInvoiceIndexes(db *sqlx.DB, nodeId int) (addIndex uint64, settleIndex uint64, err error) {
	// index starts at 1
	sqlLatest := `select coalesce(max(add_index),1), coalesce(max(settle_index),1) from invoice where node_id = $1;`

	row := db.QueryRow(sqlLatest, nodeId)
	err = row.Scan(&addIndex, &settleIndex)

	if err != nil {
		log.Error().Msgf("getting max invoice indexes: %v", err)
		return 0, 0, errors.Wrap(err, "getting max invoice indexes")
	}

	return addIndex, settleIndex, nil
}

func SubscribeAndStoreInvoices(ctx context.Context, client invoicesClient, db *sqlx.DB,
	nodeSettings cache.NodeSettingsCache) {

	serviceType := services_helpers.LndServiceInvoiceStream

	bootStrapping := true
	importCounter := 0

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		default:
		}

		// Get the latest settle and add index to prevent duplicate entries.
		addIndex, _, err := fetchLastInvoiceIndexes(db, nodeSettings.NodeId)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to obtain last know invoice for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		listInvoiceResponse, err := client.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{
			NumMaxInvoices: streamLndMaxInvoices,
			IndexOffset:    addIndex,
		})
		if err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf("Failed to obtain list invoice for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		if bootStrapping {
			importCounter = importCounter + len(listInvoiceResponse.Invoices)
			if len(listInvoiceResponse.Invoices) >= streamLndMaxInvoices {
				log.Info().Msgf("Still running bulk import of invoices (%v)", importCounter)
			}
			cache.SetInitializingNodeServiceState(serviceType, nodeSettings.NodeId)
		}
		for _, invoice := range listInvoiceResponse.Invoices {
			processInvoice(invoice, nodeSettings, db, bootStrapping)
		}
		if bootStrapping && len(listInvoiceResponse.Invoices) < streamLndMaxInvoices {
			bootStrapping = false
			log.Info().Msgf("Bulk import of invoices done (%v)", importCounter)
			cache.SetActiveNodeServiceState(serviceType, nodeSettings.NodeId)
			break
		}
	}

	for {
		select {
		case <-ctx.Done():
			cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		default:
		}

		// Get the latest settle and add index to prevent duplicate entries.
		addIndex, settleIndex, err := fetchLastInvoiceIndexes(db, nodeSettings.NodeId)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to obtain last invoice index for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		streamCtx, cancel := context.WithCancel(ctx)
		stream, err := client.SubscribeInvoices(streamCtx, &lnrpc.InvoiceSubscription{
			AddIndex:    addIndex,
			SettleIndex: settleIndex,
		})
		if err != nil {
			cancel()
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf("Failed to obtain Invoices stream for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}

		invoice, err := stream.Recv()
		if err != nil {
			cancel()
			if errors.Is(ctx.Err(), context.Canceled) {
				cache.SetInactiveNodeServiceState(serviceType, nodeSettings.NodeId)
				return
			}
			log.Error().Err(err).Msgf(
				"Failed to obtain receive Invoices from the stream for nodeId: %v", nodeSettings.NodeId)
			cache.SetFailedNodeServiceState(serviceType, nodeSettings.NodeId)
			return
		}
		processInvoice(invoice, nodeSettings, db, bootStrapping)
		cancel()
	}
}

func processInvoice(lndInvoice *lnrpc.Invoice, nodeSettings cache.NodeSettingsCache, db *sqlx.DB, bootStrapping bool) {
	invoiceEvent := core.InvoiceEvent{
		EventData: core.EventData{
			EventTime: time.Now().UTC(),
			NodeId:    nodeSettings.NodeId,
		},
	}

	var destinationPublicKey = ""
	var destinationNodeId *int
	// if empty payment request lndInvoice is likely keysend
	if lndInvoice.PaymentRequest != "" {
		// Check the running nodes network. Currently we assume we are running on Bitcoin mainnet
		nodeNetwork := getNodeNetwork(lndInvoice.PaymentRequest)

		inva, err := zpay32.Decode(lndInvoice.PaymentRequest, nodeNetwork)
		if err != nil {
			log.Error().Msgf("Subscribe and store invoices - decode payment request: %v", err)
		} else {
			destinationPublicKey = fmt.Sprintf("%x", inva.Destination.SerializeCompressed())
			destinationNodeIdValue := cache.GetChannelPeerNodeIdByPublicKey(destinationPublicKey, nodeSettings.Chain, nodeSettings.Network)
			destinationNodeId = &destinationNodeIdValue
			invoiceEvent.DestinationNodeId = destinationNodeId
		}
	}

	invoiceId, err := getInvoiceIdByAddIndex(db, lndInvoice.AddIndex)
	if err != nil {
		log.Error().Err(err).Msg("Checking for existing invoice")
		return
	}
	invoice, err := constructInvoice(lndInvoice, destinationPublicKey, destinationNodeId, nodeSettings.NodeId)
	if err != nil {
		log.Error().Err(err).Msg("Constructing invoice")
		return
	}

	invoiceEvent = completeInvoiceEvent(lndInvoice, invoice, invoiceEvent)
	defer func() {
		if !bootStrapping {
			ProcessInvoiceEvent(invoiceEvent)
		}
	}()

	if invoiceId == 0 {
		err = insertInvoice(db, invoice)
		if err != nil {
			log.Error().Err(err).Msg("Inserting invoice")
		}
		return
	}
	invoice.InvoiceId = invoiceId
	err = updateInvoice(db, invoice)
	if err != nil {
		log.Error().Err(err).Msg("Updating invoice")
	}
}

func completeInvoiceEvent(lndInvoice *lnrpc.Invoice,
	invoice Invoice,
	invoiceEvent core.InvoiceEvent) core.InvoiceEvent {

	invoiceEvent.State = lndInvoice.State
	invoiceEvent.AddIndex = lndInvoice.AddIndex
	invoiceEvent.ValueMSat = uint64(lndInvoice.ValueMsat)
	// Add other info for settled and accepted states
	//	Invoice_OPEN     = 0
	//	Invoice_SETTLED  = 1
	//	Invoice_CANCELED = 2
	//	Invoice_ACCEPTED = 3
	if lndInvoice.State == 1 || lndInvoice.State == 3 {
		invoiceEvent.AmountPaidMsat = uint64(lndInvoice.AmtPaidMsat)
		invoiceEvent.SettledDate = time.Unix(lndInvoice.SettleDate, 0)
	}
	if invoice.ChannelId != nil {
		invoiceEvent.ChannelId = *invoice.ChannelId
	}
	return invoiceEvent
}

// getNodeNetwork
// Obtained from invoice.PaymentRequest
// MainNetParams           bc
// RegressionNetParams     bcrt
// SigNetParams            tbs
// TestNet3Params          tb
// SimNetParams            sb
// Example: invoice.PaymentRequest = lnbcrt500u1p3vmd6upp5y7ndr6dmyehql..."
//   - First two characters should be "ln"
//   - Next 2+2 characters determine the network
//   - Here the network is RegressionNetParams - bcrt
//
// This values come from chaincfg.<Params>.Bech32HRPSegwit
func getNodeNetwork(pmntReq string) *chaincfg.Params {
	nodeNetworkPrefix := pmntReq[2:4]
	nodeNetworkSuffix := ""

	switch {
	case nodeNetworkPrefix == "bc":
		nodeNetworkSuffix = pmntReq[4:6]
		if nodeNetworkSuffix == "rt" {
			return &chaincfg.RegressionNetParams
		} else {
			return &chaincfg.MainNetParams
		}
	case nodeNetworkPrefix == "tb":
		nodeNetworkSuffix = pmntReq[4:5]
		if nodeNetworkSuffix == "s" {
			return &chaincfg.SigNetParams
		} else {
			return &chaincfg.TestNet3Params
		}
	case nodeNetworkPrefix == "sb":
		return &chaincfg.SimNetParams
	default:
		return &chaincfg.MainNetParams
	}
}

func constructInvoice(invoice *lnrpc.Invoice, destination string, destinationNodeId *int, nodeId int) (Invoice, error) {
	rhJson, err := json.Marshal(invoice.RouteHints)
	if err != nil {
		log.Error().Msgf("constructInvoice - json marshal route hints: %v", err)
		return Invoice{}, errors.Wrapf(err, "constructInvoice - json marshal route hints")
	}

	htlcJson, err := json.Marshal(invoice.Htlcs)
	if err != nil {
		log.Error().Msgf("constructInvoice - json marshal htlcs: %v", err)
		return Invoice{}, errors.Wrapf(err, "constructInvoice - json marshal htlcs")
	}

	featuresJson, err := json.Marshal(invoice.Features)
	if err != nil {
		log.Error().Msgf("constructInvoice - json marshal features: %v", err)
		return Invoice{}, errors.Wrapf(err, "constructInvoice - json marshal features")
	}

	aisJson, err := json.Marshal(invoice.AmpInvoiceState)
	if err != nil {
		log.Error().Msgf("")
		return Invoice{}, errors.Wrapf(err, "constructInvoice - json marshal amp invoice state")
	}

	var channelId *int
	if len(invoice.Htlcs) > 0 {
		channelId = getChannelIdByLndShortChannelId(invoice.Htlcs[len(invoice.Htlcs)-1].ChanId)
	}

	return Invoice{
		Memo:              invoice.Memo,
		RPreimage:         hex.EncodeToString(invoice.RPreimage),
		RHash:             hex.EncodeToString(invoice.RHash),
		ValueMsat:         invoice.ValueMsat,
		CreationDate:      time.Unix(invoice.CreationDate, 0).UTC(),
		SettleDate:        time.Unix(invoice.SettleDate, 0).UTC(),
		PaymentRequest:    invoice.PaymentRequest,
		Destination:       destination,
		DescriptionHash:   invoice.DescriptionHash,
		Expiry:            invoice.Expiry,
		FallbackAddr:      invoice.FallbackAddr,
		CltvExpiry:        invoice.CltvExpiry,
		RouteHints:        rhJson,
		Private:           false,
		AddIndex:          invoice.AddIndex,
		SettleIndex:       invoice.SettleIndex,
		AmtPaidSat:        invoice.AmtPaidSat,
		AmtPaidMsat:       invoice.AmtPaidMsat,
		InvoiceState:      invoice.State.String(), // ,
		Htlcs:             htlcJson,
		Features:          featuresJson,
		IsKeysend:         invoice.IsKeysend,
		PaymentAddr:       hex.EncodeToString(invoice.PaymentAddr),
		IsAmp:             invoice.IsAmp,
		AmpInvoiceState:   aisJson,
		DestinationNodeId: destinationNodeId,
		NodeId:            nodeId,
		ChannelId:         channelId,
		CreatedOn:         time.Now().UTC(),
		UpdatedOn:         time.Now().UTC(),
	}, nil
}

func getInvoiceIdByAddIndex(db *sqlx.DB, addIndex uint64) (int, error) {
	var invoiceId int
	err := db.Get(&invoiceId, `SELECT invoice_id FROM invoice WHERE add_index=$1;`, addIndex)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, errors.Wrapf(err, "Obtaining existing invoice for addIndex: %v", addIndex)
	}
	return invoiceId, nil
}

func insertInvoice(db *sqlx.DB, invoice Invoice) error {
	var sqlInvoice = `
		INSERT INTO invoice (
			memo, r_preimage, r_hash, value_msat, creation_date, settle_date, payment_request,
			destination_pub_key, description_hash, expiry, fallback_addr, cltv_expiry, route_hints, private,
			add_index, settle_index, amt_paid_msat,
			/*
			The state the invoice is in.
				OPEN = 0;
				SETTLED = 1;
				CANCELED = 2;
				ACCEPTED = 3;
			*/
			invoice_state, htlcs, features, is_keysend, payment_addr, is_amp, amp_invoice_state,
			destination_node_id, node_id, channel_id, created_on, updated_on
		) VALUES(
			:memo, :r_preimage, :r_hash, :value_msat, :creation_date, :settle_date, :payment_request,
		    :destination_pub_key, :description_hash, :expiry, :fallback_addr, :cltv_expiry, :route_hints, :private,
			:add_index, :settle_index, :amt_paid_msat,
			:invoice_state, :htlcs, :features, :is_keysend, :payment_addr, :is_amp, :amp_invoice_state,
			:destination_node_id, :node_id, :channel_id, :created_on, :updated_on
		);`

	_, err := db.NamedExec(sqlInvoice, invoice)

	if err != nil {
		return errors.Wrapf(err, "insert invoice")
	}

	return nil
}

func updateInvoice(db *sqlx.DB, invoice Invoice) error {
	var sqlInvoice = `
		UPDATE invoice
		SET
			memo=:memo, r_preimage=:r_preimage, r_hash=:r_hash, value_msat=:value_msat, creation_date=:creation_date,
			settle_date=:settle_date, payment_request=:payment_request, destination_pub_key=:destination_pub_key,
			description_hash=:description_hash, expiry=:expiry, fallback_addr=:fallback_addr, cltv_expiry=:cltv_expiry,
			route_hints=:route_hints, private=:private,
			add_index=:add_index, settle_index=:settle_index, amt_paid_msat=:amt_paid_msat,
			/*
			The state the invoice is in.
				OPEN = 0;
				SETTLED = 1;
				CANCELED = 2;
				ACCEPTED = 3;
			*/
			invoice_state=:invoice_state, htlcs=:htlcs, features=:features, is_keysend=:is_keysend,
			payment_addr=:payment_addr, is_amp=:is_amp, amp_invoice_state=:amp_invoice_state,
			destination_node_id=:destination_node_id, node_id=:node_id, channel_id=:channel_id, updated_on=:updated_on
		WHERE invoice_id=:invoice_id;`

	_, err := db.NamedExec(sqlInvoice, invoice)

	if err != nil {
		return errors.Wrapf(err, "update invoice")
	}

	return nil
}
