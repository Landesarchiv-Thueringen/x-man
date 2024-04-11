#!/usr/bin/env bash

set -e

# Go to repository root dir
cd -- "$( dirname -- "${BASH_SOURCE[0]}" )/.."

mkdir -p data/tls
cd data/tls

# Generate root key and certificate
openssl genrsa -out Test-RootCA.key 4096
openssl req -x509 -new -nodes -key Test-RootCA.key -sha256 -days 1826 \
    -out Test-RootCA.crt \
    -subj '/CN=Test Root CA/C=DE/ST=Thueringen/L=Weimar/O=Test Org'

function createServerCert() {
    # Create certificate key
    openssl req -new -nodes -out $1.csr -newkey rsa:4096 -keyout $1.key -subj "/CN=$1/C=DE/ST=Thueringen/L=Weimar/O=Test Org"
    cat > $1.v3.ext << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names
[alt_names]
DNS.1 = $1
EOF
    # Sign the certificate with Test-RootCA
    openssl x509 -req -in $1.csr -CA Test-RootCA.crt -CAkey Test-RootCA.key -CAcreateserial -out $1.crt -days 730 -sha256 -extfile $1.v3.ext
}

createServerCert ldap
createServerCert mailhog-tls

cp Test-RootCA.crt ../../server/data/ca-certificates/