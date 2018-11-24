package feign

import (
    "testing"
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

    // custom logger
    //eureka.SetLogger(func(level int, format string, a ...interface{}) {
    //    if level == eureka.LevelError {
    //        fmt.Println("[custom logger error] " + format, a)
    //    }else {
    //        fmt.Println("[custom logger debug] " + format, a)
    //    }
    //})

    // run eureka client async
    eureka.DefaultClient.Config(config).
        Register("APP_ID_CLIENT_FROM_DNS", 9000).
        Run()
}

func Test_AppUrls(t *testing.T) {
    runEurekaClient()

    lbc := DefaultFeign.App("MS-REGISTRY").R()
    res, err := lbc.SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/WM")
    if err != nil {
        log.Errorf("err=%s", err.Error())
    }
    //log.Debugf("res=%v", res)

    res, err = DefaultFeign.App("MS-REGISTRY").R().SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/SVR-CHANNEL")
    if err != nil {
        log.Errorf("err=%s", err.Error())
    }

    //log.Debugf("res=%v", res)

    res, err = DefaultFeign.App("MS-REGISTRY").R().SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/SVR-CHANNEL")
    if err != nil {
        log.Errorf("err=%s", err.Error())
    }

    //log.Debugf("res=%v", res)

    res, err = DefaultFeign.App("MS-REGISTRY").R().SetHeaders(map[string]string{
        "Content-Type": "application/json",
    }).Get("/eureka/apps/SVR-CHANNEL")
    if err != nil {
        log.Errorf("err=%s", err.Error())
    }

    log.Debugf("res=%v", res)

}
