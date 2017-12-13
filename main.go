package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/briandowns/sky-island/config"
	"github.com/briandowns/sky-island/handlers"
	"github.com/briandowns/sky-island/jail"
	"github.com/briandowns/sky-island/utils"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
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

	conf, err := config.Load(configFlag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output := os.Stderr
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(output)
	logger := logrus.New()
	log.SetOutput(output)
	log.SetFlags(0)

	r, err := exec.Command("uname", "-r").Output()
	if err != nil {
		logger.Error(err)
		os.Exit(1)
	}
	conf.Release = strings.Trim(string(r), "\n")

	statsd.Address(conf.Jails.MonitoringAddr)
	metrics, err := statsd.New()
	if err != nil {
		log.Print(err)
	}
	defer metrics.Close()

	if initFlag {
		jsvc := jail.NewJailService(conf, logger, metrics, utils.Wrap{})
		if err := jsvc.InitializeSystem(); err != nil {
			logger.Info(err)
			os.Exit(0)
		}
		os.Exit(0)
	}

	logger.Info("starting API...")

	router := mux.NewRouter()
	params := handlers.Params{
		Logger:  logger,
		Conf:    conf,
		StatsMW: stats.New(),
		Metrics: metrics,
	}
	handlers.AddHandlers(router, &params)
	n := negroni.New(
		negroni.NewRecovery(),
		negroni.NewLogger(),
	)
	n.Use(params.StatsMW)
	n.UseHandler(router)
	n.Run(":" + strconv.Itoa(conf.HTTP.Port))
}
