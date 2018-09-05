package main

import (
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/emed-appts/emedappts-mailer/pkg/config"
	"github.com/emed-appts/emedappts-mailer/pkg/version"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	app := &cli.App{
		Name:        "emedappt-mailer",
		Version:     version.Version.String(),
		Usage:       "eMedical Appointments Mailer Service",
		Description: "Runs the mailer service, sending notifications about booked/cancelled appointments",
		Compiled:    time.Now(),

		Authors: []*cli.Author{
			{
				Name:  "David Schneiderbauer",
				Email: "david.schneiderbauer@dschneiderbauer.me",
			},
		},

		Before: func(c *cli.Context) error {
			return nil
		},

		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Value:       "conf/app.ini",
				Usage:       "set config path",
				Destination: &config.Path,
			},
		},

		Action: func(c *cli.Context) error {
			// load config
			err := config.Load()
			if err != nil {
				log.Fatal().
					Msgf("%+v\n", errors.Wrap(err, "could not load config"))

				return err
			}

			// open logfile
			logFile, err := os.OpenFile(path.Join(config.General.Root, "mailer.log"), os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
			if err != nil {
				log.Fatal().
					Msgf("%+v\n", errors.Wrap(err, "could not open log file"))

				return err
			}
			defer logFile.Close()

			// configure logger
			if config.Log.Pretty {
				log.Logger = log.Output(
					zerolog.ConsoleWriter{
						Out:     logFile,
						NoColor: !config.Log.Colored,
					},
				)
			} else {
				log.Logger = log.Output(logFile)
			}

			// set configured log level
			logLvl, err := zerolog.ParseLevel(config.Log.Level)
			if err != nil {
				log.Fatal().
					Msgf("%+v\n", errors.Wrap(err, "could not parse log level"))

				return  err
			}
			zerolog.SetGlobalLevel(logLvl)

			stop := make(chan os.Signal, 1)

			// todo run action

			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
			<-stop

			close(stop)

			return nil
		},
	}

	cli.HelpFlag = &cli.BoolFlag{
		Name:    "help",
		Aliases: []string{"h"},
		Usage:   "show the help, so what you see now",
	}

	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{"v"},
		Usage:   "print the current version of that tool",
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}