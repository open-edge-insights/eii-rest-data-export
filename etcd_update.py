"""
Copyright (c) 2020 Intel Corporation.

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
"""

import argparse
import sys
import etcd3
import json
from distutils.util import strtobool
import os
config_key = "/RestDataExport/config"


def parse_args():
    """Parse command line arguments.
    """
    arg_parser = argparse.ArgumentParser(
        formatter_class=argparse.ArgumentDefaultsHelpFormatter)
    arg_parser.add_argument('-hca', '--http_cert', default=None,
                            help='The ca cert of external HTTP Server')
    arg_parser.add_argument('-ca', '--ca_cert', default=None,
                            help='The ca cert required for etcd')
    arg_parser.add_argument('-c', '--cert', default=None,
                            help='The cert required for etcd')
    arg_parser.add_argument('-k', '--key', default=None,
                            help='The key cert required for etcd')
    arg_parser.add_argument('-host','--hostname', default='localhost',
                            help='Etcd host IP')
    arg_parser.add_argument('-port','--port', default='2379',
                            help='Etcd host port')

    return arg_parser.parse_args()


def get_etcd_client(hostname, port, ca_cert, root_key, root_cert):
    """Creates an EtcdCli instance

    :param ca_cert: Path of ca_certificate.pem
    :type ca_cert: String
    :param root_key: Path of root_client_key.pem
    :type root_key: String
    :param root_cert: Path of root_client_certificate.pem
    :type root_cert: String
    :return: etcd
    :rtype: etcd3.client()
    """
    # Default to localhost if ETCD_HOST is empty
    if hostname == "":
        hostname = "localhost"

    try:
        if ca_cert is None and root_key is None and root_cert is None:
            etcd = etcd3.client(host=hostname, port=port)
        else:
            etcd = etcd3.client(host=hostname, port=port,
                                ca_cert=ca_cert,
                                cert_key=root_key,
                                cert_cert=root_cert)
    except Exception as e:
        sys.exit("Exception raised when creating etcd \
                 client instance with error:{}".format(e))
    return etcd


def main():
    # Initializing DEV mode variable
    dev_mode = strtobool(os.getenv("DEV_MODE", "false"))
    # Initializing ETCD_PREFIX variable
    prefix = os.getenv("ETCD_PREFIX", "")

    # Initializing args
    args = parse_args()

    # Initializing etcd connection
    etcd_client = None
    http_ca_cert = ""
    if dev_mode:
        etcd_client = get_etcd_client(args.hostname, args.port, None, None, None)
    else:
        if not os.path.isdir("../build/provision/Certificates"):
            print("Please provision EII before continuing further...")
            os._exit(-1)
        etcd_client = get_etcd_client(args.hostname, args.port, args.ca_cert,
                                      args.key,
                                      args.cert)
        if args.http_cert is None:
            print("Please provide HttpServer ca cert path, exiting...")
            os._exit(-1)
        # Read CA cert from given path
        with open(args.http_cert) as f:
            http_ca_cert = f.read()

    # Fetching RestDataExport config
    cmd = etcd_client.get(prefix + config_key)
    rde_config = json.loads(cmd[0].decode('utf-8'))

    # Adding HttpServer CA to RestDataExport config
    rde_config["http_server_ca"] = http_ca_cert

    # Inserting the config with CA cert to etcd
    try:
        cmd = etcd_client.put(prefix + config_key, bytes(json.dumps(rde_config,
                                                         indent=4).encode()))
    except Exception as e:
        print("Failed to insert into etcd {}".format(e))


main()
