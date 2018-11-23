package feign

import (
    "github.com/valyala/fasthttp"
    "net/url"
    "net/http"
    "time"
)

// Feign load balance client
type Lbc struct {
    client *fasthttp.LBClient

    AllowGetMethodPayload bool

    // handlers before request
    beforeRequest []func(*Lbc, *LbcReq) error
    // Hook before request
    preReqHook func(*Lbc, *LbcReq) error
    // handlers after request
    afterResponse []func(*Lbc, *LbcResp) error
}

func NewLbc(client *fasthttp.LBClient) *Lbc {
    c := &Lbc{
        client: client,
    }

    // default before request middleware
    c.beforeRequest = []func(*Lbc, *LbcReq) error{
        parseRequestURL,
        parseRequestHeader,
        parseRequestBody,
        createHTTPRequest,
    }

    // default after response middleware
    c.afterResponse = []func(*Lbc, *LbcResp) error{
        parseResponseBody,
    }

    return c
}

func (t *Lbc) R() *LbcReq {
    return &LbcReq{
        lbc:        t,
        URL:        "",
        Method:     "",
        QueryParam: url.Values{},
        FormData:   url.Values{},
        Header:     http.Header{},
        Body:       nil,
        Result:     nil,
        Error:      nil,
        pathParams: make(map[string]string),
    }
}

func (t *Lbc) do(r *LbcReq) (*LbcResp, error) {
    defer releaseBuffer(r.bodyBuf)
    var err error

    // call before middleware
    for _, f := range t.beforeRequest {
        if err = f(t, r); err != nil {
            log.Debugf("BeforeRequest handler failed, err=%s", err.Error())
            return nil, err
        }
    }

    // call pre-request if defined
    if t.preReqHook != nil {
        if err = t.preReqHook(t, r); err != nil {
            log.Debugf("PreReqHook failed, err=%s", err.Error())
            return nil, err
        }
    }

    // http request by fasthttp LBClient
    req := &fasthttp.Request{}
    resp := &fasthttp.Response{}

    for k, v := range r.Header {
        req.Header.Set(k, v[0])
    }
    req.Header.SetMethod(r.Method)
    req.URI().Update(r.URL)

    err = t.client.Do(req, resp)
    if err != nil {
        log.Errorf("Failed to request, err=%s", err.Error())
    }

    lbcResp := &LbcResp{
        Request:   r,
        Response:  resp,
        ReceiveAt: time.Now(),
    }

    return lbcResp, nil
}
