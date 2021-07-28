**Contents**

- [`RestDataExport`](#restdataexport)
  - [`Configuration`](#configuration)
    - [`Pre-requisites`](#pre-requisites)
  - [`Service bring up`](#service-bring-up)
# `RestDataExport`

RestDataExport service subscribes to any topic from EIIMessageBus and starts publishing meta data via POST requests to any external HTTP servers. It has an internal HTTP server running to respond to any GET requests for a required frame from any HTTP clients.


## `Configuration`

For more details on Etcd secrets and messagebus endpoint configuration, visit [Etcd_Secrets_Configuration.md](../Etcd_Secrets_Configuration.md) and
[MessageBus Configuration](../common/libs/ConfigMgr/README.md#interfaces) respectively.

### `Pre-requisites`

  1. Run the below one-time command to install python etcd3

      ```sh
      $ pip3 install -r requirements.txt
      ```

  2. Ensure EII is provisioned.

  3. Ensure the pre-requisites for starting the TestServer application are enabled by following [README.md](../tools/HttpTestServer/README.md#Pre-requisites-for-running-the-HttpTestServer).

  4. RestDataExport is pre-equipped with a python [tool](./etcd_update.py) to insert data into etcd which can be used to insert the required HttpServer ca cert into the config of RestDataExport before running it. The below commands should be run for running the tool which is a pre-requisite before starting RestDataExport:

      ```sh
      $ set -a && \
        source ../build/.env && \
        set +a

      # Required if running in PROD mode only
      $ sudo chmod -R 777 ../build/provision/Certificates/

      $ python3 etcd_update.py --http_cert <path to ca cert of HttpServer> --ca_cert <path to etcd client ca cert> --cert <path to etcd client cert> --key <path to etcd client key> --hostname <IP address of host system> --port <ETCD PORT>

      Eg:
      # Required if running in PROD mode
      $ python3 etcd_update.py --http_cert "../tools/HttpTestServer/certificates/ca_cert.pem" --ca_cert "../build/provision/Certificates/ca/ca_certificate.pem" --cert "../build/provision/Certificates/root/root_client_certificate.pem" --key "../build/provision/Certificates/root/root_client_key.pem" --hostname <IP address of host system> --port <ETCD PORT>

      # Required if running with k8s helm in PROD mode
      $ python3 etcd_update.py --http_cert "../tools/HttpTestServer/certificates/ca_cert.pem" --ca_cert "../build/k8s/helm-eii/eii-provision/Certificates/ca/ca_certificate.pem" --cert "../build/k8s/helm-eii/eii-provision/Certificates/root/root_client_certificate.pem" --key "../build/k8s/helm-eii/eii-provision/Certificates/root/root_client_key.pem" --hostname <IP address of ETCD host system> --port 32379

      # Required if running in DEV mode
      $ python3 etcd_update.py
      ```

  5. Start the TestServer application by following [README.md](../tools/HttpTestServer/README.md#Starting-HttpTestServer).

  6. Ensure ImageStore application is running by following [README.md](../ImageStore/README.md)

  7. Enure the topics you subscribe to are also added in the [config](config.json) with HttpServer endpoint specified
    * Update the config.json file with the following settings:

      ```json
          {
              "camera1_stream_results": "http://IP Address of Test Server:8082",
              "point_classifier_results": "http://IP Address of Test Server:8082",
              "http_server_ca": "/opt/intel/eii/cert.pem",
              "rest_export_server_host": "0.0.0.0",
              "rest_export_server_port": "8087"
          }
      ```

## `Service bring up`

Please go through the below sections to have RestDataExport service built and launch it:
- [../README.md#generate-deployment-and-configuration-files](../README.md#generate-deployment-and-configuration-files)
- [../README.md#provision](../README.md#provision)
- [../README.md#build-and-run-eii-videotimeseries-use-cases](../README.md#build-and-run-eii-videotimeseries-use-cases)

