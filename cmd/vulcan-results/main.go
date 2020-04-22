//go:generate goagen bootstrap -d github.com/adevinta/vulcan-results/design

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	api "github.com/adevinta/vulcan-results"
	"github.com/adevinta/vulcan-results/app"
	"github.com/adevinta/vulcan-results/storage"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/goadesign/goa"
	goalogrus "github.com/goadesign/goa/logging/logrus"
	"github.com/goadesign/goa/middleware"
	"github.com/sirupsen/logrus"
)

//Config represents the configuration for vulcan-results
type Config struct {
	LogFile string
	Port    int
	Debug   bool

	Storage storage.Config `toml:"Storage"`
}

func main() {
	config := mustReadConfig()

	// Setup the logger with Logrus. If no LogFile specified, Stderr will be used.
	var lw io.Writer
	if config.LogFile != "" {
		f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("error: cannot open logfile (%v)", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}()

		lw = f
	} else {
		lw = os.Stderr
	}

	log := logrus.New()
	log.Out = lw
	log.Formatter = &logrus.TextFormatter{FullTimestamp: true}
	if config.Debug {
		log.Level = logrus.DebugLevel
	}
	logger := log.WithFields(logrus.Fields{
		"app": "VULCAN-RESULTS",
	})

	// Create service
	service := goa.New("vulcan-results")
	service.WithLogger(goalogrus.FromEntry(logger))

	// Mount middleware
	service.Use(middleware.RequestID())
	service.Use(middleware.LogRequest(config.Debug == true))
	service.Use(middleware.ErrorHandler(service, true))
	service.Use(middleware.Recover())

	// Mount "Results" controller
	sess, err := session.NewSession(&aws.Config{Region: &config.Storage.Region})
	if err != nil {
		service.LogError("aws session", "err", err)
		panic(err)
	}
	svc := s3.New(sess)

	if len(config.Storage.Endpoint) > 0 {
		svc = s3.New(sess, aws.NewConfig().WithEndpoint(config.Storage.Endpoint).WithS3ForcePathStyle(config.Storage.PathStyle))
	}

	st := storage.NewS3Storage(config.Storage, logger, svc)

	c := api.NewResultsController(service, st)
	app.MountResultsController(service, c)

	// Healthcheck controller
	c2 := api.NewHealthcheckController(service)
	app.MountHealthcheckController(service, c2)

	// Start service
	if err := service.ListenAndServe(fmt.Sprintf(":%v", config.Port)); err != nil {
		service.LogError("startup", "err", err)
	}
}

func mustReadConfig() Config {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: vulcan-results config_file")
	}
	configFile := os.Args[1]

	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("error: cannot read configuration file (%v)", err)
	}

	var config Config
	if _, err := toml.Decode(string(configData), &config); err != nil {
		log.Fatalf("error: cannot decode configuration file (%v)", err)
	}

	return config
}
