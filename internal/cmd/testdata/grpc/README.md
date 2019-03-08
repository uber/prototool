# TLS Certificates

Information on generating these certificates and keys was found at:
https://bbengfort.github.io/programmer/2017/03/03/secure-grpc.html

The certificates in this directory were generated with a combination of the [openssl](https://www.openssl.org/) 
and  [certstrap](https://github.com/square/certstrap). Both can be installed with homebrew packages of the same names.
For other OSs, follow the instructions on each site.

## Generation

The following commands will generate all the certificates in the "tls" folder with a blank passphrase and
 an expiration of 100 years 
```bash
cd internal/cmd/testdata/grpc
rm -fr tls

certstrap --depot-path tls init --common-name "cacert" --passphrase ""

certstrap --depot-path tls request-cert --ip 127.0.0.1 --cn server --passphrase ""
certstrap --depot-path tls sign server --CA cacert --expires "100 years" 

certstrap --depot-path tls request-cert --cn client --passphrase "" 
certstrap --depot-path tls sign client --CA cacert --expires "100 years" 

openssl req -new -newkey rsa:2048 -x509 -sha256 -nodes -days 36500 -extensions v3_req \
    -keyout tls/self-signed-server.key -out tls/self-signed-server.crt \
    -config <(cat <<-EOF
[req]
default_bits = 2048
req_extensions = v3_req
prompt = no
distinguished_name = dn

[ dn ]
C=US
ST=Denial
L=Springfield
O=Dis
emailAddress=server@server.local
CN = server.local

[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[ alt_names ]
IP.1 = 127.0.0.1
EOF
)
openssl req -new -newkey rsa:2048 -x509 -sha256 -nodes -days 36500 \
    -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=client.local" \
    -keyout tls/self-signed-client.key -out tls/self-signed-client.crt

cd ../../../..
```
