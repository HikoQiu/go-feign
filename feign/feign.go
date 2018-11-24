package feign

import (
    "github.com/HikoQiu/go-eureka-client/eureka"
    "strings"
    "net/url"
    "github.com/go-resty/resty"
    "time"
    "sync"
)

var DefaultFeign = &Feign{
    discoveryClient: eureka.DefaultClient,
    appUrls:         make(map[string][]string),
    appNextUrlIndex: make(map[string]*uint32),
}

type Feign struct {
    // Discovery client to get Apps and instances
    discoveryClient eureka.DiscoveryClient

    // assign app => urls
    appUrls map[string][]string

    // Counter for calculate next url'index
    appNextUrlIndex map[string]*uint32

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

    //v := uint32(time.Now().UnixNano())
    //appNextUrlIndex[t.app] = &v
    for app, urls := range appUrls {

        // reset app'urls
        tmpAppUrls := make([]string, 0)
        for _, u := range urls {
            _, err := url.Parse(u)
            if err != nil {
                log.Errorf("Invalid url=%s, parse err=%s", u, err.Error())
                continue
            }

            tmpAppUrls = append(tmpAppUrls, u)
        }

        if len(tmpAppUrls) == 0 {
            log.Errorf("Empty valid urls for app=%s, skip to set app's urls", app)
            continue
        }

        t.appUrls[app] = tmpAppUrls
        if t.appNextUrlIndex[app] == nil {
            v := uint32(time.Now().UnixNano())
            t.appNextUrlIndex[app] = &v
        }
    }

    return t
}

// return resty.Client
func (t *Feign) App(app string) *resty.Client {
    defer func() {
        if err := recover(); err != nil {
            log.Errorf("App(%s) catch panic err=%v", app, err)
        }
    }()

    // daemon to update app urls periodically
    // only execute once globally
    t.once.Do(func() {
        if t.discoveryClient == nil {
            log.Infof("no discovery client, no need to update appUrls periodically.")
            return
        }

        t.updateAppUrlsIntervals()
    })

    // try update app's urls.
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
        log.Debugf("no discovery client, no need to update app'urls.")
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
    tmpAppUrls := make(map[string][]string)

    for app, appVo := range registryApps {
        var isAppAlreadyExist bool
        var curAppUrls []string
        var isUpdate bool

        // if app is already exist in t.appUrls, check whether app's urls are updated.
        // if app's urls are updated, t.appUrls
        if curAppUrls, isAppAlreadyExist = t.GetAppUrls(app); isAppAlreadyExist {
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
        }

        // app are not exist in t.appUrls or app's urls has been update
        if !isAppAlreadyExist || isUpdate {
            tmpAppUrls[app] = make([]string, 0)
            for _, insVo := range appVo.Instances {
                tmpAppUrls[app] = append(tmpAppUrls[app], strings.TrimRight(insVo.HomePageUrl, "/"))
            }
        }
    }

    // update app's urls to feign
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
