package feign

import (
    "github.com/go-resty/resty"
    "sync/atomic"
    "sync"
    "time"
)

var onceIndex = sync.Once{}
var appNextUrlIndex = make(map[string]*uint32)

// Feign load balance client
type Lbc struct {
    feign    *Feign
    app      string
    urlIndex uint32

    // resty.Client
    client *resty.Client
}

// pick a server to send request
func (t *Lbc) pickServer() string {
    onceIndex.Do(func() {
        if appNextUrlIndex[t.app] == nil {
            v := uint32(time.Now().UnixNano())
            appNextUrlIndex[t.app] = &v
        }
    })

    urls, ok := t.feign.GetAppUrls(t.app)
    if !ok {
        log.Errorf("Failed to pick server, reason: no available urls for app=%s", t.app)
        return ""
    }
    idx := atomic.AddUint32(appNextUrlIndex[t.app], 1)
    idx %= uint32(len(urls))
    t.urlIndex = idx
    // @TODO reset index
    //atomic.CompareAndSwapUint32(appNextUrlIndex[t.app], *(appNextUrlIndex[t.app]), uint32(len(urls)))

    return urls[idx]
}


func (t *Lbc) pick() *Lbc {
    t.client = resty.New()
    t.client.HostURL = t.pickServer()
    log.Debugf("Picked url=%s", t.client.HostURL)
    return t
}

