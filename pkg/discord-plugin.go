package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type DiscordQuery struct {
	Constant      float64 `json:"constant"`
	Datasource    string  `json:"datasource"`
	DatasourceID  int     `json:"datasourceId"`
	IntervalMs    int     `json:"intervalMs"`
	MaxDataPoints int     `json:"maxDataPoints"`
	OrgID         int     `json:"orgId"`
	QueryText     string  `json:"queryText"`
	RefID         string  `json:"refId"`
}

type DiscordPluginConfig struct {
	DiscordToken string `json:"discordToken"`
}

// newDiscordDataSource returns datasource.ServeOpts.
func newDiscordDataSource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)

	token := os.Getenv("DISCORD_TOKEN")
	ds := &DiscordDataSource{
		im:           im,
		discordToken: token,
	}

	return datasource.ServeOpts{
		QueryDataHandler:   ds,
		CheckHealthHandler: ds,
	}
}

// DiscordDataSource.... all things discord!
type DiscordDataSource struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirements
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager

	// Discord Token
	discordToken string
	host         string
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifer).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (td *DiscordDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	configBytes, _ := req.PluginContext.DataSourceInstanceSettings.JSONData.MarshalJSON()
	var config DiscordPluginConfig
	err := json.Unmarshal(configBytes, &config)
	if err != nil {
		return nil, err
	}
	td.discordToken = config.DiscordToken

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res, err := td.query(ctx, q)
		if err != nil {
			return nil, err
		}

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = *res
	}

	return response, nil
}

type queryModel struct {
	Format string `json:"format"`
}

func (td *DiscordDataSource) queryDiscord(fromDate time.Time, toDate time.Time) (string, error) {
	return "discord answers!!!", nil
}

func (td *DiscordDataSource) query(ctx context.Context, query backend.DataQuery) (*backend.DataResponse, error) {
	// Unmarshal the json into our queryModel
	var qm queryModel

	queryBytes, _ := query.JSON.MarshalJSON()
	var dQuery DiscordQuery
	err := json.Unmarshal(queryBytes, &dQuery)
	if err != nil {
		// empty response? or real error? figure out later.
		return nil, err
	}

	response := backend.DataResponse{}
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return nil, err
	}

	// Log a warning if `Format` is empty.
	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
	}

	// sgStats, err := td.querySendGrid(query.TimeRange.From, query.TimeRange.To)
	if err != nil {
		return nil, err
	}

	// create data frame response
	frame := data.NewFrame("response")

	// generate time slice.
	times := []time.Time{}
	something := []int64{}

	/*
		// go through
		for _, res := range *sgStats {
			t, _ := time.Parse("2006-01-02", res.Date)
			times = append(times, t)

			requests = append(requests, int64(res.Stats[0].Metrics.Requests))
			blocks = append(blocks, int64(res.Stats[0].Metrics.Blocks))
			bounceDrops = append(bounceDrops, int64(res.Stats[0].Metrics.BounceDrops))
			bounces = append(bounces, int64(res.Stats[0].Metrics.Bounces))
			clicks = append(clicks, int64(res.Stats[0].Metrics.Clicks))
			deferred = append(deferred, int64(res.Stats[0].Metrics.Deferred))
			delivered = append(delivered, int64(res.Stats[0].Metrics.Delivered))
			invalidEmails = append(invalidEmails, int64(res.Stats[0].Metrics.InvalidEmails))
			opens = append(opens, int64(res.Stats[0].Metrics.Opens))
			processed = append(processed, int64(res.Stats[0].Metrics.Processed))
			spamReportDrops = append(spamReportDrops, int64(res.Stats[0].Metrics.SpamReportDrops))
			spamReports = append(spamReports, int64(res.Stats[0].Metrics.SpamReports))
			uniqueClicks = append(uniqueClicks, int64(res.Stats[0].Metrics.UniqueClicks))
			uniqueOpens = append(uniqueOpens, int64(res.Stats[0].Metrics.UniqueOpens))
			unsubscribeDrops = append(unsubscribeDrops, int64(res.Stats[0].Metrics.UnsubscribeDrops))
			unsubscribes = append(unsubscribes, int64(res.Stats[0].Metrics.Unsubscribes))
		}
	*/

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
	)

	frame.Fields = append(frame.Fields,
		data.NewField("something", nil, something),
	)

	// add the frames to the response
	response.Frames = append(response.Frames, frame)
	return &response, nil
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (td *DiscordDataSource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	configBytes, _ := req.PluginContext.DataSourceInstanceSettings.JSONData.MarshalJSON()
	var config DiscordPluginConfig
	err := json.Unmarshal(configBytes, &config)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Cannot get healthcheck query : %s", err.Error()))
		status = backend.HealthStatusError
		message = "Unable to contact Sendgrid"
	}

	/*
	  td.sendgridApiKey = config.SendgridAPIKey
		from := time.Now().UTC().Add(-24*time.Hour)
	  to := time.Now().UTC()
	*/

	/*
	  _, err = td.querySendGrid(from, to)
	  if err != nil {
	    log.DefaultLogger.Error(fmt.Sprintf("Cannot query sendgrid for healthcheck: %s", err.Error()))
	    status = backend.HealthStatusError
	    message = "Unable to contact Sendgrid"
	  }
	*/

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

type instanceSettings struct {
	httpClient *http.Client
}

func newDataSourceInstance(setting backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &instanceSettings{
		httpClient: &http.Client{},
	}, nil
}

func (s *instanceSettings) Dispose() {
	// Called before creatinga a new instance to allow plugin authors
	// to cleanup.
}
