package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/devopsext/sre/common"
	"github.com/devopsext/utils"
	"io/ioutil"
	"net/http"
	"time"
)

type SlackOptions struct {
	WebHook string
	Tags    string
	Timeout int
}

type SlackEventer struct {
	options SlackOptions
	logger  common.Logger
	tags    []string
	client  *http.Client
	ctx     context.Context
}

func (se *SlackEventer) Now(name string, attributes map[string]string) {
	se.At(name, attributes, time.Now())
}

func (se *SlackEventer) At(name string, attributes map[string]string, when time.Time) {
	se.Interval(name, attributes, when, when)
}

func (se *SlackEventer) Interval(name string, attributes map[string]string, begin, end time.Time) {
	var body string

	body = fmt.Sprintf("{\"text\": \"%s\"}", name)
	if payload, ok := attributes["payload"]; ok {
		body = payload
	}

	resp, err := http.Post(se.options.WebHook, "application/json", bytes.NewBuffer([]byte(body)))
	if err != nil {
		se.logger.Error("slack post:", err)
		return
	}

	defer resp.Body.Close()

	rBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		se.logger.Error("slack post response:", err)
		return
	}
	se.logger.Debug(string(rBody))
}

func (se *SlackEventer) Stop() {
	se.client.CloseIdleConnections()
	se.logger.Info("Slack Eventer stopped.")
}

func NewSlackEventer(options SlackOptions, logger common.Logger, stdout *Stdout) *SlackEventer {
	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.WebHook) {
		logger.Debug("Slack Eventer is disabled")
		return nil
	}

	logger.Info("Slack Eventer is up…")

	return &SlackEventer{
		options: options,
		logger:  logger,
		tags:    common.MapToArray(common.GetKeyValues(options.Tags)),
		client:  common.MakeHttpClient(options.Timeout),
		ctx:     context.Background(),
	}
}
