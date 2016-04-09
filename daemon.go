package mnemosyne

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/piotrkowalczuk/sklog"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// DaemonOpts ...
type DaemonOpts struct {
	Namespace              string
	Subsystem              string
	MonitoringEngine       string
	TLS                    bool
	TLSCertFile            string
	TLSKeyFile             string
	StorageEngine          string
	StoragePostgresAddress string
	StoragePostgresTable   string
	Logger                 log.Logger
	RPCOptions             []grpc.ServerOption
	RPCListener            net.Listener
	DebugListener          net.Listener
}

// Daemon ...
type Daemon struct {
	opts          *DaemonOpts
	monitor       *monitoring
	rpcOptions    []grpc.ServerOption
	storage       Storage
	logger        log.Logger
	rpcListener   net.Listener
	debugListener net.Listener
}

// NewDaemon ...
func NewDaemon(opts *DaemonOpts) *Daemon {
	d := &Daemon{
		opts:          opts,
		logger:        opts.Logger,
		rpcOptions:    opts.RPCOptions,
		rpcListener:   opts.RPCListener,
		debugListener: opts.DebugListener,
	}

	if d.opts.StorageEngine == "" {
		d.opts.StorageEngine = StorageEnginePostgres
	}
	if d.opts.StoragePostgresTable == "" {
		d.opts.StoragePostgresTable = "session"
	}
	return d
}

func (d *Daemon) Run() (err error) {
	if err = d.initMonitoring(); err != nil {
		return
	}
	if err = d.initStorage(); err != nil {
		return
	}

	if d.opts.TLS {
		creds, err := credentials.NewServerTLSFromFile(d.opts.TLSCertFile, d.opts.TLSKeyFile)
		if err != nil {
			return err
		}
		d.rpcOptions = append(d.rpcOptions, grpc.Creds(creds))
	}

	gRPCServer := grpc.NewServer(d.rpcOptions...)
	mnemosyneServer := newRPCServer(d.logger, d.storage, d.monitor)
	RegisterRPCServer(gRPCServer, mnemosyneServer)

	go func() {
		sklog.Info(d.logger, "rpc server is running", "address", d.rpcListener.Addr().String(), "subsystem", d.opts.Subsystem, "namespace", d.opts.Namespace)

		if err := gRPCServer.Serve(d.rpcListener); err != nil {
			if err == grpc.ErrServerStopped {
				return
			}

			sklog.Error(d.logger, err)
		}
	}()

	if d.debugListener != nil {
		go func() {
			sklog.Info(d.logger, "debug server is running", "address", d.debugListener.Addr().String(), "subsystem", d.opts.Subsystem, "namespace", d.opts.Namespace)
			// TODO: implement keep alive
			sklog.Error(d.logger, http.Serve(d.debugListener, nil))
		}()
	}

	return
}

// Close implements io.Closer interface.
func (d *Daemon) Close() (err error) {
	if err = d.rpcListener.Close(); err != nil {
		return
	}
	if d.debugListener != nil {
		err = d.debugListener.Close()
	}
	return
}

// Addr returns net.Addr that rpc service is listening on.
func (d *Daemon) Addr() net.Addr {
	return d.rpcListener.Addr()
}

func (d *Daemon) initStorage() (err error) {
	var db *sql.DB

	switch d.opts.StorageEngine {
	case StorageEngineInMemory:
		return errors.New("mnemosyne: in memory storage is not implemented yet")
	case StorageEnginePostgres:
		db, err = initPostgres(
			d.opts.StoragePostgresAddress,
			d.logger,
		)
		if err != nil {
			return
		}
		if d.storage, err = initStorage(newPostgresStorage(d.opts.StoragePostgresTable, db, d.monitor), d.logger); err != nil {
			return
		}
		return
	case StorageEngineRedis:
		return errors.New("mnemosyne: redis storage is not implemented yet")
	default:
		return errors.New("mnemosyne: unknown storage engine")
	}
}

func (d *Daemon) initMonitoring() (err error) {
	hostname, err := os.Hostname()
	if err != nil {
		return errors.New("mnemosyne: getting hostname failed")
	}

	switch d.opts.MonitoringEngine {
	case "":
		d.monitor = &monitoring{}
		return
	case MonitoringEnginePrometheus:
		d.monitor = initPrometheus(d.opts.Namespace, d.opts.Subsystem, prometheus.Labels{"server": hostname})
		return
	default:
		return errors.New("mnemosyne: unknown monitoring engine")
	}
}