/*
Copyright (c) 2021 Intel Corporation

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package main

import (
	eiicfgmgr "github.com/open-edge-insights/eii-configmgr-go/eiiconfigmgr"
	eiimsgbus "github.com/open-edge-insights/eii-messagebus-go/eiimsgbus"

	util "restdataexport/util"
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang/glog"
)

type restExport struct {
	rdeCaCertPool  *x509.CertPool
	extCaCertPool  *x509.CertPool
	clientCert     tls.Certificate
	rdeConfig      map[string]interface{}
	imgStoreConfig map[string]interface{}
	service        *eiimsgbus.ServiceRequester
	host           string
	port           string
	devMode        bool
	metadata       map[string]interface{}
}

const (
	rdeCertPath = "/opt/intel/eii/rde_server_cert.der"
	rdeKeyPath  = "/opt/intel/eii/rde_server_key.der"
)

// init is used to initialize and fetch required config
func (r *restExport) init() {

	flag.Parse()
	flag.Set("logtostderr", "true")
	defer glog.Flush()

	confHandler, err := eiicfgmgr.ConfigManager()
	if err != nil {
		glog.Fatal("Config Manager initialization failed...")
	}
	defer confHandler.Destroy()

	flag.Set("stderrthreshold", os.Getenv("GO_LOG_LEVEL"))
	flag.Set("v", os.Getenv("GO_VERBOSE"))

	// Setting devMode
	r.devMode, err = confHandler.IsDevMode()
	if err != nil {
		glog.Errorf("string to bool conversion error")
		os.Exit(1)
	}

	// Fetching required etcd config
	value, err := confHandler.GetAppConfig()
	if err != nil {
		glog.Errorf("Error while fetching config : %s\n", err.Error())
		os.Exit(1)
	}
	r.rdeConfig = value

	// Reading schema json
	schema, err := ioutil.ReadFile("./schema.json")
	if err != nil {
		glog.Errorf("Schema file not found")
		os.Exit(1)
	}

	//Convert the App config to a string
	b, err := json.Marshal(value)
	if err != nil {
		glog.Errorf("json Marshalling failed for config")
		os.Exit(1)
	}
	StringifiedValue := string(b)
	// Validating config json
	if util.ValidateJSON(string(schema), StringifiedValue) != true {
		glog.Errorf("Error while validating JSON\n")
		os.Exit(1)
	}

	r.host = value["rest_export_server_host"].(string)
	r.port = value["rest_export_server_port"].(string)

	numOfClients, _ := confHandler.GetNumClients()
	if numOfClients != -1 {
		// Fetching ImageStore config
		clntContext, err := confHandler.GetClientByIndex(0)
		if err != nil {
			glog.Errorf("Failed to get client context: %v", err)
			return
		}
		defer clntContext.Destroy()

		imgStoreConfig, err := clntContext.GetMsgbusConfig()
		if err != nil {
			glog.Errorf("Failed to fetch msgbus config : %v", err)
			return
		}

		appInterface, err := clntContext.GetInterfaceValue("Name")
		if err != nil {
			glog.Errorf("Failed to fetch interface value for Name: %v", err)
			return
		}

		serviceName, err := appInterface.GetString()
		if err != nil {
			glog.Errorf("Error to GetString value %v\n", err)
		}

		r.imgStoreConfig = imgStoreConfig

		client, err := eiimsgbus.NewMsgbusClient(r.imgStoreConfig)
		if err != nil {
			glog.Errorf("-- Error initializing message bus context: %v\n", err)
		}
		service, err := client.GetService(serviceName)
		if err != nil {
			glog.Errorf("-- Error initializing service requester: %v\n", err)
		}
		r.service = service
	} else {
		glog.Infof("No client instances found, not initializing ImageStore...")
	}

	// Getting required certs from etcd
	if !r.devMode {
		rdeCerts := []string{rdeCertPath, rdeKeyPath}
		rdeExportKeys := []string{"server_cert", "server_key"}

		i := 0
		for _, rdeExportKey := range rdeExportKeys {
			rdeCertFile, _ := value[rdeExportKey].(string)
			certFile := []byte(rdeCertFile)
			err = ioutil.WriteFile(rdeCerts[i], certFile, 0400)
			i++
		}

		// Fetching and storing required CA certs
		serverCa := value["http_server_ca"].(string)
		caCert := []byte(serverCa)

		rdeCaFile, _ := value["ca_cert"].(string)
		caFile := []byte(rdeCaFile)

		// Adding Rest Data Export and server CA to certificate pool
		extCaCertPool := x509.NewCertPool()
		extCaCertPool.AppendCertsFromPEM(caCert)

		rdeCaCertPool := x509.NewCertPool()
		rdeCaCertPool.AppendCertsFromPEM(caFile)

		r.rdeCaCertPool = rdeCaCertPool
		r.extCaCertPool = extCaCertPool

		// Read the key pair to create certificate struct
		certFile, _ := value["server_cert"].(string)
		rdeCertFile := []byte(certFile)

		keyFile, _ := value["server_key"].(string)
		rdeKeyFile := []byte(keyFile)

		cert, err := tls.X509KeyPair(rdeCertFile, rdeKeyFile)
		if err != nil {
			glog.Errorf("Error : %s", err)
		}
		r.clientCert = cert
	}

	numOfSubscriber, _ := confHandler.GetNumSubscribers()
	for i := 0; i < numOfSubscriber; i++ {
		subctx, err := confHandler.GetSubscriberByIndex(i)
		if err != nil {
			glog.Errorf("Failed to get subscriber context: %v", err)
			return
		}

		subTopics, err := subctx.GetTopics()
		if err != nil {
			glog.Errorf("Failed to fetch topics : %v", err)
			return
		}

		config, err := subctx.GetMsgbusConfig()
		if err != nil {
			glog.Errorf("-- Error getting message bug config: %v\n", err)
		}
		go r.startEiiSubscriber(config, subTopics[0])
		subctx.Destroy()
	}

}

// startEiiSubscriber is used to start EIIMbus subscribers over specified topic
func (r *restExport) startEiiSubscriber(config map[string]interface{}, topic string) {

	client, err := eiimsgbus.NewMsgbusClient(config)
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
			r.metadata = msg.Data
			if os.Getenv("HTTP_METHOD_FETCH_METADATA") == "POST" {
				r.postMetaData(msg.Data, topic)
			}
		case err := <-subscriber.ErrorChannel:
			glog.Errorf("Error receiving message: %v\n", err)
		}
	}
}

// postMetaData is used to send metadata via POST requests to external server
func (r *restExport) postMetaData(metadata map[string]interface{}, topic string) {

	// Adding meta-data to http request
	requestBody, err := json.Marshal(metadata)
	if err != nil {
		glog.Errorf("Error marshalling json : %s", err)
	}

	// Timeout for every request
	timeout := time.Duration(60 * time.Second)

	// Getting endpoint of server
	endpoint := fmt.Sprintf("%v", r.rdeConfig[topic])
	dialEndpoint := strings.Replace(endpoint, "http://", "", 1)
	if !r.devMode {
		// Replace http with https for PROD mode
		endpoint = strings.Replace(endpoint, "http", "https", 1)
	}

	// Check if HttpServer is running
	serverPresent := false
	for !serverPresent {
		timeout := 1 * time.Second
		conn, err := net.DialTimeout("tcp", dialEndpoint, timeout)
		if err != nil {
			glog.Errorf("HTTP Server not found, retrying...")
			time.Sleep(timeout)
		} else {
			serverPresent = true
		}
		if conn != nil {
			conn.Close()
		}
	}

	if r.devMode {

		client := &http.Client{
			Timeout: timeout,
		}

		// Making a post request to external server
		r, err := client.Post(endpoint+"/metadata", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			glog.Errorf("Remote HTTP server is not responding : %s", err)
			return
		}

		if r != nil {
			// Read the response body
			defer r.Body.Close()
			response, err := ioutil.ReadAll(r.Body)
			if err != nil {
				glog.Errorf("Failed to receive response from server : %s", err)
			}
			glog.Infof("Response : %s", string(response))
		}
	} else {

		// Create a HTTPS client and supply the created CA pool and certificate
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs:      r.extCaCertPool,
					Certificates: []tls.Certificate{r.clientCert},
				},
			},
			Timeout: timeout,
		}

		// Making a post request to external server
		r, err := client.Post(endpoint+"/metadata", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			glog.Errorf("Remote HTTP server is not responding : %s", err)
			return
		}

		if r != nil {
			// Read the response body
			defer r.Body.Close()
			response, err := ioutil.ReadAll(r.Body)
			if err != nil {
				glog.Errorf("Failed to receive response from server : %s", err)
			}
			glog.Infof("Response : %s", string(response))
		}
	}
}

// restExportServer starts a http server to serve GET requests
func (r *restExport) restExportServer() {

	http.HandleFunc("/metadata", r.getMetaData)
	http.HandleFunc("/image", r.getImage)

	if r.devMode {
		err := http.ListenAndServe(r.host+":"+r.port, nil)
		if err != nil {
			glog.Errorf("%v", err)
			os.Exit(-1)
		}
	} else {
		// Create a Server instance to listen on port with the TLS config
		server := &http.Server{
			Addr:              r.host + ":" + r.port,
			ReadTimeout:       60 * time.Second,
			ReadHeaderTimeout: 60 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20,
		}

		// Listen to HTTPS connections with the server certificate and wait
		err := server.ListenAndServeTLS(rdeCertPath, rdeKeyPath)
		if err != nil {
			glog.Errorf("%v", err)
			os.Exit(-1)
		}
	}
}

// getMetaData is used to fetch the subscrubed metadata
func (r *restExport) getMetaData(w http.ResponseWriter, re *http.Request) {
	w.Header().Set("Content-type", "text/json")
	responseData, err := json.Marshal(r.metadata)
	if err != nil {
		glog.Errorf("Error marshalling json : %s", err)
	}
	switch re.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	default:
		glog.Infof("Only GET supported")
	}
}

// readImage is used to fetch required image from ImageStore
func (r *restExport) readImage(imgHandle string) []byte {

	// Send Read command & get the frame data
	response := map[string]interface{}{"command": "read", "img_handle": imgHandle}
	err1 := r.service.Request(response)
	if err1 != nil {
		glog.Errorf("-- Error sending request: %v\n", err1)
		return nil
	}

	resp, err := r.service.ReceiveResponse(-1)
	if err != nil {
		glog.Errorf("-- Error receiving response: %v\n", err)
		return nil
	}

	return resp.Blob[0]
}

// getImage publishes image frame via GET request to external server
func (r *restExport) getImage(w http.ResponseWriter, re *http.Request) {
	// Setting content type for encoding
	w.Header().Set("Content-type", "image/jpeg; charset=utf-8")

	switch re.Method {
	case "GET":
		w.WriteHeader(http.StatusOK)
		reqEndpoint := strings.Split(re.URL.RawQuery, "=")
		img := make(map[string]string)
		img[reqEndpoint[0]] = reqEndpoint[1]
		// Send imgHandle to read from ImageStore
		imgHandle := img["img_handle"]
		frame := r.readImage(imgHandle)
		glog.Infof("Imghandle %s and md5sum %v", imgHandle, md5.Sum(frame))
		w.Write(frame)
	case "POST":
		fmt.Fprintf(w, "Received a POST request")
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func main() {

	// initializing constructor
	r := new(restExport)
	r.init()

	// start the Rest Export server to serve images via GET requests
	go r.restExportServer()

	select {}
}
