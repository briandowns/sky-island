package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"

	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/handlers"
	"github.com/briandowns/sky-island/jail"
	"github.com/briandowns/sky-island/log"
	"github.com/briandowns/sky-island/utils"
	"github.com/codegangsta/negroni"
	"github.com/thoas/stats"
	"gopkg.in/alexcesaro/statsd.v2"
)

var (
	configFlag string
	initFlag   bool
)

var signalsChan = make(chan os.Signal, 1)

func main() {
	signal.Notify(signalsChan, os.Interrupt)
	go func() {
		for range signalsChan {
			signalsChan = nil
			os.Exit(1)
		}
	}()
	if os.Getgid() != 0 {
		fmt.Println("must be run with super user permissions")
		os.Exit(1)
	}
	flag.StringVar(&configFlag, "c", "", "sky-island configuration file")
	flag.BoolVar(&initFlag, "i", false, "initialize system")
	flag.Parse()
	if configFlag == "" {
		flag.Usage()
		os.Exit(1)
	}

	conf, err := config.Load(configFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logger, err := log.Logger(conf, "sky-island")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	statsd.Address(conf.Jails.MonitoringAddr)
	metrics, err := statsd.New()
	if err != nil {
		logger.Log("error", err.Error())
	}
	defer metrics.Close()

	if initFlag {
		jsvc := jail.NewJailService(conf, logger, metrics, utils.Wrap{})
		if err := jsvc.InitializeSystem(); err != nil {
			logger.Log("error", err.Error())
			os.Exit(0)
		}
		os.Exit(0)
	}

	logger.Log("msg", "starting API...")

	params := handlers.Params{
		Logger:  logger,
		Conf:    conf,
		StatsMW: stats.New(),
		Metrics: metrics,
	}
	router, err := handlers.AddHandlers(&params)
	if err != nil {
		logger.Log("error", err.Error())
		os.Exit(0)
	}
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
	)
	n.Use(params.StatsMW)
	n.UseHandler(router)
	n.Run(":" + strconv.Itoa(conf.HTTPPort))
}
