package feign

import (
    "github.com/HikoQiu/go-eureka-client/eureka"
    "github.com/valyala/fasthttp"
    "strings"
    "fmt"
    "sync"
    "net/url"
    "github.com/go-resty/resty"
    "time"
)

var DefaultFeign = &Feign{
    discoveryClient: eureka.DefaultClient,
    appUrls:         make(map[string][]string),
}

type Feign struct {
    // Discovery client to get Apps and instances
    discoveryClient eureka.DiscoveryClient

    // assign app => urls
    appUrls map[string][]string

    // ensure some daemon task only run one time
    once sync.Once

    mu sync.RWMutex
}

// use discovery client to get all registry app => instances
func (t *Feign) UseDiscoveryClient(client eureka.DiscoveryClient) *Feign {
    t.discoveryClient = client
    return t
}

// assign static app => urls
func (t *Feign) UseUrls(appUrls map[string][]string) *Feign {
    t.mu.Lock()
    defer t.mu.Unlock()

    for app, urls := range appUrls {
        var lbc fasthttp.LBClient
        if t.appUrls[app] == nil {
            t.appUrls[app] = make([]string, 0)
        }

        for _, u := range urls {
            tmpU, err := url.Parse(u)
            if err != nil {
                log.Errorf("Invalid url=%s, parse err=%s", u, err.Error())
                continue
            }

            c := &fasthttp.HostClient{
                Addr: fmt.Sprintf("%s:%s", tmpU.Hostname(), tmpU.Port()),
            }
            lbc.Clients = append(lbc.Clients, c)
            t.appUrls[app] = append(t.appUrls[app], u)
        }
    }

    return t
}

// return resty.Client
func (t *Feign) App(app string) *resty.Client {
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("App(%s) catch panic err=%v", app, r)
        }
    }()

    // daemon to update app urls periodically
    // only execute once globally
    t.once.Do(func() {
        if t.discoveryClient == nil ||
            len(t.discoveryClient.GetRegistryApps()) == 0 {
            log.Debugf("no discovery client, no need to update LBClient")
            return
        }

        t.updateAppUrlsIntervals()
    })

    // try update app's urls
    // if app's urls is exist, do nothing
    t.tryRefreshAppUrls(app)

    lbc := &Lbc{
        feign: t,
        app:   app,
    }
    return lbc.pick().client
}

// try update app's urls
// if app's urls is exist, do nothing
func (t *Feign) tryRefreshAppUrls(app string) {
    if _, ok := t.GetAppUrls(app); ok {
        return
    }

    if t.discoveryClient == nil ||
        len(t.discoveryClient.GetRegistryApps()) == 0 {
        log.Debugf("no discovery client, no need to update LBClient")
        return
    }

    t.updateAppUrls()
}

// update app urls periodically
func (t *Feign) updateAppUrlsIntervals() {
    go func() {
        for {
            t.updateAppUrls()

            time.Sleep(time.Second * 60)
            log.Debugf("Update app urls interval...ok")
        }
    }()
}

// Update app urls from registry apps
func (t *Feign) updateAppUrls() {
    registryApps := t.discoveryClient.GetRegistryApps()
    tmpAppUrls := map[string][]string{}
    for app, appVo := range registryApps {
        // if app is exist in t.appUrls, check whether app's urls are updated
        // if app's urls are updated, then reset LBClient
        if curAppUrls, ok := t.GetAppUrls(app); ok {
            isUpdate := false
            for _, insVo := range appVo.Instances {
                isExist := false

                for _, v := range curAppUrls {
                    insHomePageUrl := strings.TrimRight(insVo.HomePageUrl, "/")
                    if v == insHomePageUrl {
                        isExist = true
                        break
                    }
                }

                if !isExist {
                    isUpdate = true
                    break
                }
            }

            if isUpdate {
                tmpAppUrls[app] = make([]string, 0)
                for _, insVo := range appVo.Instances {
                    tmpAppUrls[app] = append(tmpAppUrls[app], strings.TrimRight(insVo.HomePageUrl, "/"))
                }
            }
        } else {
            // app are not exist in t.appUrls
            for _, insVo := range appVo.Instances {
                tmpAppUrls[app] = append(tmpAppUrls[app], strings.TrimRight(insVo.HomePageUrl, "/"))
            }
        }
    }

    t.UseUrls(tmpAppUrls)
}

// get app's urls
func (t *Feign) GetAppUrls(app string) ([]string, bool) {
    t.mu.RLock()
    defer t.mu.RUnlock()

    if _, ok := t.appUrls[app]; !ok {
        return nil, false
    }

    return t.appUrls[app], true
}
