package main

import (
    "github.com/HikoQiu/go-feign/feign"
    "log"
)


func main() {
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

    // test: request to another url balancely.
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
