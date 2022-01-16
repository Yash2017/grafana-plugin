package plugin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"

	//"math/rand"
	"time"

	"github.com/buger/jsonparser"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/live"
	"golang.org/x/net/websocket"
)

// Make sure SampleDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler, backend.StreamHandler interfaces. Plugin should not
// implement all these interfaces - only those which are required for a particular task.
// For example if plugin does not need streaming functionality then you are free to remove
// methods that implement backend.StreamHandler. Implementing instancemgmt.InstanceDisposer
// is useful to clean up resources used by previous datasource instance when a new datasource
// instance created upon datasource settings changed.
var (
	_ backend.QueryDataHandler      = (*SampleDatasource)(nil)
	_ backend.CheckHealthHandler    = (*SampleDatasource)(nil)
	_ backend.StreamHandler         = (*SampleDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*SampleDatasource)(nil)
)

// NewSampleDatasource creates a new datasource instance.
func NewSampleDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	return &SampleDatasource{}, nil
}

// SampleDatasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type SampleDatasource struct{}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *SampleDatasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *SampleDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	log.DefaultLogger.Info("QueryData called", "request", req)

	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		log.DefaultLogger.Info("This is the Query ", q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type qModel struct {
	queryText     string `json:"queryText"`
	WithStreaming bool   `json:"withStreaming"`

	//constant      int    `json:"constant"`
}

func (d *SampleDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}
	log.DefaultLogger.Info("query is ", query.JSON)

	val, err := jsonparser.GetString(query.JSON, "queryText")
	if err != nil {
		log.DefaultLogger.Info("JsonParser error is", err)
	}
	log.DefaultLogger.Info("This is the val string", val)
	dat, err := base64.StdEncoding.DecodeString(string(val))
	if err != nil {
		//
	}

	log.DefaultLogger.Info("This is the val string after decoding", dat)
	// Unmarshal the JSON into our queryModel.
	var qm qModel

	response.Error = json.Unmarshal(query.JSON, &qm)
	log.DefaultLogger.Info("Error is", response.Error)
	if response.Error != nil {
		return response
	}

	log.DefaultLogger.Info("Qm is ", qm)

	// create data frame response.
	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []int64{10, 20}),
	)

	// If query called with streaming on then return a channel
	// to subscribe on a client-side and consume updates from a plugin.
	// Feel free to remove this if you don't need streaming for your datasource.
	if qm.WithStreaming {
		channel := live.Channel{
			Scope:     live.ScopeDatasource,
			Namespace: pCtx.DataSourceInstanceSettings.UID,
			Path:      "stream",
		}
		frame.SetMeta(&data.FrameMeta{Channel: channel.String()})
	}

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *SampleDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	log.DefaultLogger.Info("CheckHealth called", "request", req)

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

// SubscribeStream is called when a client wants to connect to a stream. This callback
// allows sending the first message.
func (d *SampleDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	log.DefaultLogger.Info("SubscribeStream called", "request", req)

	status := backend.SubscribeStreamStatusPermissionDenied
	if req.Path == "stream" {
		// Allow subscribing only on expected path.
		status = backend.SubscribeStreamStatusOK
	}
	return &backend.SubscribeStreamResponse{
		Status: status,
	}, nil
}

// RunStream is called once for any open channel.  Results are shared with everyone
// subscribed to the same channel.

func DoneAsync(val chan []byte) {
	//r := make(chan []byte, 5096)

	config, err := websocket.NewConfig("ws://localhost:8080/vui/platforms/volttron1/pubsub/devices/Campus/Building1/Fake1/all", "ws://localhost:8080/vui/platforms/volttron1/pubsub/devices/Campus/Building1/Fake1/all")
	if err != nil {

	}
	config.Header = http.Header{
		"Authorization": {"Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJncm91cHMiOlsiYWRtaW4iXSwiaWF0IjoxNjM3MDIyMzE2LCJuYmYiOjE2MzcwMjIzMTYsImV4cCI6MTYzNzAyMzIxNiwiZ3JhbnRfdHlwZSI6ImFjY2Vzc190b2tlbiJ9.V7WhUBwEc0rBq4jAtKbeWizSJErr5nVMd6kGcoh4k9g"},
	}
	ws, err := websocket.DialConfig(config)
	if err != nil {

	}
	var n int
	retMsg := make([]byte, 5096)
	for {
		n, err = ws.Read(retMsg)
		val <- retMsg[:n]
		//log.DefaultLogger.Info("This is the response", string(<-val))

	}
	//if err != nil {
	//log.DefaultLogger.Info(string(retMsg[:n]))
	//}

	//log.DefaultLogger.Info(string(retMsg[:n]))
	//r <- retMsg[:n]
	//log.DefaultLogger.Info("This is the method", string(<-r))
	//return r

}

func (d *SampleDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	log.DefaultLogger.Info("RunStream called", "request", req)

	// Create the same data frame as for query data.
	frame := data.NewFrame("response")

	// Add fields (matching the same schema used in QueryData).
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, make([]time.Time, 1)),
		data.NewField("values", nil, make([]string, 1)),
	)

	counter := 0
	val := make(chan []byte, 5096)
	go DoneAsync(val)
	//log.DefaultLogger.Info("This is the val variable", string(<-val))
	// Stream data frames periodically till stream closed by Grafana.
	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Info("Context done, finish streaming", "path", req.Path)
			return nil
		case al := <-val:
			// Send new data periodically.
			log.DefaultLogger.Info("This is the strn variable", string(al))
			frame.Fields[0].Set(0, time.Now())
			frame.Fields[1].Set(0, string(al))

			counter++

			err := sender.SendFrame(frame, data.IncludeAll)
			if err != nil {
				log.DefaultLogger.Error("Error sending frame", "error", err)
				continue
			}
		}
	}
}

// PublishStream is called when a client sends a message to the stream.
func (d *SampleDatasource) PublishStream(_ context.Context, req *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	log.DefaultLogger.Info("PublishStream called", "request", req)

	// Do not allow publishing at all.
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}
