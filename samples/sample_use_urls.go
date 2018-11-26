package main

import (
    "github.com/HikoQiu/go-eureka-client/eureka"
    "github.com/HikoQiu/go-feign/feign"
    "log"
)

func runEurekaClient() {
    config := eureka.GetDefaultEurekaClientConfig()
    config.UseDnsForFetchingServiceUrls = true
    config.Region = "region-cn-hd-1"
    config.AvailabilityZones = map[string]string{
        "region-cn-hd-1": "zone-cn-hz-1",
    }
    config.EurekaServerDNSName = "dev.ms-registry.xf.io"
    config.EurekaServerUrlContext = "eureka"
    config.EurekaServerPort = "9001"

    // run eureka client async
    eureka.DefaultClient.Config(config).
        Register("APP_ID_CLIENT_FROM_DNS", 9000).
        Run()
}

func main() {
    runEurekaClient()

    // 1.1 configure maps, app => app's urls
    feign.DefaultFeign.UseUrls(map[string][]string{
        "MS-REGISTRY-URLS": {"http://192.168.20.236:9001", "http://192.168.20.237:9001"},
    })

    // 2.1 Use app name to load balance request all instances
    res, err := feign.DefaultFeign.App("MS-REGISTRY-URLS").R().SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/MS-REGISTRY")
    if err != nil {
        log.Println("err: ", err.Error())
        return
    }
    log.Println("res=", string(res.Body()))

    // load balance request to another url.
    res, err = feign.DefaultFeign.App("MS-REGISTRY-URLS").R().SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/MS-REGISTRY")
    if err != nil {
        log.Println("err: ", err.Error())
        return
    }
    log.Println("res=", string(res.Body()))

    select{}
}
