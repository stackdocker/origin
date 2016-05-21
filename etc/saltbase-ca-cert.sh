#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

ETC_ROOT="$(cd $(dirname "${BASH_SOURCE}"); pwd)"
mkdir -p ~/kube
cp -u easy-rsa.tar.gz ~/kube

CERT_DIR=${ETC_ROOT}/kubernetes/cacerts
#use_cn=true
MASTER_IP=172.17.4.50
SERVICE_CLUSTER_IP_RANGE=10.3.0.0/24
EXTRA_SANS=IP:${MASTER_IP},IP:${SERVICE_CLUSTER_IP_RANGE%.*}.1,DNS:kubernetes,DNS:kubernetes.default,DNS:kubernetens.default.svc,DNS:kubernetes.default.svc.cluster.local


DEBUG='true' CERT_DIR=$CERT_DIR CERT_GROUP=$(id -gn) ${ETC_ROOT}/kubernetes/make-ca-cert.sh ${MASTER_IP} ${EXTRA_SANS}
