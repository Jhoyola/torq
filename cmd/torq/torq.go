package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/andres-erbsen/clock"
	"github.com/cockroachdb/errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"google.golang.org/grpc"

	"github.com/lncapital/torq/build"
	"github.com/lncapital/torq/cmd/torq/internal/amboss_ping"
	"github.com/lncapital/torq/cmd/torq/internal/services"
	"github.com/lncapital/torq/cmd/torq/internal/subscribe"
	"github.com/lncapital/torq/cmd/torq/internal/torqsrv"
	"github.com/lncapital/torq/cmd/torq/internal/vector_ping"
	"github.com/lncapital/torq/internal/corridors"
	"github.com/lncapital/torq/internal/database"
	"github.com/lncapital/torq/internal/settings"
	"github.com/lncapital/torq/internal/tags"
	"github.com/lncapital/torq/internal/workflows"
	"github.com/lncapital/torq/pkg/commons"
	"github.com/lncapital/torq/pkg/lnd"
	"github.com/lncapital/torq/pkg/lnd_connect"
)

const servicesErrorSleepSeconds = 60

var debuglevels = map[string]zerolog.Level{ //nolint:gochecknoglobals
	"panic": zerolog.PanicLevel,
	"fatal": zerolog.FatalLevel,
	"error": zerolog.ErrorLevel,
	"warn":  zerolog.WarnLevel,
	"info":  zerolog.InfoLevel,
	"debug": zerolog.DebugLevel,
	"trace": zerolog.TraceLevel,
}

