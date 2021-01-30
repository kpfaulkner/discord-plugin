package main

import (
	"context"
	"encoding/json"
	"fmt"
	log2 "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const (
	NumberOfUsersQuery       string = "numusers"
	NumberOfUsersJoinedQuery string = "numjoined"
	NumberOfUsersLeftQuery   string = "numleft"
	NumberOfMessagesQuery    string = "nummessages"
)

type DataPoint struct {
	TimeStamp time.Time
	Val       int64
}

type DiscordQuery struct {
	Constant      float64 `json:"constant"`
	Datasource    string  `json:"datasource"`
	DatasourceID  int     `json:"datasourceId"`
	IntervalMs    int     `json:"intervalMs"`
	MaxDataPoints int     `json:"maxDataPoints"`
	OrgID         int     `json:"orgId"`
	RGSplit       string  `json:"rgSplit"`
	RefID         string  `json:"refId"`
}

type DiscordPluginConfig struct {
	DiscordGuildID    string `json:"discordGuildID"`
	DiscordGuildIDKey string `json:"discordGuildIDKey"`
}

// newDiscordDataSource returns datasource.ServeOpts.
func newDiscordDataSource() datasource.ServeOpts {
	// creates a instance manager for your plugin. The function passed
	// into `NewInstanceManger` is called when the instance is created
	// for the first time or when a datasource configuration changed.
	im := datasource.NewInstanceManager(newDataSourceInstance)

	//token := os.Getenv("DISCORD_TOKEN")
	ds := &DiscordDataSource{
		im: im,
	}

	pluginExecutablePath := os.Args[0]
	pluginDir := filepath.Dir(pluginExecutablePath)
	config, err := LoadConfig(filepath.Join(pluginDir, "discord.json"))
	if err != nil {
		log2.Fatalf("Unable to read discord config. Exiting")
	}

	ds.dbHelper = NewAzureDBHelper(*config, "botbotgo", "bbgadmin")
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

	dbHelper AzureDBHelper

	// Discord Token
	discordGuildID    string
	discordGuildIDKey string

	host string
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
	td.discordGuildID = config.DiscordGuildID
	td.discordGuildIDKey = config.DiscordGuildIDKey

	log.DefaultLogger.Warn(fmt.Sprintf("guildid is %s\n", td.discordGuildID))

	fmt.Printf("req is %v\n", *req)
	log.DefaultLogger.Warn(fmt.Sprintf("req is %v\n", req.Queries))

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

	log.DefaultLogger.Warn(fmt.Sprintf("single query is  is %v\n", dQuery.RGSplit))

	title, queryResponse, err := td.doDiscordQuery(dQuery, query.TimeRange.From, query.TimeRange.To)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Unable to query discord", err.Error()))
		return nil, err
	}

	// consolidate data into grafana response.
	response := backend.DataResponse{}
	response.Error = json.Unmarshal(query.JSON, &qm)
	if response.Error != nil {
		return nil, err
	}

	// Log a warning if `Format` is empty.
	if qm.Format == "" {
		log.DefaultLogger.Warn("format is empty. defaulting to time series")
	}

	// create data frame response
	frame := data.NewFrame("response")

	times := []time.Time{}
	counts := []int64{}

	// go through
	for _, res := range queryResponse {
		//t, _ := time.Parse("2006-01-02", res.TimeStamp)
		times = append(times, res.TimeStamp)
		counts = append(counts, res.Val)
	}

	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, times),
	)

	frame.Fields = append(frame.Fields,
		data.NewField(title, nil, counts),
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

	// TODO(kpfaulkner) do check against SQL

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func (td *DiscordDataSource) getNumberOfMessages(query DiscordQuery, fromTime time.Time, toTime time.Time) ([]DataPoint, error) {

	// dummy data for now.
	data := []DataPoint{}

	resp, err := td.dbHelper.GetNumberMessageForGuildBetweenTimes(td.discordGuildID, fromTime, toTime, 5)
	if err != nil {
		return nil, err
	}

	for _, r := range resp {
		data = append(data, DataPoint{TimeStamp: r.TimeStamp, Val: r.Count})
	}

	return data, nil
}

func (td *DiscordDataSource) getNumberOfUsers(query DiscordQuery, fromTime time.Time, toTime time.Time) ([]DataPoint, error) {
	// dummy data for now.
	data := []DataPoint{}

	resp, err := td.dbHelper.GetMemberCountForGuildBetweenTimes(td.discordGuildID, fromTime, toTime, 5)
	if err != nil {
		return nil, err
	}

	for _, r := range resp {
		log2.Infof("Guild %s : time %s : %d members", td.discordGuildID, r.TimeStamp, r.Count)
		data = append(data, DataPoint{TimeStamp: r.TimeStamp, Val: r.Count})
	}

	return data, nil
}

func (td *DiscordDataSource) getNumberOfUsersJoined(query DiscordQuery, fromTime time.Time, toTime time.Time) ([]DataPoint, error) {
	// dummy data for now.
	data := []DataPoint{}

	resp, err := td.dbHelper.GetNumberJoinsForGuildBetweenTimes(td.discordGuildID, fromTime, toTime, 5)
	if err != nil {
		return nil, err
	}

	for _, r := range resp {
		data = append(data, DataPoint{TimeStamp: r.TimeStamp, Val: r.Count})
	}

	return data, nil
}

func (td *DiscordDataSource) getNumberOfUsersLeft(query DiscordQuery, fromTime time.Time, toTime time.Time) ([]DataPoint, error) {
	// dummy data for now.
	data := []DataPoint{}

	resp, err := td.dbHelper.GetNumberLeftForGuildBetweenTimes(td.discordGuildID, fromTime, toTime, 5)
	if err != nil {
		return nil, err
	}

	for _, r := range resp {
		data = append(data, DataPoint{TimeStamp: r.TimeStamp, Val: r.Count})
	}

	return data, nil
}

func (td *DiscordDataSource) doDiscordQuery(dQuery DiscordQuery, fromTime time.Time, toTime time.Time) (string, []DataPoint, error) {

	data := []DataPoint{}
	var err error
	var title string

	switch dQuery.RGSplit {
	case NumberOfMessagesQuery:
		data, err = td.getNumberOfMessages(dQuery, fromTime, toTime)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Unable to get number of discord messages: %s", err.Error()))
			return "", nil, err
		}
		title = "no-messages"

	case NumberOfUsersJoinedQuery:
		data, err = td.getNumberOfUsersJoined(dQuery, fromTime, toTime)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Unable to get number of discord users joined: %s", err.Error()))
			return "", nil, err
		}
		title = "no-joins"

	case NumberOfUsersLeftQuery:
		data, err = td.getNumberOfUsersLeft(dQuery, fromTime, toTime)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Unable to get number of discord users left: %s", err.Error()))
			return "", nil, err
		}
		title = "no-left"

	case NumberOfUsersQuery:
		data, err = td.getNumberOfUsers(dQuery, fromTime, toTime)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Unable to get number of discord users now: %s", err.Error()))
			return "", nil, err
		}
		title = "no-users"
	}

	return title, data, nil
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
