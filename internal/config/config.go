package config

import (
	"github.com/rs/zerolog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "github.com/kardianos/minwinsvc" // import minwinsvc for windows services
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"gopkg.in/robfig/cron.v2"
)

var (
	// Path of config file
	Path string

	// General config
	General = &general{}
	// Mail config
	Mail = &mail{}
	// DB config
	DB = &db{}
	// Log config
	Log = &log{}

	// AppWorkPath of binary
	AppWorkPath string
	isWindows   bool
)

// general defines the general configuration.
type general struct {
	Root     string `ini:"ROOT"`
	Schedule string `ini:"SCHEDULE"`
}

// mail defines the mailer configuration.
type mail struct {
	Server   string `ini:"SERVER"`
	Port     int    `ini:"PORT"`
	User     string `ini:"USER"`
	Password string `ini:"PASSWORD"`

	From    string `ini:"FROM"`
	To      string `ini:"TO"`
	Subject string `ini:"SUBJECT"`
}

// db defines the database configuration.
type db struct {
	Server   string `ini:"SERVER"`
	Port     int    `ini:"PORT"`
	User     string `ini:"USER"`
	Password string `ini:"PASSWORD"`

	Database string `ini:"DATABASE"`
}

// log defines the logging configuration.
type log struct {
	Level   string `ini:"LEVEL"`
	Colored bool   `ini:"COLORED"`
	Pretty  bool   `ini:"PRETTY"`
}

// Load loads the configuration from `Path`
func Load() error {
	isWindows = runtime.GOOS == "windows"

	var appPath string
	var err error
	if appPath, err = getAppPath(); err != nil {
		return errors.Wrap(err, "could not get application path")
	}

	AppWorkPath = getWorkPath(appPath)

	if !filepath.IsAbs(Path) {
		Path = path.Join(AppWorkPath, Path)
	}

	config, err := ini.Load(Path)
	if err != nil {
		return errors.Wrap(err, "could not load ini config")
	}

	if err = config.Section("general").MapTo(General); err != nil {
		return errors.Wrap(err, "could not map general section")
	}

	if !filepath.IsAbs(General.Root) {
		General.Root = path.Join(AppWorkPath, General.Root)
	}
	if err := os.MkdirAll(General.Root, os.ModePerm); err != nil {
		return errors.Wrap(err, "could not create folders of root path")
	}

	schedule, err := cron.Parse(General.Schedule)
	if err != nil {
		return errors.Wrap(err, "could not parse cron expression")
	}
	// calculate interval
	nextExecutionTime := schedule.Next(time.Now())
	interval := schedule.Next(nextExecutionTime).Sub(nextExecutionTime)
	// check if interval is longer than 15 minutes
	if interval.Minutes() < 15 {
		return errors.New("schedule interval shorter than 15 minutes")
	}

	if err = config.Section("mail").MapTo(Mail); err != nil {
		return errors.Wrap(err, "could not map mail section")
	}

	if err = config.Section("db").MapTo(DB); err != nil {
		return errors.Wrap(err, "could not map db section")
	}

	if err = config.Section("log").MapTo(Log); err != nil {
		return errors.Wrap(err, "could not map log section")
	}

	if _, err := zerolog.ParseLevel(Log.Level); err != nil {
		return errors.Wrap(err, "could not parse log level")
	}

	return nil
}

func getAppPath() (string, error) {
	var appPath string

	if isWindows && filepath.IsAbs(os.Args[0]) {
		appPath = filepath.Clean(os.Args[0])
	} else {
		var err error
		appPath, err = exec.LookPath(os.Args[0])
		if err != nil {
			return "", errors.Wrapf(err, "could not find %s", os.Args[0])
		}
	}

	appPath, err := filepath.Abs(appPath)
	if err != nil {
		return "", errors.Wrapf(err, "could not create the absolute filepath of %s", appPath)
	}

	// Note: we don't use path.Dir here because it does not handle case
	//		 which path starts with two "/" in Windows: "//psf/Home/..."
	return strings.Replace(appPath, "\\", "/", -1), nil
}

func getWorkPath(appPath string) string {
	workPath := ""

	i := strings.LastIndex(appPath, "/")
	if i == -1 {
		workPath = appPath
	} else {
		workPath = appPath[:i]
	}

	// Note: we don't use path.Dir here because it does not handle case
	//		 which path starts with two "/" in Windows: "//psf/Home/..."
	return strings.Replace(workPath, "\\", "/", -1)
}
