package provider

import (
	"github.com/devopsext/sre/common"
	utils "github.com/devopsext/utils"
)

type GrafanaOptions struct {
	URL     string
	ApiKey  string
	Tags    string
	Version string
}

type GrafanaEventerOptions struct {
	GrafanaOptions
	Endpoint string
}

type GrafanaEventer struct {
	options GrafanaEventerOptions
	logger  common.Logger
	tags    map[string]interface{}
}

func (ge *GrafanaEventer) Trigger() {

}

func NewGrafanaEventer(options GrafanaEventerOptions, logger common.Logger, stdout *Stdout) *GrafanaEventer {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.URL) {
		stdout.Debug("Grafana eventer is disabled.")
		return nil
	}

	tags := make(map[string]interface{})
	m := common.GetKeyValues(options.Tags)
	for k, v := range m {
		tags[k] = v
	}

	logger.Info("Grafana eventer is up...")

	return &GrafanaEventer{
		options: options,
		logger:  logger,
		tags:    tags,
	}
}
