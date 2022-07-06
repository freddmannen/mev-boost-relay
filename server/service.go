// Package server contains the webserver serving the proposer and block-builder APIs
package server

import (
	"context"
	"net/http"

	"github.com/flashbots/boost-relay/api/builderapi"
	"github.com/flashbots/boost-relay/api/proposerapi"
	"github.com/flashbots/boost-relay/beaconclient"
	"github.com/flashbots/boost-relay/common"
	"github.com/flashbots/boost-relay/datastore"
	"github.com/flashbots/go-utils/httplogger"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var (
	// Status API
	pathStatus = "/eth/v1/builder/status"
)

// RelayServiceOpts contains the options for a relay
type RelayServiceOpts struct {
	Ctx context.Context
	Log *logrus.Entry

	ListenAddr   string
	BeaconClient beaconclient.BeaconNodeClient
	Datastore    datastore.ProposerDatastore

	// // Whitelisted Builders
	// builders []*common.BuilderEntry

	// GenesisForkVersion for validating signatures
	GenesisForkVersionHex string

	// Which APIs and services to spin up
	ProposerAPI bool
	BuilderAPI  bool
}

// RelayService represents a single Relay instance
type RelayService struct {
	common.BaseAPI

	opts RelayServiceOpts
	srv  *http.Server
	apis []common.APIComponent
}

// NewRelayService creates a new service. if builders is nil, allow any builder
func NewRelayService(opts RelayServiceOpts) (*RelayService, error) {
	rs := RelayService{
		opts: opts,
		apis: make([]common.APIComponent, 0),
	}

	rs.Log = opts.Log.WithField("module", "relay")

	if opts.ProposerAPI {
		api, err := proposerapi.NewProposerAPI(opts.Ctx, opts.Log, opts.Datastore, opts.GenesisForkVersionHex)
		if err != nil {
			return nil, err
		}
		rs.apis = append(rs.apis, api)
	}

	if opts.BuilderAPI {
		api, err := builderapi.NewBuilderAPI(opts.Ctx, opts.Log, opts.Datastore, opts.GenesisForkVersionHex)
		if err != nil {
			return nil, err
		}
		rs.apis = append(rs.apis, api)
	}
	return &rs, nil
}

func (m *RelayService) getRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", m.handleRoot).Methods(http.MethodGet)
	r.HandleFunc(pathStatus, m.handleStatus).Methods(http.MethodGet)

	for _, api := range m.apis {
		api.RegisterHandlers(r)
	}

	// r.Use(mux.CORSMethodMiddleware(r))
	loggedRouter := httplogger.LoggingMiddlewareLogrus(m.Log, r)
	return loggedRouter
}

// StartComponents starts all the components
func (m *RelayService) StartComponents() (err error) {
	for _, api := range m.apis {
		err := api.Start()
		if err != nil {
			return err
		}
	}
	return nil
}

// StartServer starts the HTTP server for this instance
func (m *RelayService) StartServer() (err error) {
	if m.srv != nil {
		return common.ErrServerAlreadyRunning
	}

	err = m.StartComponents()
	if err != nil {
		return err
	}

	// // start everyting up
	// syncStatus, err := m.validatorService.SyncStatus()
	// if err != nil {
	// 	return err
	// }
	// if syncStatus.IsSyncing {
	// 	m.log.Error("Beacon node is syncing!")
	// 	return errors.New("beacon node is syncing")
	// }
	// m.slotCurrent = syncStatus.HeadSlot
	// m.log.WithField("slot", m.slotCurrent).Info("current slot")

	// go m.startBeaconNodeSlotUpdates()
	// // go m.startBeaconNodeValidatorUpdates()

	m.srv = &http.Server{
		Addr:    m.opts.ListenAddr,
		Handler: m.getRouter(),
	}

	err = m.srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// func (m *RelayService) startBeaconNodeSlotUpdates() {
// 	c := make(chan uint64)
// 	go m.validatorService.SubscribeToHeadEvents(c)
// 	for {
// 		m.slotCurrent = <-c
// 		m.log.WithField("slot", m.slotCurrent).Info("new slot")
// 	}
// }

// func (m *RelayService) startBeaconNodeValidatorUpdates() {
// 	for {
// 		// Wait for one epoch (at the beginning, because initially the validators have already been queried)
// 		time.Sleep(common.DurationPerEpoch)

// 		// Load validators from BN
// 		m.log.Info("Querying validators from beacon node... (this may")
// 		err := m.validatorService.FetchValidators()
// 		if err != nil {
// 			m.log.WithError(err).Fatal("failed to fetch validators from beacon node")
// 		}
// 		m.log.Infof("Got %d validators from BN", m.validatorService.NumValidators())
// 	}
// }

func (m *RelayService) handleRoot(w http.ResponseWriter, req *http.Request) {
	m.RespondOKEmpty(w)
}

func (m *RelayService) handleStatus(w http.ResponseWriter, req *http.Request) {
	m.RespondOKEmpty(w)
}