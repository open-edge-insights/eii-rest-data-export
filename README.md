# `RestDataExport`

RestDataExport service subscribes to any topic from EISMessageBus and starts publishing meta data via POST requests to any external HTTP servers. It has an internal HTTP server running to respond to any GET requests for a required frame from any HTTP clients.


## `Configuration`

For more details on Etcd and MessageBus endpoint configuration, visit [Etcd_and_MsgBus_Endpoint_Configuration](../Etcd_and_MsgBus_Endpoint_Configuration.md).

## `Pre-requisites`

        1. Make sure TestServer application is running by following [README.md](../tools/HttpTestServer/README.md)

        2. Make sure ImageStore application is running by following [README.md](../ImageStore/README.md)

## `Installation`

* Follow [provision/README.md](../README#provision-eis.md) for EIS provisioning
  if not done already as part of EIS stack setup

* Run RestDataExport

	1. Build and Run RestDataExport as container
        ```
        $ cd [repo]/docker_setup
        $ docker-compose up --build ia_rest_export
       ```

> **NOTE:** For running in PROD mode, please copy the required certs of TestServer [cert.pem](../tools/HttpTestServer/cert.pem) and [key.pem](../tools/HttpTestServer/key.pem) to /opt/intel/eis/ directory.
