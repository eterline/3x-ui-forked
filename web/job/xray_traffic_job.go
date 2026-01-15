package job

import (
	"encoding/json"
	"time"

	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/websocket"
	"github.com/mhsanaei/3x-ui/v2/xray"

	"github.com/valyala/fasthttp"
)

// XrayTrafficJob collects and processes traffic statistics from Xray, updating the database and optionally informing external APIs.
type XrayTrafficJob struct {
	settingService  service.SettingService
	xrayService     service.XrayService
	inboundService  service.InboundService
	outboundService service.OutboundService
}

// NewXrayTrafficJob creates a new traffic collection job instance.
func NewXrayTrafficJob() *XrayTrafficJob {
	return new(XrayTrafficJob)
}

// Run collects traffic statistics from Xray and updates the database, triggering restart if needed.
func (j *XrayTrafficJob) Run() {
	if !j.xrayService.IsXrayRunning() {
		return
	}
	traffics, clientTraffics, err := j.xrayService.GetXrayTraffic()
	if err != nil {
		return
	}
	err, needRestart0 := j.inboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Warning("add inbound traffic failed:", err)
	}
	err, needRestart1 := j.outboundService.AddTraffic(traffics, clientTraffics)
	if err != nil {
		logger.Warning("add outbound traffic failed:", err)
	}
	if ExternalTrafficInformEnable, err := j.settingService.GetExternalTrafficInformEnable(); ExternalTrafficInformEnable {
		j.informTrafficToExternalAPI(traffics, clientTraffics)
	} else if err != nil {
		logger.Warning("get ExternalTrafficInformEnable failed:", err)
	}
	if needRestart0 || needRestart1 {
		j.xrayService.SetToNeedRestart()
	}

	// Get online clients and last online map for real-time status updates
	onlineClients := j.inboundService.GetOnlineClients()
	lastOnlineMap, err := j.inboundService.GetClientsLastOnline()
	if err != nil {
		logger.Warning("get clients last online failed:", err)
		lastOnlineMap = make(map[string]int64)
	}

	// Fetch updated inbounds from database with accumulated traffic values
	// This ensures frontend receives the actual total traffic, not just delta values
	updatedInbounds, err := j.inboundService.GetAllInbounds()
	if err != nil {
		logger.Warning("get all inbounds for websocket failed:", err)
	}

	updatedOutbounds, err := j.outboundService.GetOutboundsTraffic()
	if err != nil {
		logger.Warning("get all outbounds for websocket failed:", err)
	}

	// Broadcast traffic update via WebSocket with accumulated values from database
	trafficUpdate := map[string]interface{}{
		"traffics":       traffics,
		"clientTraffics": clientTraffics,
		"onlineClients":  onlineClients,
		"lastOnlineMap":  lastOnlineMap,
	}
	websocket.BroadcastTraffic(trafficUpdate)

	// Broadcast full inbounds update for real-time UI refresh
	if updatedInbounds != nil {
		websocket.BroadcastInbounds(updatedInbounds)
	}

	if updatedOutbounds != nil {
		websocket.BroadcastOutbounds(updatedOutbounds)
	}

}

type Network struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}
type Client struct {
	ID         int     `json:"id"`
	InboundID  int     `json:"inbound_id"`
	Enable     bool    `json:"enable"`
	Email      string  `json:"email"`
	UUID       string  `json:"uuid"`
	SubID      string  `json:"sub_id"`
	Network    Network `json:"network"`
	AllTime    int64   `json:"all_time"`
	ExpiryTime int64   `json:"expiry_time"`
	Total      int64   `json:"total"`
	LastOnline int64   `json:"last_online"`
	Reset      int     `json:"reset"`
}
type Endpoint struct {
	Tag     string  `json:"tag"`
	Type    string  `json:"type"`
	Network Network `json:"network"`
}

func setType(isInbound, isOutbound bool) string {
	switch {
	case isInbound:
		return "inbound"
	case isOutbound:
		return "outbound"
	default:
		return "unknown"
	}
}

func xrayTrafficsToEndpoints(inboundTraffics []*xray.Traffic) []Endpoint {
	s := make([]Endpoint, len(inboundTraffics))
	for i, tr := range inboundTraffics {
		s[i] = Endpoint{
			Tag:  tr.Tag,
			Type: setType(tr.IsInbound, tr.IsOutbound),
			Network: Network{
				Up:   tr.Up,
				Down: tr.Down,
			},
		}
	}
	return s
}

type TrafficToExternal struct {
	Timestamp string     `json:"timestamp"`
	Clients   []Client   `json:"clients"`
	Endpoints []Endpoint `json:"endpoints"`
}

func newTrafficToExternal(inboundTraffics []*xray.Traffic, clientTraffics []*xray.ClientTraffic) TrafficToExternal {
	return TrafficToExternal{
		Timestamp: time.Now().Format(time.RFC3339),
		Clients:   xrayClientsToClients(clientTraffics),
		Endpoints: xrayTrafficsToEndpoints(inboundTraffics),
	}
}

func xrayClientsToClients(clientTraffics []*xray.ClientTraffic) []Client {
	s := make([]Client, len(clientTraffics))
	for i, tr := range clientTraffics {
		s[i] = Client{
			ID:        tr.Id,
			InboundID: tr.InboundId,
			Enable:    tr.Enable,
			Email:     tr.Email,
			UUID:      tr.UUID,
			SubID:     tr.SubId,
			Network: Network{
				Up:   tr.Up,
				Down: tr.Down,
			},
			AllTime:    tr.AllTime,
			ExpiryTime: tr.ExpiryTime,
			Total:      tr.Total,
			Reset:      tr.Reset,
			LastOnline: tr.LastOnline,
		}
	}
	return s
}

func (j *XrayTrafficJob) informTrafficToExternalAPI(inboundTraffics []*xray.Traffic, clientTraffics []*xray.ClientTraffic) {
	informURL, err := j.settingService.GetExternalTrafficInformURI()
	if err != nil {
		logger.Warning("get ExternalTrafficInformURI failed:", err)
		return
	}

	stats := newTrafficToExternal(inboundTraffics, clientTraffics)

	requestBody, err := json.Marshal(stats)
	if err != nil {
		logger.Warning("parse client/inbound traffic failed:", err)
		return
	}
	request := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(request)
	request.Header.SetMethod("POST")
	request.Header.SetContentType("application/json; charset=UTF-8")
	request.SetBody([]byte(requestBody))
	request.SetRequestURI(informURL)
	response := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(response)
	if err := fasthttp.Do(request, response); err != nil {
		logger.Warning("POST ExternalTrafficInformURI failed:", err)
	}
}
