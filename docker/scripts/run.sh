#!/usr/bin/env bash

set -x
set -eo pipefail

mkdir -p $GGBRIDGE_SSL_CERT_DIR
mkdir -p $GGBRIDGE_SSL_PRIVATE_CERT_DIR

cat $SSL_CERT_FILE >$GGBRIDGE_SSL_CERT_FILE

# Add private CA certificates bundle
if [[ -s $GGBRIDGE_SSL_PRIVATE_CERT_FILE ]]; then
  cat $GGBRIDGE_SSL_PRIVATE_CERT_FILE >>$GGBRIDGE_SSL_CERT_FILE
fi

# Add mTLS CA certificate
if [[ -s /etc/ggbridge/tls/ca.crt ]]; then
  cat /etc/ggbridge/tls/ca.crt >>$GGBRIDGE_SSL_CERT_FILE
fi

export SSL_CERT_FILE="$GGBRIDGE_SSL_CERT_FILE"
export NGINX_EMBEDDED="${NGINX_EMBEDDED:-true}"
exec /usr/bin/ggbridge $@