func main() {

	app := cli.NewApp()
	app.Name = "torq"
	app.EnableBashCompletion = true
	app.Version = build.ExtendedVersion()

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal().Msgf("error finding home directory of user: %v", err)
	}

	cmdFlags := []cli.Flag{

		// All these flags can be set though a common config file.
		&cli.StringFlag{
			Name:    "config",
			Value:   homedir + "/.torq/torq.conf",
			Aliases: []string{"c"},
			Usage:   "Path to config file",
		},

		// Torq details
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.pprof.path",
			Value: "localhost:6060",
			Usage: "Set pprof path",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.vector.url",
			Value: commons.VectorUrl,
			Usage: "Enable test mode",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.debuglevel",
			Value: "info",
			Usage: "Specify different debuglevels (panic|fatal|error|warn|info|debug|trace)",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.cookie-path",
			Usage: "Path to auth cookie file",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.password",
			Usage: "Password used to access the API and frontend",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "torq.port",
			Value: "8080",
			Usage: "Port to serve the HTTP API",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:  "torq.no-sub",
			Value: false,
			Usage: "Start the server without subscribing to node data",
		}),
		altsrc.NewBoolFlag(&cli.BoolFlag{
			Name:  "torq.auto-login",
			Value: false,
			Usage: "Allows logging in without a password",
		}),

		// Torq database
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.name",
			Value: "torq",
			Usage: "Name of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.port",
			Value: "5432",
			Usage: "Port of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.host",
			Value: "localhost",
			Usage: "Host of the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.user",
			Value: "torq",
			Usage: "Name of the postgres user with access to the database",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "db.password",
			Value: "password",
			Usage: "Password used to access the database",
		}),

		// LND connection details
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.url",
			Usage: "Host:Port of the LND node",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.macaroon-path",
			Usage: "Path on disk to LND Macaroon",
		}),
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:  "lnd.tls-path",
			Usage: "Path on disk to LND TLS file",
		}),
	}

	start := &cli.Command{
		Name:  "start",
		Usage: "Start the main daemon",
		Action: func(c *cli.Context) error {

			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if debuglevel, ok := debuglevels[strings.ToLower(c.String("torq.debuglevel"))]; ok {
				zerolog.SetGlobalLevel(debuglevel)
				log.Debug().Msgf("DebugLevel: %v enabled", debuglevel)
			}

			// Print startup message
			fmt.Printf("Starting Torq %s\n", build.ExtendedVersion())

			fmt.Println("Connecting to the Torq database")
			db, err := database.PgConnect(c.String("db.name"), c.String("db.user"),
				c.String("db.password"), c.String("db.host"), c.String("db.port"))
			if err != nil {
				return errors.Wrap(err, "start cmd")
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			// initialise package level var for keeping state of subsciptions
			commons.RunningServices = make(map[commons.ServiceType]*commons.Services, 0)
			for _, serviceType := range commons.GetServiceTypes() {
				commons.RunningServices[serviceType] = &commons.Services{ServiceType: serviceType}
			}

			ctxGlobal, cancelGlobal := context.WithCancel(context.Background())
			defer cancelGlobal()

			go commons.ManagedChannelGroupCache(commons.ManagedChannelGroupChannel, ctxGlobal)
			go commons.ManagedChannelStateCache(commons.ManagedChannelStateChannel, ctxGlobal)
			go commons.ManagedSettingsCache(commons.ManagedSettingsChannel, ctxGlobal)
			go commons.ManagedNodeCache(commons.ManagedNodeChannel, ctxGlobal)
			go commons.ManagedNodeAliasCache(commons.ManagedNodeAliasChannel, ctxGlobal)
			go commons.ManagedChannelCache(commons.ManagedChannelChannel, ctxGlobal)
			go commons.ManagedTaggedCache(commons.ManagedTaggedChannel, ctxGlobal)
			go commons.ManagedTriggerCache(commons.ManagedTriggerChannel, ctxGlobal)
			go tags.ManagedTagCache(tags.ManagedTagChannel, ctxGlobal)
			go workflows.ManagedRebalanceCache(workflows.ManagedRebalanceChannel, ctxGlobal)
			go services.ManagedServiceCache(services.ManagedServiceChannel, ctxGlobal)

			commons.SetVectorUrlBase(c.String("torq.vector.url"))

			commons.RunningServices[commons.TorqService].AddSubscription(commons.TorqDummyNodeId, cancelGlobal)

			// This function initiates the database migration(s) and parses command line parameters
			// When done the TorqService is set to Initialising
			go migrateAndProcessArguments(db, c)

			// go routine that responds to commands to boot and kill services
			if !c.Bool("torq.no-sub") {
				go processDelayedServiceCommands(db)
				go serviceChannelRoutine(db)
			} else {
				go serviceChannelDummyRoutine()
			}

			if c.String("torq.pprof.path") != "" {
				go pprofStartup(c)
			}

			go processServiceStatusChanges(db)

			if err = torqsrv.Start(c.Int("torq.port"), c.String("torq.password"),
				c.String("torq.cookie-path"),
				db, c.Bool("torq.auto-login")); err != nil {
				return errors.Wrap(err, "Starting torq webserver")
			}

			return nil
		},
	}

	migrateUp := &cli.Command{
		Name:  "migrate_up",
		Usage: "Migrates the database to the latest version",
		Action: func(c *cli.Context) error {
			db, err := database.PgConnect(c.String("db.name"), c.String("db.user"),
				c.String("db.password"), c.String("db.host"), c.String("db.port"))
			if err != nil {
				return errors.Wrap(err, "Database connect")
			}

			defer func() {
				cerr := db.Close()
				if err == nil {
					err = cerr
				}
			}()

			err = database.MigrateUp(db)
			if err != nil {
				return errors.Wrap(err, "Migrating database up")
			}

			return nil
		},
	}

	app.Flags = cmdFlags

	app.Before = altsrc.InitInputSourceWithContext(cmdFlags, loadFlags())

	app.Commands = cli.Commands{
		start,
		migrateUp,
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

func pprofStartup(c *cli.Context) {
	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	runtime.SetCPUProfileRate(1)
	err := http.ListenAndServe(c.String("torq.pprof.path"), nil) //nolint:gosec
	if err != nil {
		log.Error().Err(err).Msg("Torq could not start pprof")
	}
}

func loadFlags() func(context *cli.Context) (altsrc.InputSourceContext, error) {
	return func(context *cli.Context) (altsrc.InputSourceContext, error) {
		if _, err := os.Stat(context.String("config")); err != nil {
			return altsrc.NewMapInputSource("", map[interface{}]interface{}{}), nil
		}
		tomlSource, err := altsrc.NewTomlSourceFromFile(context.String("config"))
		if err != nil {
			return nil, errors.Wrap(err, "Creating new toml config from file")
		}
		return tomlSource, nil
	}
}

func migrateAndProcessArguments(db *sqlx.DB, c *cli.Context) {
	fmt.Println("Checking for migrations..")
	// Check if the database needs to be migrated.
	err := database.MigrateUp(db)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Error().Err(err).Msg("Torq could not migrate the database.")
		commons.RunningServices[commons.TorqService].RemoveSubscription(commons.TorqDummyNodeId)
		return
	}

	for {
		// if node specified on cmd flags then check if we already know about it
		if c.String("lnd.url") != "" && c.String("lnd.macaroon-path") != "" && c.String("lnd.tls-path") != "" {
			macaroonFile, err := os.ReadFile(c.String("lnd.macaroon-path"))
			if err != nil {
				log.Error().Err(err).Msg("Reading macaroon file from disk path from config")
				log.Error().Err(err).Msg("LND is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			tlsFile, err := os.ReadFile(c.String("lnd.tls-path"))
			if err != nil {
				log.Error().Err(err).Msg("Reading tls file from disk path from config")
				log.Error().Err(err).Msg("LND is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			grpcAddress := c.String("lnd.url")
			nodeId, err := settings.GetNodeIdByGRPC(db, grpcAddress)
			if err != nil {
				log.Error().Err(err).Msg("Checking if node specified in config exists")
				log.Error().Err(err).Msg("LND is probably not ready (will retry in 10 seconds)")
				time.Sleep(10 * time.Second)
				continue
			}
			if nodeId == 0 {
				log.Info().Msgf("Node specified in config is not in DB, obtaining public key from GRPC: %v", grpcAddress)
				var nodeConnectionDetails settings.NodeConnectionDetails
				for {
					nodeConnectionDetails, err = settings.AddNodeToDB(db, commons.LND, grpcAddress, tlsFile, macaroonFile)
					if err == nil && nodeConnectionDetails.NodeId != 0 {
						break
					} else {
						log.Error().Err(err).Msg("Adding node specified in config to database, LND is probably booting (will retry in 10 seconds)")
						time.Sleep(10 * time.Second)
					}
				}
				nodeConnectionDetails.Name = "Auto configured node"
				nodeConnectionDetails.CustomSettings = commons.NodeConnectionDetailCustomSettings(commons.NodeConnectionDetailCustomSettingsMax - int(commons.ImportFailedPayments))
				_, err = settings.SetNodeConnectionDetails(db, nodeConnectionDetails)
				if err != nil {
					log.Error().Err(err).Msg("Failed to update the node name (cosmetics problem).")
				}
			} else {
				log.Info().Msg("Node specified in config is present, updating Macaroon and TLS files")
				if err = settings.SetNodeConnectionDetailsByConnectionDetails(db, nodeId, commons.Active, grpcAddress, tlsFile, macaroonFile); err != nil {
					log.Error().Err(err).Msg("Problem updating node files")
					commons.RunningServices[commons.TorqService].RemoveSubscription(commons.TorqDummyNodeId)
				}
			}
		}
		break
	}

	commons.RunningServices[commons.TorqService].Initialising(commons.TorqDummyNodeId)
}

func serviceChannelDummyRoutine() {
	for {
		serviceCmd := <-commons.ServiceChannel
		log.Warn().Msgf("Ignoring Service call for node id: %v", serviceCmd.NodeId)
	}
}

func serviceChannelRoutine(db *sqlx.DB) {
	for {
		serviceCmd := <-commons.ServiceChannel
		runningServices := commons.RunningServices[serviceCmd.ServiceType]
		var nodes []settings.ConnectionDetails
		var serviceNode settings.ConnectionDetails
		var err error
		var enforcedServiceStatus *commons.ServiceStatus
		name := serviceCmd.ServiceType.String()
		if name != commons.UnknownEnumString {
			if serviceCmd.ServiceCommand == commons.Boot {
				if serviceCmd.NodeId != 0 {
					enforcedServiceStatus = runningServices.GetEnforcedServiceStatusCheck(serviceCmd.NodeId)
				}
				if serviceCmd.EnforcedServiceStatus != nil {
					enforcedServiceStatus = serviceCmd.EnforcedServiceStatus
				}
				if serviceCmd.NodeId != 0 {
					if serviceCmd.NodeId == commons.TorqDummyNodeId {
						nodes = []settings.ConnectionDetails{{
							NodeId: commons.TorqDummyNodeId,
							Status: commons.Active,
						}}
					} else {
						if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.ServiceInactive {
							nodes = []settings.ConnectionDetails{}
						} else {
							serviceNode, err = settings.GetConnectionDetailsById(db, serviceCmd.NodeId)
							if err != nil {
								log.Error().Err(errors.Wrapf(err, "%v Service Boot: Getting connection details", name)).Send()
								return
							}
							if enforcedServiceStatus != nil && *enforcedServiceStatus == commons.ServiceActive {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								if serviceNode.Status != commons.Active {
									nodes = []settings.ConnectionDetails{}
								}
							}
						}
					}
				}
				log.Info().Msgf("%v Service: Verifying requirement.", name)
				if nodes == nil {
					if serviceCmd.NodeId == 0 {
						if serviceCmd.ServiceType == commons.LndService {
							commons.RunningServices[commons.TorqService].Booted(commons.TorqDummyNodeId, nil)
						}
						switch serviceCmd.ServiceType {
						case commons.VectorService:
							nodes, err = settings.GetVectorPingNodesConnectionDetails(db)
						case commons.AmbossService:
							nodes, err = settings.GetAmbossPingNodesConnectionDetails(db)
						default:
							nodes, err = settings.GetActiveNodesConnectionDetails(db)
						}
						if err != nil {
							log.Error().Err(err).Msgf("%v Service: Getting connection details", name)
						}
					} else {
						switch serviceCmd.ServiceType {
						case commons.VectorService:
							if serviceNode.Status == commons.Active && serviceNode.HasPingSystem(settings.Vector) {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								nodes = []settings.ConnectionDetails{}
							}
						case commons.AmbossService:
							if serviceNode.Status == commons.Active && serviceNode.HasPingSystem(settings.Amboss) {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								nodes = []settings.ConnectionDetails{}
							}
						default:
							if serviceNode.Status == commons.Active {
								nodes = []settings.ConnectionDetails{serviceNode}
							} else {
								nodes = []settings.ConnectionDetails{}
							}
						}
					}
				}

				if len(nodes) == 0 {
					log.Info().Msgf("%v Service: No active nodes found (nodeId: %v).", name, serviceCmd.NodeId)
					if serviceCmd.NodeId != 0 {
						runningServices.RemoveSubscription(serviceCmd.NodeId)
					}
				} else {
					if serviceCmd.DelaySeconds != nil {
						log.Error().Msgf("%v Service: Sleeping for %v seconds before attempting to boot.", name, *serviceCmd.DelaySeconds)
						services.SetDelayedServiceCommand(services.DelayedServiceCommand{
							Name:                  name,
							ServiceChannelMessage: serviceCmd,
							Nodes:                 nodes,
							StartTime:             time.Now().Add(time.Duration(*serviceCmd.DelaySeconds) * time.Second),
						})
					} else {
						for _, node := range nodes {
							bootLock := runningServices.GetBootLock(node.NodeId)
							successful := bootLock.TryLock()
							if successful {
								go processServiceBoot(name, db, node, bootLock, runningServices, serviceCmd)
							} else {
								log.Error().Msgf("%v Service: Requested start failed. A start is already running.", name)
							}
						}
					}
				}
			}
		}
	}
}

func processDelayedServiceCommands(db *sqlx.DB) {
	ticker := clock.New().Tick(1 * time.Second)

	for {
		<-ticker
		for {
			delayedServiceCommand := services.PopDelayedServiceCommand()
			if delayedServiceCommand.Name == "" {
				break
			}
			runningServices := commons.RunningServices[delayedServiceCommand.ServiceChannelMessage.ServiceType]
			for _, node := range delayedServiceCommand.Nodes {
				bootLock := runningServices.GetBootLock(node.NodeId)
				successful := bootLock.TryLock()
				if successful {
					go processServiceBoot(delayedServiceCommand.Name, db, node, bootLock, runningServices,
						delayedServiceCommand.ServiceChannelMessage)
				} else {
					log.Error().Msgf("%v Service: Requested start failed. A start is already running.", delayedServiceCommand.Name)
				}
			}
		}
	}
}

func processServiceStatusChanges(db *sqlx.DB) {
	ticker := clock.New().Tick(1 * time.Second)
	torqDummyServiceStatus := make(map[commons.ServiceType]commons.ServiceStatus)
	torqDummyServiceStatus[commons.TorqService] = commons.ServiceInactive
	torqDummyServiceStatus[commons.MaintenanceService] = commons.ServiceInactive
	torqDummyServiceStatus[commons.CronService] = commons.ServiceInactive
	torqDummyServiceStatus[commons.AutomationService] = commons.ServiceInactive

	nodeServiceStatus := make(map[commons.ServiceType]map[int]commons.ServiceStatus)
	nodeServiceStatus[commons.RebalanceService] = make(map[int]commons.ServiceStatus)
	nodeServiceStatus[commons.VectorService] = make(map[int]commons.ServiceStatus)
	nodeServiceStatus[commons.AmbossService] = make(map[int]commons.ServiceStatus)

	subscriptionStreamServiceStatus := make(map[commons.SubscriptionStream]map[int]commons.ServiceStatus)
	for _, subscriptionStream := range commons.SubscriptionStreams {
		subscriptionStreamServiceStatus[subscriptionStream] = make(map[int]commons.ServiceStatus)
	}

	for {
		<-ticker
		processedNodeId := make(map[int]bool)
		torqNodeIds := commons.GetAllTorqNodeIds()
		processDummyService(db, torqDummyServiceStatus, commons.TorqService)
		processDummyService(db, torqDummyServiceStatus, commons.MaintenanceService)
		processDummyService(db, torqDummyServiceStatus, commons.CronService)
		processDummyService(db, torqDummyServiceStatus, commons.AutomationService)
		for _, torqNodeId := range torqNodeIds {
			processService(db, nodeServiceStatus, commons.RebalanceService, torqNodeId)
			processService(db, nodeServiceStatus, commons.VectorService, torqNodeId)
			processService(db, nodeServiceStatus, commons.AmbossService, torqNodeId)
			for ix, subscriptionStream := range commons.SubscriptionStreams {
				currentNodeStreamStatus := commons.RunningServices[commons.LndService].GetStreamStatus(torqNodeId, subscriptionStream)
				if currentNodeStreamStatus != subscriptionStreamServiceStatus[subscriptionStream][torqNodeId] {
					sendServiceEvent(db, torqNodeId,
						subscriptionStreamServiceStatus[subscriptionStream][torqNodeId], currentNodeStreamStatus, commons.LndService, &commons.SubscriptionStreams[ix])
					subscriptionStreamServiceStatus[subscriptionStream][torqNodeId] = currentNodeStreamStatus
				}
			}
			processedNodeId[torqNodeId] = true
		}
		for nodeId := range nodeServiceStatus[commons.RebalanceService] {
			if processedNodeId[nodeId] {
				continue
			}
			processService(db, nodeServiceStatus, commons.RebalanceService, nodeId)
		}
		for nodeId := range nodeServiceStatus[commons.VectorService] {
			if processedNodeId[nodeId] {
				continue
			}
			processService(db, nodeServiceStatus, commons.VectorService, nodeId)
		}
		for nodeId := range nodeServiceStatus[commons.AmbossService] {
			if processedNodeId[nodeId] {
				continue
			}
			processService(db, nodeServiceStatus, commons.AmbossService, nodeId)
		}
		for ix, subscriptionStream := range commons.SubscriptionStreams {
			for nodeId, subscriptionStreamStatus := range subscriptionStreamServiceStatus[subscriptionStream] {
				currentNodeStreamStatus := commons.RunningServices[commons.LndService].GetStreamStatus(nodeId, subscriptionStream)
				if currentNodeStreamStatus != subscriptionStreamStatus {
					sendServiceEvent(db, nodeId,
						subscriptionStreamStatus, currentNodeStreamStatus, commons.LndService, &commons.SubscriptionStreams[ix])
					subscriptionStreamServiceStatus[subscriptionStream][nodeId] = currentNodeStreamStatus
				}
			}
		}
	}
}

func processService(db *sqlx.DB, nodeServiceStatus map[commons.ServiceType]map[int]commons.ServiceStatus,
	serviceType commons.ServiceType,
	nodeId int) {

	currentNodeServiceStatus := commons.RunningServices[serviceType].GetStatus(nodeId)
	if currentNodeServiceStatus != nodeServiceStatus[serviceType][nodeId] {
		sendServiceEvent(db, nodeId,
			nodeServiceStatus[serviceType][nodeId], currentNodeServiceStatus, serviceType, nil)
		nodeServiceStatus[serviceType][nodeId] = currentNodeServiceStatus
	}
}

func processDummyService(db *sqlx.DB, torqDummyServiceStatus map[commons.ServiceType]commons.ServiceStatus,
	serviceType commons.ServiceType) {

	currentTorqServiceStatus := commons.RunningServices[serviceType].GetStatus(commons.TorqDummyNodeId)
	if currentTorqServiceStatus != torqDummyServiceStatus[serviceType] {
		sendServiceEvent(db, commons.TorqDummyNodeId,
			torqDummyServiceStatus[serviceType], currentTorqServiceStatus, serviceType, nil)
		torqDummyServiceStatus[serviceType] = currentTorqServiceStatus
	}
}

func sendServiceEvent(db *sqlx.DB,
	nodeId int,
	previousStatus commons.ServiceStatus,
	status commons.ServiceStatus,
	serviceType commons.ServiceType,
	subscriptionStream *commons.SubscriptionStream) {

	if previousStatus != status {
		log.Debug().Msgf("Updated service %v (%v) from %v to %v for nodeId: %v",
			serviceType.String(), subscriptionStream.String(), previousStatus.String(), status.String(), nodeId)
		serviceEvent := commons.ServiceEvent{
			EventData: commons.EventData{
				EventTime: time.Now().UTC(),
				NodeId:    nodeId,
			},
			Type:               serviceType,
			SubscriptionStream: subscriptionStream,
			Status:             status,
			PreviousStatus:     previousStatus,
		}
		processServiceEvents(db, serviceEvent)
		lnd.ProcessServiceEvent(serviceEvent)
	}
}

func processServiceEvents(db *sqlx.DB, serviceEvent commons.ServiceEvent) {
	log.Debug().Msgf("Service event received for type: %v (%v), previousStatus: %v, status: %v",
		serviceEvent.Type.String(), serviceEvent.SubscriptionStream.String(),
		serviceEvent.PreviousStatus.String(), serviceEvent.Status.String())
	if serviceEvent.Type == commons.TorqService {
		switch serviceEvent.Status {
		case commons.ServiceInactive:
			log.Info().Msg("Torq is dead.")
			panic("TorqService cannot be bootstrapped")
		case commons.ServicePending:
			log.Info().Msg("Torq is booting.")
		case commons.ServiceInitializing:
			log.Info().Msg("Torq is initialising.")

			err := settings.InitializeManagedSettingsCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain settings for ManagedSettings cache.")
			}

			err = settings.InitializeManagedNodeCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain torq nodes for ManagedNode cache.")
			}

			err = settings.InitializeManagedChannelCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain channels for ManagedChannel cache.")
			}

			settings.InitializeManagedNodeAliasCache(db)

			err = settings.InitializeManagedTaggedCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain tags for ManagedTagged cache.")
			}

			err = tags.InitializeManagedTagCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Failed to obtain tags for ManagedTag cache.")
			}

			log.Info().Msg("Loading caches in memory.")
			err = corridors.RefreshCorridorCache(db)
			if err != nil {
				log.Error().Err(err).Msg("Torq cannot be initialized (Loading caches in memory).")
			}
			log.Debug().Msgf("LightningCommunicationService boot for nodeId: %v", serviceEvent.NodeId)
			commons.ServiceChannel <- commons.ServiceChannelMessage{
				ServiceCommand: commons.Boot,
				ServiceType:    commons.LndService,
			}
		case commons.ServiceActive:
			log.Debug().Msgf("MaintenanceService boot for nodeId: %v", serviceEvent.NodeId)
			commons.ServiceChannel <- commons.ServiceChannelMessage{
				ServiceCommand: commons.Boot,
				ServiceType:    commons.MaintenanceService,
				NodeId:         commons.TorqDummyNodeId,
			}
			log.Debug().Msgf("AutomationService boot for nodeId: %v", serviceEvent.NodeId)
			commons.ServiceChannel <- commons.ServiceChannelMessage{
				ServiceCommand: commons.Boot,
				ServiceType:    commons.AutomationService,
				NodeId:         commons.TorqDummyNodeId,
			}
			log.Debug().Msgf("CronService boot for nodeId: %v", serviceEvent.NodeId)
			commons.ServiceChannel <- commons.ServiceChannelMessage{
				ServiceCommand: commons.Boot,
				ServiceType:    commons.CronService,
				NodeId:         commons.TorqDummyNodeId,
			}
		}
	}
	if serviceEvent.Type == commons.LndService {
		if serviceEvent.Status == commons.ServiceActive {
			log.Debug().Msgf("LndService booted for nodeId: %v", serviceEvent.NodeId)
			if commons.RunningServices[commons.RebalanceService].GetStatus(serviceEvent.NodeId) != commons.ServiceActive {
				log.Debug().Msgf("Starting Rebalance Service for nodeId: %v", serviceEvent.NodeId)
				commons.ServiceChannel <- commons.ServiceChannelMessage{
					ServiceCommand: commons.Boot,
					ServiceType:    commons.RebalanceService,
					NodeId:         serviceEvent.NodeId,
				}
			}
			if commons.RunningServices[commons.VectorService].GetStatus(serviceEvent.NodeId) != commons.ServiceActive {
				log.Debug().Msgf("Checking for Vector activation for nodeId: %v", serviceEvent.NodeId)
				commons.ServiceChannel <- commons.ServiceChannelMessage{
					ServiceCommand: commons.Boot,
					ServiceType:    commons.VectorService,
					NodeId:         serviceEvent.NodeId,
				}
			}
			if commons.RunningServices[commons.AmbossService].GetStatus(serviceEvent.NodeId) != commons.ServiceActive {
				log.Debug().Msgf("Checking for Amboss activation for nodeId: %v", serviceEvent.NodeId)
				commons.ServiceChannel <- commons.ServiceChannelMessage{
					ServiceCommand: commons.Boot,
					ServiceType:    commons.AmbossService,
					NodeId:         serviceEvent.NodeId,
				}
			}
		}
	}
	if serviceEvent.SubscriptionStream == nil {
		switch serviceEvent.Status {
		case commons.ServiceBootRequestedWithDelay:
			log.Info().Msgf("Service will be restarted (when active) in %v seconds for node id: %v",
				servicesErrorSleepSeconds, serviceEvent.NodeId)

			delaySeconds := servicesErrorSleepSeconds
			commons.ServiceChannel <- commons.ServiceChannelMessage{
				ServiceCommand: commons.Boot,
				ServiceType:    serviceEvent.Type,
				NodeId:         serviceEvent.NodeId,
				DelaySeconds:   &delaySeconds,
			}
		case commons.ServiceBootRequested:
			log.Info().Msgf("Service will be restarted (when active) for node id: %v", serviceEvent.NodeId)
			commons.ServiceChannel <- commons.ServiceChannelMessage{
				ServiceCommand: commons.Boot,
				ServiceType:    serviceEvent.Type,
				NodeId:         serviceEvent.NodeId,
			}
		}
	}
}

