# Contents

- [Contents](#contents)
  - [RestDataExport service](#restdataexport-service)
  - [Configuration](#configuration)
  - [HTTP GET API of RDE](#http-get-api-of-rde)
    - [Get the classifier results metadata](#get-the-classifier-results-metadata)
      - [Request to get metadata](#request-to-get-metadata)
    - [Get images using the image handle](#get-images-using-the-image-handle)
      - [Request to get images](#request-to-get-images)
    - [Prerequisites for running RestDataExport to POST on HTTP Servers](#prerequisites-for-running-restdataexport-to-post-on-http-servers)

> **Note:** In this document, you will find labels of 'Edge Insights for Industrial (EII)' for filenames, paths, code snippets, and so on. Consider the references of EII as Open Edge Insights (OEI). This is due to the product name change of EII as OEI.

## RestDataExport service

RestDataExport (RDE) service is a data service that serves GET and POST APIs. By default, the RDE service subscribes to a topic from the Message Bus and serves as GET API Server to respond to any GET requests for required metadata and frames. By enabling the POST API, the RDE service can publish the subscribed metadata to an external HTTP server.

>**Important:**
> RestDataExport service can subscribe classified results from the VideoAnalytics(video) or the InfluxDBConnector(time-series) use cases. In the Subscribers configuration of the [config.json](config.json) file, specify the required service to subscribe from.

## Configuration

For more details on the etcd secrets and messagebus endpoint configuration, refer to the following:

- [Etcd_Secrets_Configuration.md](https://github.com/open-edge-insights/eii-core/blob/master/Etcd_Secrets_Configuration.md)
- [MessageBus Configuration](https://github.com/open-edge-insights/eii-core/blob/master/common/libs/ConfigMgr/README.md#interfaces)

## HTTP GET API of RDE

The HTTP GET API of RDE allows you to get metadata and images. The following sections provide information about how to request metadata and images using the curl commands.

### Get the classifier results metadata

To get the classifier results metadata, refer to the following:

#### Request to get metadata

You can get the metadata for DEV mode and PROD mode.

- For the DEV mode:
`GET /metadata`

Run the following command:

```sh
curl -i -H 'Accept: application/json' http://<machine_ip_address>:8087/metadata
```

Refer to the following example:

```sh
curl -i -H 'Accept: application/json' http://localhost:8087/metadata
```

- For the PROD mode:
`GET /metadata`

Run the following command:

```sh
curl --cacert ../build/Certificates/rootca/cacert.pem -i -H 'Accept: application/json' https://<machine_ip_address>:8087/metadata
```

Refer to the following example:

```sh
curl --cacert ../build/Certificates/rootca/cacert.pem -i -H 'Accept: application/json' https://localhost:8087/metadata
```

Output:

The output for the previous command is as follows:

```sh
HTTP/1.1 200 OK
Content-Type: text/json
Date: Fri, 08 Oct 2021 07:51:07 GMT
Content-Length: 175
{"channels":3,"defects":[],"encoding_level":95,"encoding_type":"jpeg","frame_number":558,"height":1200,"img_handle":"21af429f85","topic":"camera1_stream_results","width":1920}
```

### Get images using the image handle

> **Note:** For the `image` API, the `imagestore` module is mandatory. From the `imagestore`, the server fetches the data, and returns it over the REST API. Include the `imagestore` module as a part of your use case.

#### Request to get images

`GET /image`

Run the following command:

```sh
curl -i -H 'Accept: image/jpeg' http://<machine_ip_address>:8087/image?img_handle=<imageid>
```

Refer to the following example to store image to the disk using curl along with `img_handle`:

```sh
curl -i -H 'Accept: application/image' http://localhost:8087/image?img_handle=21af429f85 > img.jpeg

% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
Dload  Upload   Total   Spent    Left  Speed
100  324k    0  324k    0     0  63.3M      0 --:--:-- --:--:-- --:--:-- 63.3M
```

> **Note:** You can find the `imageid` of the image in the `metadata` API response.

### Prerequisites for running RestDataExport to POST on HTTP Servers

>**Note:** By default, RDE will serve the metadata as `GET` *only* request server. By Enabling this, you can get the metadata using the `GET` request. Also, RDE will post the metadata to an HTTP server.
  
As a prerequisites, complete the following steps:

1. Update `[WORKDIR]/IEdgeInsights/build/.env` file `HTTP_METHOD_FETCH_METADATA` environment value as follows.

   ```sh
   HTTP_METHOD_FETCH_METADATA="POST"
   ```

   >**Note:** Make sure post changes you have rerun builder.py for generating updated deployment yml files.

2. If you are using the `HttpTestServer` then ensure that the server's IP address is added to the `no_proxy/NO_PROXY` vars in:

   - /etc/environment (Needs restart/relogin)
   - ./docker-compose.yml (Needs to re-run the 'builder' step)

   ```sh
   environment:
    AppName: "RestDataExport"
    DEV_MODE: ${DEV_MODE}
    no_proxy: ${ETCD_HOST}, <IP of HttpTestServer>
   ```

3. Run the following command to install `python etcd3`

   ```sh
   pip3 install -r requirements.txt
   ```

4. Build and provision OEI.

5. Ensure the prerequisites for starting the TestServer application are enabled. For more information, refer to the [README.md](https://github.com/open-edge-insights/eii-tools/blob/master/HttpTestServer/README.md#Pre-requisites-for-running-the-HttpTestServer).

6. As a prerequisite, before starting RestDataExport service, run the following commands.
  
   > **Note:** RestDataExport is pre-equipped with a python [tool](./etcd_update.py) to insert data into etcd, which can be used to insert the required `HttpServer ca cert` in the config of RestDataExport before running it.

   ```sh
   set -a && \
   source ../build/.env && \
   set +a
  
   # Required if running in the PROD mode only
   sudo chmod -R 777 ../build/Certificates/
  
   python3 etcd_update.py --http_cert <path to ca cert of HttpServer> --ca_cert <path to etcd   client ca cert> --cert <path to etcd client cert> --key <path to etcd client key> --hostname   <IP address of host system> --port <ETCD PORT>
  
   Example:
   # Required if running in the PROD mode
   python3 etcd_update.py --http_cert "../tools/HttpTestServer/certificates/ca_cert.pem"   --ca_cert "../build/Certificates/rootca/cacert.pem" --cert "../build/Certificates/root/root_client_certificate.pem" --key "../build/Certificates/root/root_client_key.pem" --hostname <IP address of host system> --port <ETCD PORT>
  
   # Required if running with k8s helm in the PROD mode
   python3 etcd_update.py --http_cert "../tools/HttpTestServer/certificates/ca_cert.pem" --ca_cert "../build/helm-eii/eii-deploy/Certificates/rootca/cacert.pem" --cert "../build/helm-eii/eii-deploy/Certificates/root/root_client_certificate.pem" --key "../build/helm-eii/eii-deploy/Certificates/root/root_client_key.pem" --hostname <Master Node IP address of ETCD host system> --port 32379

7. Start the TestServer application. For more information, refer to the [README.md](https://github.com/open-edge-insights/eii-tools/blob/master/HttpTestServer/README.md#Starting-HttpTestServer).

8. Ensure that the `ImageStore` application is running. For more information refer to the [README.md](https://github.com/open-edge-insights/video-imagestore/blob/master/README.md)

9. Ensure the topics you subscribe to are also added in the [config](config.json) with the `HttpServer endpoint` specified

10. Update the `config.json` file as follows:

    ```json
     {
      "camera1_stream_results": "http://IP Address of Test Server:8082",
      "point_classifier_results": "http://IP Address of Test Server:8082",
      "http_server_ca": "/opt/intel/eii/cert.pem",
      "rest_export_server_host": "0.0.0.0",
      "rest_export_server_port": "8087"
     }
    ```

    >**Note**: Ensure to rerun the `builder.py' script to generate the updated deployment yml files.

## Launch RestDataExport service

To build and launch the RestDataExport service, refer to the following:

- [../README.md#generate-deployment-and-configuration-files](https://github.com/open-edge-insights/eii-core/blob/master/README.md#generate-deployment-and-configuration-files)
- [../README.md#provision](https://github.com/open-edge-insights/eii-core/blob/master/README.md#provision)
- [../README.md#build-and-run-eii-videotimeseries-use-cases](https://github.com/open-edge-insights/eii-core/blob/master/README.md#build-and-run-eii-videotimeseries-use-cases)

## Setting environment proxy settings

To set the environment proxy settings for RDE, refer to the following:

1. To update the host-ip for http, run the following command:

   ```sh
   sudo vi /etc/systemd/system/docker.service.d/http-proxy.conf
   ```

2. To update the host-ip for https, run the following command:

   ```sh
   sudo vi /etc/systemd/system/docker.service.d/https-proxy.conf (update host-ip)
   ```

3. To check if the proxy settings have been applied, run the following command:

   ```sh
   env | grep proxy
   ```

4. To update the `no_proxy env` variable, run the following command:

   ```sh
   export no_proxy=$no_proxy,<host-ip>
   ```

5. To update docker proxy settings, run the following command:
  
   ```sh
   sudo vi ~/.docker/config.json (update host-ip in no_proxy)
   ```

6. To reload the docker daemon, run the following command:

   ```sh
   sudo systemctl daemon-reload
   ```

7. To restart the docker service with the updated proxy configurations, run the following command:
  
   ```sh
   sudo systemctl restart docker
   ```
