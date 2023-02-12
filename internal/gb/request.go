package gb

import (
	"github.com/chenjianhao66/go-GB28181/internal/config"
	"github.com/chenjianhao66/go-GB28181/internal/log"
	"github.com/chenjianhao66/go-GB28181/internal/model"
	"github.com/ghettovoice/gosip/sip"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type (
	SIPFactory struct{}
	Sender     struct{}
)

// 发送请求之后的回调
type successCallback func(sip.ClientTransaction, error)

const (
	letterBytes    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	contentTypeXML = "Application/MANSCDP+xml"
	contentTypeSDP = "APPLICATION/SDP"
)

var (
	SipFactory SIPFactory
	SipSender  Sender
)

// TransmitRequest 发送sip协议请求
func (sender Sender) TransmitRequest(req sip.Request, callback successCallback) {
	log.Info("发送SIP Request消息，Method为: ", req.Method())
	transaction, err := s.s.Request(req)
	if callback != nil {
		callback(transaction, err)
	}
}

// CreateMessageRequest 创建Message类型请求
func (p SIPFactory) CreateMessageRequest(d model.Device, body string) sip.Request {
	requestBuilder := sip.NewRequestBuilder()
	requestBuilder.SetFrom(newFromAddress(newParams(map[string]string{"tag": randString(32)})))

	to := newTo(d.DeviceId, d.Ip)
	requestBuilder.SetTo(to)
	requestBuilder.SetRecipient(to.Uri)
	requestBuilder.AddVia(newVia(d.Transport))
	contentType := sip.ContentType(contentTypeXML)
	requestBuilder.SetContentType(&contentType)
	requestBuilder.SetMethod(sip.MESSAGE)
	userAgent := sip.UserAgentHeader("go-gb")
	requestBuilder.SetUserAgent(&userAgent)
	requestBuilder.SetBody(body)
	req, _ := requestBuilder.Build()
	return req
}

// CreateInviteRequest 创建invite请求
func (p SIPFactory) CreateInviteRequest() sip.Request {
	body := createSdpInfo()

	requestBuilder := sip.NewRequestBuilder()
	to := newTo("44010200491318000001", "192.168.1.222")
	requestBuilder.SetMethod(sip.INVITE)
	requestBuilder.SetFrom(newFromAddress(newParams(map[string]string{"tag": randString(32)})))
	requestBuilder.SetTo(to)
	sipUri := &sip.SipUri{
		FUser: sip.String{Str: "44010200491318000001"},
		FHost: to.Uri.Host(),
	}
	requestBuilder.SetRecipient(sipUri)
	requestBuilder.AddVia(newVia("UDP"))
	requestBuilder.SetContact(newTo("44010200492000000001", "192.168.1.223:5060"))
	contentType := sip.ContentType(contentTypeSDP)
	requestBuilder.SetContentType(&contentType)
	requestBuilder.SetBody(body)
	header := sip.GenericHeader{
		HeaderName: "Subject",
		Contents:   "44010200491318000001:0102008374,44010200492000000001:0",
	}
	requestBuilder.AddHeader(&header)
	request, err := requestBuilder.Build()
	if err != nil {
		log.Error("发生错误：", err)
		return nil
	}

	return request

	//log.Info("请求：\n", request)
	//tx, err := s.s.Request(request)
	//if err != nil {
	//	panic(err)
	//}
	//resp := getResponse(tx)
	//log.Infof("收到invite响应：\n%s", resp)
	//log.Infof("\ntx key: %s", tx.Key().String())
	//
	//ackRequest := sip.NewAckRequest("", request, resp, "", nil)
	//ackRequest.SetRecipient(request.Recipient())
	//ackRequest.AppendHeader(&sip.ContactHeader{
	//	Address: request.Recipient(),
	//	Params:  nil,
	//})
	//SipSender.TransmitRequest(ackRequest, nil)

}

// 从自身SIP服务获取地址返回FromHeader
func newFromAddress(params sip.Params) *sip.Address {
	log.Info(config.SIPUser())
	return &sip.Address{
		Uri: &sip.SipUri{
			FUser: sip.String{Str: config.SIPUser()},
			FHost: config.SIPDomain(),
		},
		Params: params,
	}
}

func newTo(user, host string) *sip.Address {
	return &sip.Address{
		Uri: &sip.SipUri{
			FUser: sip.String{Str: user},
			FHost: host,
		},
	}
}

func newParams(m map[string]string) sip.Params {
	params := sip.NewParams()
	for k, v := range m {
		params.Add(k, sip.String{Str: v})
	}
	return params
}

func newVia(transport string) *sip.ViaHop {
	port, err := strconv.ParseInt(config.SIPPort(), 10, 64)
	if err != nil {
		log.Error("解析Via头部端口失败", err)
	}
	p := sip.Port(port)

	params := newParams(map[string]string{
		"branch": sip.GenerateBranch(),
	})

	return &sip.ViaHop{
		ProtocolName:    "SIP",
		ProtocolVersion: "2.0",
		Transport:       transport,
		Host:            config.SIPAddress(),
		Port:            &p,
		Params:          params,
	}
}

func randString(n int) string {
	rand.Seed(time.Now().UnixNano())
	output := make([]byte, n)
	randomness := make([]byte, n)

	_, err := rand.Read(randomness)
	if err != nil {
		panic(err)
	}
	l := len(letterBytes)

	for pos := range output {
		random := randomness[pos]
		randomPos := random % uint8(l)
		output[pos] = letterBytes[randomPos]
	}

	return string(output)
}

func getResponse(tx sip.ClientTransaction) sip.Response {
	for {
		resp := <-tx.Responses()
		if resp.StatusCode() == sip.StatusCode(http.StatusContinue) ||
			resp.StatusCode() == sip.StatusCode(http.StatusSwitchingProtocols) {
			continue
		}
		return resp
	}
}
