package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

const (
	projectName = "merkle-tree"
)

var (
	// var
	cfgFile string

	// commands
	rootCmd = &cobra.Command{
		Use:   "./" + projectName,
		Short: "merkle tree",
	}
)

// Execute launches the CLI
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Panic(1)
	}
}

// initLogger inits the global logger with configuration from the conf file
func initLogger() {
	logLvl := viper.GetString(projectName + ".log.verbosity-level")
	logrusLogLvl, err := log.ParseLevel(logLvl)
	if err != nil {
		log.WithError(err).Fatalf("log.ParseLevel(%s): unable to parse log level", logLvl)
	}
	log.SetLevel(logrusLogLvl)

	log.WithFields(log.Fields{
		"verbosity-level": logrusLogLvl.String(),
	}).Info("logger has been initialized")
}

// initConfig loads the conf file and its properties as well as mapping env variable overloading the properties
// from the conf file
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			log.WithFields(log.Fields{
				"config-file": cfgFile,
			}).WithError(err).Fatal("viper.ReadInConfig(): unable to read configuration file")
		}

		log.WithFields(log.Fields{
			"config-file": cfgFile,
		}).Info("starting service using config file")
	} else {
		log.Info("starting service without config file")
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.AutomaticEnv()
}

func init() {
	cobra.OnInitialize(initConfig, initLogger)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	rootCmd.Flags().String("log-level", "info", "log verbosity level")
	_ = viper.BindPFlag(projectName+".log.verbosity-level", rootCmd.Flag("log-level"))
}
