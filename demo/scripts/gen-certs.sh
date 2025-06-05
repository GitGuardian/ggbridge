#!/usr/bin/env bash

#set -x
set -eo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CERTS_DIR=${SCRIPT_DIR}/../certs

cd $CERTS_DIR

# Generates root CA
openssl genrsa -out ca.key 4096
openssl req -x509 -new -nodes -key ca.key -sha256 -days 3650 -out ca.crt \
  -subj "/CN=GGBridge CA"

# Generates server certificate
cat <<EOF > "server.ext"
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = ggbridge.gitguardian.public
EOF

openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr \
  -subj "/CN=ggbridge.gitguardian.public"
openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days 365 -sha256 -extfile server.ext
cat ca.crt >> server.crt

# Generates client certificate
cat <<EOF > "client.ext"
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

openssl genrsa -out client.key 2048
openssl req -new -key client.key -out client.csr \
  -subj "/CN=client.gitguardian.public"
openssl x509 -req -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out client.crt -days 365 -sha256 -extfile client.ext

# Generates GitGuardian server certificate
cat <<EOF > "gitguardian.ext"
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = api.gitguardian.public
DNS.2 = dashboard.gitguardian.public
DNS.3 = hook.gitguardian.public
EOF

openssl genrsa -out gitguardian.key 2048
openssl req -new -key gitguardian.key -out gitguardian.csr \
  -subj "/CN=*.gitguardian.public"
openssl x509 -req -in gitguardian.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out gitguardian.crt -days 365 -sha256 -extfile gitguardian.ext

# Generates VCS server certificate
cat <<EOF > "vcs.ext"
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = vcs.client.internal
EOF

openssl genrsa -out vcs.key 2048
openssl req -new -key vcs.key -out vcs.csr \
  -subj "/CN=vcs.client.internal"
openssl x509 -req -in vcs.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out vcs.crt -days 365 -sha256 -extfile vcs.ext

chmod 644 *
