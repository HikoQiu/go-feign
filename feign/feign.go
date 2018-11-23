package feign

import (
    "github.com/HikoQiu/go-eureka-client/eureka"
    "github.com/valyala/fasthttp"
    "strings"
    "fmt"
    "sync"
)

var DefaultFeign = &Feign{
    appLBClient:     make(map[string]*Lbc),
    discoveryClient: eureka.DefaultClient,
    appUrls:         make(map[string][]string),
}

type Feign struct {
    // app => balance instance for sending http requests
    appLBClient map[string]*Lbc

    // Discovery client to get Apps and instances
    discoveryClient eureka.DiscoveryClient

    // assign app => urls
    appUrls map[string][]string

    mu sync.RWMutex
}

func (t *Feign) Config() *Feign {
    return t
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
            u = strings.TrimRight(u, "/")
            c := &fasthttp.HostClient{
                Addr: u,
            }
            lbc.Clients = append(lbc.Clients, c)
            t.appUrls[app] = append(t.appUrls[app], u)
        }
        t.appLBClient[app] = NewLbc(&lbc)
    }

    return t
}

// 返回负载均衡的客户端
func (t *Feign) App(app string) *Lbc {
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("catch panic err=%v", r)
        }
    }()

    t.tryUpdateLBClients(app)
    return t.appLBClient[app]
}

func (t *Feign) tryUpdateLBClients(app string) {
    if t.discoveryClient == nil ||
        len(t.discoveryClient.GetRegistryApps()) == 0 {
        log.Debugf("no discovery client, no need to update LBClient")
        return
    }

    registryApps := t.discoveryClient.GetRegistryApps()
    if _, ok := registryApps[app]; !ok || len(registryApps[app].Instances) == 0 {
        return
    }

    tmpAppUrls := map[string][]string{}
    for keyApp, appVo := range registryApps {
        // if app is exist in t.appUrls, check whether app's urls are updated
        // if app's urls are updated, then reset LBClient
        if _, ok := t.appUrls[keyApp]; ok {
            isUpdate := false
            for _, insVo := range appVo.Instances {
                isExist := false
                for _, v := range t.appUrls[keyApp] {
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

            fmt.Sprintf("%v", isUpdate)
            tmpAppUrls[keyApp] = make([]string, 0)
            for _, insVo := range appVo.Instances {
                tmpAppUrls[keyApp] = append(tmpAppUrls[app], strings.TrimRight(insVo.HomePageUrl, "/"))
            }
            // @TODO Update LBClient

            continue
        }

        // app are not exist in t.appUrls
        for _, insVo := range appVo.Instances {
            tmpAppUrls[keyApp] = append(tmpAppUrls[app], strings.TrimRight(insVo.HomePageUrl, "/"))
        }
        // @TODO Update LBClient
    }

    t.UseUrls(tmpAppUrls)
    log.Debugf("TmpAppUrls, app urls=%v ", tmpAppUrls)
}

//feign.App("APP_ID_CLIENT_FROM_DNS").Put()
//feign.App("APP_ID_CLIENT_FROM_DNS").Get()
//feign.App("APP_ID_CLIENT_FROM_DNS").Post()