func processServiceBoot(name string, db *sqlx.DB, node settings.ConnectionDetails,
	bootLock *sync.Mutex,
	runningServices *commons.Services,
	serviceCmd commons.ServiceChannelMessage) {

	defer func() {
		if commons.MutexLocked(bootLock) {
			bootLock.Unlock()
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	log.Info().Msgf("Generating %v Service for node id: %v", name, node.NodeId)
	runningServices.AddSubscription(node.NodeId, cancel)

	var conn *grpc.ClientConn
	var err error
	switch serviceCmd.ServiceType {
	case commons.VectorService, commons.AmbossService, commons.RebalanceService, commons.LndService:
		conn, err = lnd_connect.Connect(
			node.GRPCAddress,
			node.TLSFileBytes,
			node.MacaroonFileBytes)
		if err != nil {
			log.Error().Err(err).Msgf("%v Service Failed to connect to lnd for node id: %v", name, node.NodeId)
			runningServices.RemoveSubscription(node.NodeId)
			if serviceCmd.ServiceType == commons.LndService {
				log.Info().Msgf("LND Service will be restarted (when active) for node id: %v", node.NodeId)
				runningServices.SetBootRequestedStatus(node.NodeId, true)
				serviceInactive := commons.ServiceInactive
				commons.RunningServices[commons.VectorService].Cancel(node.NodeId, &serviceInactive, false)
				commons.RunningServices[commons.AmbossService].Cancel(node.NodeId, &serviceInactive, false)
				commons.RunningServices[commons.RebalanceService].Cancel(node.NodeId, &serviceInactive, false)
				commons.RunningServices[commons.LndService].Cancel(node.NodeId, &serviceInactive, false)
			}
			return
		}
		commons.SetNodeConnectionDetailsByNodeId(node.NodeId, node.GRPCAddress, node.TLSFileBytes, node.MacaroonFileBytes)
	}

	runningServices.Booted(node.NodeId, bootLock)
	if serviceCmd.ServiceType == commons.LndService {
		commons.RunningServices[commons.LndService].SetNodeConnectionDetailCustomSettings(node.NodeId, node.CustomSettings)
	}
	log.Info().Msgf("%v Service booted for node id: %v", name, node.NodeId)
	switch serviceCmd.ServiceType {
	// NODE SPECIFIC
	case commons.VectorService:
		err = vector_ping.Start(ctx, conn, node.NodeId)
	case commons.AmbossService:
		err = amboss_ping.Start(ctx, conn, node.NodeId)
	case commons.RebalanceService:
		services.StartRebalanceService(ctx, conn, db, node.NodeId)
	case commons.LndService:
		err = subscribe.Start(ctx, conn, db, node.NodeId)
	// NOT NODE ID SPECIFIC
	case commons.AutomationService:
		services.Start(ctx, db)
	case commons.MaintenanceService:
		services.StartMaintenanceService(ctx, db)
	case commons.CronService:
		services.StartCronService(ctx, db)
	}
	if err != nil {
		log.Error().Err(err).Msgf("%v Service ended for node id: %v", name, node.NodeId)
	}
	log.Info().Msgf("%v Service stopped for node id: %v", name, node.NodeId)
	runningServices.RemoveSubscription(node.NodeId)
	if runningServices.IsNoDelay(node.NodeId) || serviceCmd.NoDelay {
		log.Info().Msgf("%v Service will be restarted (when active) for node id: %v", name, node.NodeId)
	} else {
		log.Info().Msgf("%v Service will be restarted (when active) in %v seconds for node id: %v", name, servicesErrorSleepSeconds, node.NodeId)
	}
	runningServices.SetBootRequestedStatus(node.NodeId, runningServices.IsNoDelay(node.NodeId) || serviceCmd.NoDelay)
}
