package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	gitlabapi "github.com/janusky/gitlab-api-client/gitlab"
	"github.com/janusky/gitlab-api-client/logging"
	"github.com/janusky/gitlab-api-client/utils"
	"github.com/pkg/errors"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:           "gitlab-api-client",
	Short:         "Gitlab API Client",
	Example:       "",
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

const (
	trustedCertificates = "trusted-certificates"
	gitlabAPIURL        = "gitlab.api-url"
	gitlabPrivateToken  = "gitlab.private-token"
)

var (
	trustedCertificatesVal []string
	cfgFile                string
	logformat              string
	logfile                string
	debug                  bool
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.api-client.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "v", false, "Print debug messages (includes info)")
	rootCmd.PersistentFlags().StringArrayVar(&trustedCertificatesVal, trustedCertificates, []string{}, "PEM encoded trusted certificate chain")
	rootCmd.PersistentFlags().StringVar(&logformat, "log-format", "dev", "Log format (json, log, dev, cli)")
	rootCmd.PersistentFlags().StringVar(&logfile, "log-file", "", "Log file path (''=Stderr|'-'=Stdout)")

	rootCmd.PersistentFlags().StringP("api-url", "u", "https://gitlab.localhost/api/v4/", "Gitlab URL")
	rootCmd.PersistentFlags().StringP("private-token", "t", "", "Your private token (grab it from https://gitlab.localhost/account)")
	viper.BindPFlags(rootCmd.PersistentFlags())

	viper.BindPFlag(gitlabAPIURL, rootCmd.PersistentFlags().Lookup("api-url"))
	viper.BindPFlag(gitlabPrivateToken, rootCmd.PersistentFlags().Lookup("private-token"))
}

func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName("." + rootCmd.Use)
	}

	viper.SetEnvPrefix(rootCmd.Use)
	viper.AutomaticEnv()
	viper.ReadInConfig()

	log.SetHandler(cli.New(os.Stderr))
	log.SetLevel(log.WarnLevel)

	err := logging.SetupLog(logformat, logfile, debug)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func gitlabAPI() (*gitlabapi.GitlabApi, error) {
	httpClient, err := httpClient()
	if err != nil {
		return nil, err
	}
	return gitlabapi.NewGitlabApi(httpClient, viper.GetString(gitlabAPIURL), viper.GetString(gitlabPrivateToken)), nil
}

func httpClient() (*http.Client, error) {
	if trustedCertificatesVal == nil || len(trustedCertificatesVal) == 0 {
		trustedCertificatesVal = viper.GetStringSlice(trustedCertificates)
	}

	certs, err := dereference(trustedCertificatesVal)
	if err != nil {
		return nil, err
	}

	httpClient, err := utils.HTTPClient(certs)
	if err != nil {
		return nil, errors.Wrap(err, "creating http client")
	}
	return httpClient, nil
}

func dereference(files []string) ([][]byte, error) {
	res := [][]byte{}
	for _, file := range files {
		if file[0] == '@' {
			bs, err := ioutil.ReadFile(file[1:len(file)])
			if err != nil {
				return nil, errors.Wrapf(err, "trying to dereference '%s'", file)
			}
			res = append(res, bs)
		} else {
			res = append(res, []byte(file))
		}
	}
	return res, nil
}
