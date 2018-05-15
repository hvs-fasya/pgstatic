package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"fmt"
	"github.com/miniflux/miniflux/logger"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"service-for-telegram-bot/env/gopat/src/github.com/pkg/errors"
)

const (
	appVersion        = "0.0.1"
	DefaultConfigPath = "config.json"
)

var (
	configPath string
	debug      bool
	Logger     *AppLogger
)

//App general application structure; contains initialized at entry-point components reusable while application run
type App struct {
	Cfg    *Cfg
	DB     *sql.DB
	Client *http.Client
}

// Cfg structure containing application components config structures
type Cfg struct {
	Db        string `yaml:"db"`
	ClientUrl string `yaml:"client"`
}

type AppLogger struct {
	Logger *log.Logger
	Debug  bool
}

func main() {

	var err error

	app := cli.NewApp()
	app.Usage = "cli agent for database structure revision"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config, c",
			Value:       DefaultConfigPath,
			Usage:       "path to config file",
			Destination: &configPath,
		},
		cli.BoolFlag{
			Name:        "verbose, V",
			Usage:       "print logging info if true",
			Destination: &debug,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "start application",
			Action: func(c *cli.Context) error {
				err := startApp(c)
				if err != nil {
					return cli.NewExitError(err, 1)
				}
				return nil
			},
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		Logger.Logger.Fatal(err)
	}
}

func startApp(c *cli.Context) (err error) {
	app := new(App)

	//read config
	app.Cfg, err = newCfg(configPath)
	if err != nil {
		return err
	}

	//set logging
	if err = setLogger(debug); err != nil {
		return err
	}

	//init database conn
	if err = app.initDb(); err != nil {
		return err
	}
	defer app.DB.Close()

	//init http client
	if err = app.initClient(); err != nil {
		return err
	}

	return nil
}

func newCfg(path string) (*Cfg, error) {
	var content []byte
	var err error
	cfg := new(Cfg)
	if content, err = ioutil.ReadFile(path); err != nil {
		return cfg, errors.New(fmt.Sprintf("failed to read config file %s: %s", path, err))
	}
	if err = yaml.Unmarshal(content, &cfg); err != nil {
		return cfg, errors.New(fmt.Sprintf("failed to unmarshal config %s: %s", path, err))
	}
	return cfg, nil
}

func setLogger(d bool) error {
	Logger := new(AppLogger)
	Logger.Logger = new(log.Logger)
	Logger.Debug = d
	return nil
}

//Debugf print logging info to stdOut only if debug mode is true
func (al *AppLogger) Debugf(format string, v ...interface{}) {
	if al.Debug {
		al.Logger.Printf(format, v)
	}
}
