package main

import (
	"flag"
	"fmt"
	"github.com/heyuhang0/DSCProject/pkg/kvclient"
	"log"
	"net/http"
)

var client *kvclient.KeyValueStoreClient

func get(w http.ResponseWriter, req *http.Request) {
	// convert http request to grpc request
	fmt.Fprintf(w, "hello\n")
}

func put(w http.ResponseWriter, req *http.Request) {

	for name, headers := range req.Header {
		for _, h := range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}

func main() {

	http.HandleFunc("/get", get)
	http.HandleFunc("/put", put)

	// initialize client
	configPath := flag.String("config", "./configs/default_client.json", "config path")
	flag.Parse()

	config, err := kvclient.NewClientConfigFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	client = kvclient.NewKeyValueStoreClient(config)

	http.ListenAndServe(":8090", nil)
}