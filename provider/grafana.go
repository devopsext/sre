package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/devopsext/sre/common"
	utils "github.com/devopsext/utils"
)

type GrafanaAnnotationResponse struct {
	Message string `json:"message"`
	ID      int    `json:"id"`
}

type GrafanaAnnotation struct {
	Time    int      `json:"time"`
	TimeEnd int      `json:"timeEnd"`
	Tags    []string `json:"tags"`
	Text    string   `json:"text"`
}

type GrafanaOptions struct {
	URL     string
	ApiKey  string
	Tags    string
	Version string
	Timeout int
}

type GrafanaEventerOptions struct {
	GrafanaOptions
	Endpoint string
	Duration int
}

type GrafanaEventer struct {
	options GrafanaEventerOptions
	logger  common.Logger
	tags    []string
	client  *http.Client
	ctx     context.Context
}

func (ge *GrafanaEventer) httpDoRequest(method, query string, params url.Values, buf io.Reader) ([]byte, int, error) {
	u, _ := url.Parse(ge.options.URL)
	u.Path = path.Join(u.Path, query)
	if params != nil {
		u.RawQuery = params.Encode()
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, 0, err
	}
	req = req.WithContext(ge.ctx)
	if !strings.Contains(ge.options.ApiKey, ":") {
		req.Header.Set("Authorization", "Bearer "+ge.options.ApiKey)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := ge.client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return data, resp.StatusCode, err
}

func (ge *GrafanaEventer) httpPost(query string, params url.Values, body []byte) ([]byte, int, error) {
	return ge.httpDoRequest("POST", query, params, bytes.NewBuffer(body))
}

func (ge *GrafanaEventer) createAnnotation(a GrafanaAnnotation) (*GrafanaAnnotationResponse, error) {

	b, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}

	raw, code, err := ge.httpPost(ge.options.Endpoint, nil, b)
	if err != nil {
		return nil, err
	}
	if code != 200 {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}

	var res GrafanaAnnotationResponse
	err = json.Unmarshal(raw, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (ge *GrafanaEventer) Interval(name string, attributes map[string]string, begin, end time.Time) {

	tags := utils.MapToArray(attributes)
	for _, v := range ge.tags {
		if !utils.Contains(tags, v) {
			tags = append(tags, v)
		}
	}

	a := GrafanaAnnotation{
		Time:    int(begin.UTC().UnixMilli()),
		TimeEnd: int(end.UTC().UnixMilli()),
		Tags:    tags,
		Text:    name,
	}

	ar, err := ge.createAnnotation(a)
	if err != nil {
		ge.logger.Error(err)
		return
	}
	ge.logger.Debug("Annotation %d. %s", ar.ID, ar.Message)
}

func (ge *GrafanaEventer) Now(name string, attributes map[string]string) {
	ge.At(name, attributes, time.Now())
}

func (ge *GrafanaEventer) At(name string, attributes map[string]string, when time.Time) {
	ge.Interval(name, attributes, when, when.Add(time.Second*time.Duration(ge.options.Duration)))
}

func (ge *GrafanaEventer) Stop() {
	// nothing here
}

func NewGrafanaEventer(options GrafanaEventerOptions, logger common.Logger, stdout *Stdout) *GrafanaEventer {

	if logger == nil {
		logger = stdout
	}

	if utils.IsEmpty(options.URL) && utils.IsEmpty(options.Endpoint) {
		stdout.Debug("Grafana eventer is disabled.")
		return nil
	}

	logger.Info("Grafana eventer is up...")

	return &GrafanaEventer{
		options: options,
		logger:  logger,
		tags:    utils.MapToArray(utils.MapGetKeyValues(options.Tags)),
		client:  utils.NewHttpInsecureClient(options.Timeout),
		ctx:     context.Background(),
	}
}
