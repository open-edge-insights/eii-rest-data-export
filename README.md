**Contents**

- [RestDataExport](#restdataexport)
  - [Configuration](#configuration)
    - [HTTP Get APIs of RDE](#http-get-api-of-rde)
    - [Pre-requisites for posting metadata to HTTPServer](#pre-requisites-for-running-Rest-Data-Export-to-Post-on-HTTPServer)
  - [Service bring up](#service-bring-up)
# RestDataExport

RestDataExport service subscribes to any topic from EIIMessageBus and starts publishing meta data via POST requests to any external HTTP servers. It has an internal HTTP server running to respond to any GET requests for a required frame from any HTTP clients.

> IMPORTANT:
> RestDataExport service can subscribe classified results from both VideoAnalytics(video) or InfluxDBConnector(time-series) use cases. Please ensure the required service to subscribe from is mentioned in the Subscribers configuration in [config.json](config.json).

## Configuration

For more details on Etcd secrets and messagebus endpoint configuration, visit [Etcd_Secrets_Configuration.md](https://github.com/open-edge-insights/eii-core/blob/master/Etcd_Secrets_Configuration.md) and
[MessageBus Configuration](https://github.com/open-edge-insights/eii-core/blob/master/common/libs/ConfigMgr/README.md#interfaces) respectively.

## HTTP Get API of RDE

1. Getting the classifier results metadata
    ### Request

    `GET /metadata`
    ```sh
    $ curl -i -H 'Accept: application/json' http://<machine_ip_address>:8087/metadata
    ```
    For Eg:
    ```sh
      curl -i -H 'Accept: application/json' http://localhost:8087/metadata
    ```
    Output:

    ```sh
    HTTP/1.1 200 OK
    Content-Type: text/json
    Date: Fri, 08 Oct 2021 07:51:07 GMT
    Content-Length: 175
    {"channels":3,"defects":[],"encoding_level":95,"encoding_type":"jpeg","frame_number":558,"height":1200,"img_handle":"21af429f85","topic":"camera1_stream_results","width":1920}
    ```

2. Getting the Image using image handle
    > **Note**: For `image` API `imagestore` module is mandatory, from the `imagestore` only server fetches the data and return over REST API. Please include `imagestore` as a part of your usecase.
    ### Request

    `GET /image`
    ```sh
    $ curl -i -H 'Accept: image/jpeg' http://<machine_ip_address>:8087/image?img_handle=<imageid>
    ```
    For Eg.
      Storing the Image to Disk using curl along with `img_handle`
    ```sh
      curl -i -H 'Accept: application/image' http://localhost:8087/image?img_handle=21af429f85 > img.jpeg

      % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
      Dload  Upload   Total   Spent    Left  Speed
      100  324k    0  324k    0     0  63.3M      0 --:--:-- --:--:-- --:--:-- 63.3M
    ```
    > **Note**: This imageid which you are passing image can be found in the `metadata` API response as a part of metadata

### Pre-requisites for running Rest Data Export to Post on HTTPServer
  >**Note**: By default RDE will be serving the metadata as `GET` *only* request server.
  > By Enabling this, you can able to get the metadata using direct `GET` request. Also RDE will post the metadata to a http server
  Please follow the below steps as a prerequisites.

  1. Update the `RestDataExport/docker-compose.yml` file `HTTP_METHOD_FETCH_METADATA` environment value as follows.
    ```sh
    $   HTTP_METHOD_FETCH_METADATA="POST"
    ```

  2. If using the HttpTestServer, make sure that the server's IP has been added to 'no_proxy/NO_PROXY' vars in:

        - /etc/environment     (Needs restart/relogin)
        - ./docker-compose.yml (Needs to re-run the 'builder' step)

          ```sh
            environment:
              AppName: "RestDataExport"
              DEV_MODE: ${DEV_MODE}
              no_proxy: ${ETCD_HOST}, <IP of HttpTestServer>
          ```

  3. Run the below one-time command to install python etcd3

      ```sh
      $ pip3 install -r requirements.txt
      ```

  4. Ensure EII is provisioned and built.

  5. Ensure the pre-requisites for starting the TestServer application are enabled by following [README.md](https://github.com/open-edge-insights/eii-tools/blob/master/HttpTestServer/README.md#Pre-requisites-for-running-the-HttpTestServer).

  6. RestDataExport is pre-equipped with a python [tool](./etcd_update.py) to insert data into etcd which can be used to insert the required HttpServer ca cert into the config of RestDataExport before running it. The below commands should be run for running the tool which is a pre-requisite before starting RestDataExport:

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
      $ python3 etcd_update.py --http_cert "../tools/HttpTestServer/certificates/ca_cert.pem" --ca_cert "../build/helm-eii/eii-provision/Certificates/ca/ca_certificate.pem" --cert "../build/helm-eii/eii-provision/Certificates/root/root_client_certificate.pem" --key "../build/helm-eii/eii-provision/Certificates/root/root_client_key.pem" --hostname <IP address of ETCD host system> --port 32379

      # Required if running in DEV mode
      $ python3 etcd_update.py
      ```

  7. Start the TestServer application by following [README.md](https://github.com/open-edge-insights/eii-tools/blob/master/HttpTestServer/README.md#Starting-HttpTestServer).

  8. Ensure ImageStore application is running by following [README.md](https://github.com/open-edge-insights/video-imagestore/blob/master/README.md)

  9. Enure the topics you subscribe to are also added in the [config](config.json) with HttpServer endpoint specified
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

## Service bring up

Please go through the below sections to have RestDataExport service built and launch it:
- [../README.md#generate-deployment-and-configuration-files](https://github.com/open-edge-insights/eii-core/blob/master/README.md#generate-deployment-and-configuration-files)
- [../README.md#provision](https://github.com/open-edge-insights/eii-core/blob/master/README.md#provision)
- [../README.md#build-and-run-eii-videotimeseries-use-cases](https://github.com/open-edge-insights/eii-core/blob/master/README.md#build-and-run-eii-videotimeseries-use-cases)


## FAQs

Following are some of the FAQs of RestDataExport

### For setting environment proxy settings

  1. sudo vi /etc/systemd/system/docker.service.d/http-proxy.conf (update host-ip)
  2. sudo vi /etc/systemd/system/docker.service.d/https-proxy.conf (update hot-ip)
  3. env | grep proxy
  4. export no_proxy=$no_proxy,<host-ip>
  5. sudo vi ~/.docker/config.json (update host-ip in no_proxy)
  6. sudo systemctl daemon-reload
  7. sudo systemctl restart docker
