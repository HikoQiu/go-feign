package feign

import (
    "github.com/valyala/fasthttp"
    "time"
)

type LbcResp struct {
    Request *LbcReq
    Response *fasthttp.Response
    ReceiveAt  time.Time
}
