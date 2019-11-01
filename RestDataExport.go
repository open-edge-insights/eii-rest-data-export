package main

import (
	eismsgbus "EISMessageBus/eismsgbus"
	configmgr "IEdgeInsights/common/libs/ConfigManager"
	util "IEdgeInsights/common/util"
	msgbusutil "IEdgeInsights/common/util/msgbusutil"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
)

const (
	rdeCaPath   = "/opt/intel/eis/ca_cert.der"
	rdeCertPath = "/opt/intel/eis/rde_server_cert.der"
	rdeKeyPath  = "/opt/intel/eis/rde_server_key.der"
)

func startEisSubscriber(config map[string]interface{}, topic string, rdeConfig map[string]interface{}, devMode bool) {

	client, err := eismsgbus.NewMsgbusClient(config)
	if err != nil {
		glog.Errorf("-- Error initializing message bus context: %v\n", err)
		return
	}
	defer client.Close()
	subscriber, err := client.NewSubscriber(topic)
	if err != nil {
		glog.Errorf("-- Error subscribing to topic: %v\n", err)
		return
	}
	defer subscriber.Close()

	for {
		select {
		case msg := <-subscriber.MessageChannel:
			glog.V(1).Infof("-- Received Message --")
			// Adding topic to meta-data for easy differentitation in external server
			msg.Data["topic"] = topic
			publishMetaData(msg.Data, topic, rdeConfig, devMode)
		case err := <-subscriber.ErrorChannel:
			glog.Errorf("Error receiving message: %v\n", err)
		}
	}
}

func publishMetaData(metadata map[string]interface{}, topic string, rdeConfig map[string]interface{}, devMode bool) {

	// Adding meta-data to http request
	requestBody, err := json.Marshal(metadata)
	if err != nil {
		glog.Errorf("Error marshalling json : %s", err)
	}

	// Timeout for every request
	timeout := time.Duration(10 * time.Second)

	if devMode {

		client := &http.Client{
			Timeout: timeout,
		}

		// Getting endpoint of server
		endpoint := fmt.Sprintf("%v", rdeConfig[topic])

		// Making a post request to external server
		r, err := client.Post(endpoint+"/metadata", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			glog.Errorf("Remote HTTP server is not responding : %s", err)
		}

		// Read the response body
		defer r.Body.Close()
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Errorf("Failed to receive response from server : %s", err)
		}

		glog.Infof("Response : %s", string(response))

	} else {

		// Getting endpoint and ca of server
		endpoint := fmt.Sprintf("%v", rdeConfig[topic])
		serverCaPath := fmt.Sprintf("%v", rdeConfig["http_server_ca"])

		// Read the key pair to create certificate
		cert, err := tls.LoadX509KeyPair(rdeCertPath, rdeKeyPath)
		if err != nil {
			glog.Errorf("Error : %s", err)
		}

		// Adding server CA to certificate pool
		caCert, err := ioutil.ReadFile(serverCaPath)
		if err != nil {
			glog.Errorf("Error : %s", err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Create a CA certificate pool
		rdeCaCert, err := ioutil.ReadFile(rdeCaPath)
		if err != nil {
			glog.Errorf("Error : %s", err)
		}

		rdeCaCertPool := x509.NewCertPool()
		rdeCaCertPool.AppendCertsFromPEM(rdeCaCert)

		// Create a HTTPS client and supply the created CA pool and certificate
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      caCertPool,
					ClientCAs:    rdeCaCertPool,
					Certificates: []tls.Certificate{cert},
				},
			},
			Timeout: timeout,
		}

		// Replace http with https for PROD mode
		endpoint = strings.Replace(endpoint, "http", "https", 1)
		// Making a post request to external server
		r, err := client.Post(endpoint+"/metadata", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			glog.Errorf("Remote HTTP server is not responding : %s", err)
		}

		// Read the response body
		defer r.Body.Close()
		response, err := ioutil.ReadAll(r.Body)
		if err != nil {
			glog.Errorf("Failed to receive response from server : %s", err)
		}

		glog.Infof("Response : %s", string(response))

	}
}

// // restExportServer starts a http server to serve GET requests
// func restExportServer() {
// 	http.HandleFunc("/", getImage)
//  http.ListenAndServe(":8080", nil)
// }

// TODO : Implement getImage API
// // getImage publishes image frame
// func getImage(w http.ResponseWriter, r *http.Request) {
// }

func main() {

	flag.Parse()
	flag.Set("logtostderr", "true")
	defer glog.Flush()

	appName := os.Getenv("AppName")
	config := util.GetCryptoMap(appName)
	confHandler := configmgr.Init("etcd", config)

	flag.Set("stderrthreshold", os.Getenv("GO_LOG_LEVEL"))
	flag.Set("v", os.Getenv("GO_VERBOSE"))

	devMode, err := strconv.ParseBool(os.Getenv("DEV_MODE"))
	if err != nil {
		glog.Errorf("string to bool conversion error")
		os.Exit(1)
	}

	// Fetching required etcd config
	value, err := confHandler.GetConfig("/" + appName + "/config")
	if err != nil {
		glog.Infof("Error while fetching config : %s\n", err.Error())
		os.Exit(1)
	}

	s := strings.NewReader(value)
	var rdeConfig map[string]interface{}
	err = json.NewDecoder(s).Decode(&rdeConfig)
	if err != nil {
		glog.Infof("Error while decoding JSON : %s\n", err.Error())
		os.Exit(1)
	}

	// Getting required certs from etcd
	if !devMode {

		rdeCerts := []string{rdeCertPath, rdeKeyPath, rdeCaPath}
		rdeExportKeys := []string{"/RestDataExport/server_cert", "/RestDataExport/server_key", "/RestDataExport/ca_cert"}

		i := 0
		for _, rdeExportKey := range rdeExportKeys {
			rdeCertFile, err := confHandler.GetConfig(rdeExportKey)
			if err != nil {
				glog.Errorf("Error : %s", err)
			}
			certFile := []byte(rdeCertFile)
			err = ioutil.WriteFile(rdeCerts[i], certFile, 0644)
			i++
		}
	}

	// Starting EISMbus subcribers
	var subTopics []string
	subTopics = msgbusutil.GetTopics("SUB")
	for _, subTopicCfg := range subTopics {
		msgBusConfig := msgbusutil.GetMessageBusConfig(subTopicCfg, "SUB", devMode, config)
		subTopicCfg := strings.Split(subTopicCfg, "/")
		go startEisSubscriber(msgBusConfig, subTopicCfg[1], rdeConfig, devMode)
	}
	select {}
}
