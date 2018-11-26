package main

import (
    "github.com/HikoQiu/go-feign/feign"
    "log"
    "github.com/HikoQiu/go-eureka-client/eureka"
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

    res, err := feign.DefaultFeign.App("MS-REGISTRY").R().
        SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/WM")
    if err != nil {
        log.Println("err=", err.Error())
        return
    }

    log.Println("res=", string(res.Body()))

    select {}
}

