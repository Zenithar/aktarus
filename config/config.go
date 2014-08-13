package config

import (
	"flag"
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
)

type (
	ircSettings struct {
		Host          string
		Port          string
		Nick          string
		NickPass      string
		Username      string
		Pass          string
		Ssl           bool
		NormalChannel string
		StaffChannel  string
		MaxFailures   int
		Timeout       uint
		KeepAlive     uint
		PingFreq      uint
		AutoVoice     bool
		Version       string
		Debug         bool
		PluginsDir    string
	}

	Settings struct {
		Irc   ircSettings
		Debug bool
	}
)

func Load() *Settings {
	log.New(os.Stdout, "[config]", log.LstdFlags)

	// Gets the current executable path for use a cfgfile path
	cwd, err := os.Getwd()
	// Bail out if the executable can not be found
	if err != nil {
		// Init here since we haven't done so yet
		log.Fatal("Can not find current working directory!")
	}

	log.Println("Loading...")

	// Default config path
	configPath := "config.toml"

	// Allow setting of the cfg file to load stuff from
	flag.StringVar(&configPath, "cfg", configPath, "sets the config file")

	// Parse command line settings
	if !flag.Parsed() {
		flag.Parse()
	}

	var cfg Settings
	if _, err := toml.DecodeFile(configPath, &cfg); err != nil {
		log.Fatalln(err)
	}

	if cfg.Irc.PluginsDir == "" {
		cfg.Irc.PluginsDir = filepath.Join(cwd, "js")
	}

	log.Println("Loaded config")

	return &cfg
}
