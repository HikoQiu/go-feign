package main

import (
    "github.com/olivere/balancers/roundrobin"
    "github.com/olivere/balancers"
    "log"
)

func main() {
    // Get a balancer that performs round-robin scheduling between two servers.
    balancer, err := roundrobin.NewBalancerFromURL("https://server1.com", "https://server2.com")
    if err != nil {
        log.Println(err.Error())
        return
    }

    // Get a HTTP client based on that balancer.
    client := balancers.NewClient(balancer)

    // Now request some data. The scheme, host, and user info will be rewritten
    // by the balancer; you'll never get data from http://example.com, only data
    // from http://server1.com or http://server2.com.
    client.Get("http://example.com/path1?foo=bar") // rewritten to https://server1.com/path1?foo=bar
    client.Get("http://example.com/path1?foo=bar") // rewritten to https://server2.com/path1?foo=bar
    client.Get("http://example.com/path1?foo=bar") // rewritten to https://server1.com/path1?foo=bar
    client.Get("/path1?foo=bar")                   // rewritten to https://server2.com/path1?foo=bar
}

