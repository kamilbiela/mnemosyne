package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/piotrkowalczuk/mnemosyne"
)

const VERSION = "0.2.1"

type configuration struct {
	host      string
	port      int
	namespace string
	subsystem string
	logger    struct {
		adapter string
		format  string
		level   int
	}
	monitoring struct {
		engine string
	}
	storage struct {
		engine   string
		postgres struct {
			address string
			table   string
		}
	}
	tls struct {
		enabled  bool
		certFile string
		keyFile  string
	}
}

func (c *configuration) init() {
	if c == nil {
		*c = configuration{}
	}

	flag.StringVar(&c.host, "host", "127.0.0.1", "host")
	flag.IntVar(&c.port, "port", 8080, "port")
	flag.StringVar(&c.namespace, "namespace", "", "namespace")
	flag.StringVar(&c.subsystem, "subsystem", "mnemosyne", "subsystem")
	flag.StringVar(&c.logger.adapter, "l.adapter", mnemosyne.LoggerAdapterStdOut, "logger adapter")
	flag.StringVar(&c.logger.format, "l.format", mnemosyne.LoggerFormatJSON, "logger format")
	flag.IntVar(&c.logger.level, "l.level", 6, "logger level")
	flag.StringVar(&c.monitoring.engine, "m.engine", mnemosyne.MonitoringEnginePrometheus, "monitoring engine")
	flag.StringVar(&c.storage.engine, "s.engine", mnemosyne.StorageEnginePostgres, "storage engine") // TODO: change to in memory when implemented
	flag.StringVar(&c.storage.postgres.address, "s.p.address", "postgres://localhost:5432?sslmode=disable", "storage postgres connection string")
	flag.StringVar(&c.storage.postgres.table, "s.p.table", "session", "storage postgres table name")
	flag.BoolVar(&c.tls.enabled, "tls", false, "tls enable flag")
	flag.StringVar(&c.tls.certFile, "tls.certfile", "", "path to tls cert file")
	flag.StringVar(&c.tls.keyFile, "tls.keyfile", "", "path to tls key file")
}

func (c *configuration) parse() {
	if !flag.Parsed() {
		ver := flag.Bool("version", false, "print version and exit")
		flag.Parse()
		if *ver {
			fmt.Printf("%s", VERSION)
			os.Exit(0)
		}
	}
}