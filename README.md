# `RestDataExport`

RestDataExport service subscribes to any topic from EISMessageBus and starts publishing meta data via POST requests to any external HTTP servers.


## `Configuration`

For more details on Etcd and MessageBus endpoint configuration, visit [Etcd_and_MsgBus_Endpoint_Configuration](../Etcd_and_MsgBus_Endpoint_Configuration.md).

## `Installation`

* Follow [provision/README.md](../README#provision-eis.md) for EIS provisioning
  if not done already as part of EIS stack setup

* Run RestDataExport

	1. Build and Run RestDataExport as container
        ```
        $ cd [repo]/docker_setup
        $ docker-compose up --build ia_rest_export
       ```

> **NOTE:** For running in PROD mode, please copy the required certs as [cert.pem](./test/cert.pem) and [key.pem](./test/key.pem) files to /opt/intel/eis/ directory.

* To Run test rest data export test application

        1. Make sure TestServer application is running
        ```
        $ cd [repo]/tools/HttpTestServer
        $ go run TestServer.go --dev_mode <Dev/Prod mode> --host <address of server> --port <port of server>
        ```