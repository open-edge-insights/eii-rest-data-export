# `RestDataExport`

RestDataExport service subscribes to any topic from EISMessageBus and starts publishing meta data via POST requests to any external HTTP servers. It has an internal HTTP server running to respond to any GET requests for a required frame from any HTTP clients.


## `Configuration`

For more details on Etcd and MessageBus endpoint configuration, visit [Etcd_Secrets_and_MsgBus_Endpoint_Configuration](../Etcd_Secrets_and_MsgBus_Endpoint_Configuration.md).

## `Pre-requisites`

        1. Make sure TestServer application is running by following [README.md](../tools/HttpTestServer/README.md)

        2. Make sure ImageStore application is running by following [README.md](../ImageStore/README.md)

        3. Make sure the topics you subscribe to are also added in the [config](config.json) with HttpServer endpoint specified
           Eg: If you are adding a new subscription topic 'dc_point_data_results', the new config will be
           ```
                {
                    "camera1_stream_results": "http://localhost:8082",
                    "point_classifier_results": "http://localhost:8082",
                    "dc_point_data_results": "http://localhost:8082",
                    "http_server_ca": "/opt/intel/eis/cert.pem",
                    "rest_export_server_host": "localhost",
                    "rest_export_server_port": "8087"
                }
            ```

## `Installation`

* Follow steps 1-5 of main [EIS README](../README.md) if not done already as part of EIS stack setup

> **NOTE:** For running in PROD mode, please copy the required certs of TestServer [cert.pem](../tools/HttpTestServer/cert.pem) and [key.pem](../tools/HttpTestServer/key.pem) to /opt/intel/eis/ directory.

## `Running in CSL or Kubernetes setup`

* Update the config.json file with the following settings:

  ```
    {
        "camera1_stream_results": "http://IP Address of Test Server:8082",
        "point_classifier_results": "http://IP Address of Test Server:8082",
        "http_server_ca": "/opt/intel/eis/cert.pem",
        "rest_export_server_host": "0.0.0.0",
        "rest_export_server_port": "8087"
    }
  ```
