# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.12.1-stretch AS build_kubernetes
# rsync: building k8s
RUN apt update && apt install -y --no-install-recommends rsync
ARG K8S_VERSION=v1.14.0
WORKDIR /go/src/k8s.io/kubernetes
RUN git clone https://github.com/kubernetes/kubernetes . \
    && git checkout -b temp $K8S_VERSION \
    && make WHAT=cmd/kubectl \
    && make WHAT=test/e2e/e2e.test \
    && make ginkgo \
    && go build hack/e2e.go \
    && ./e2e
RUN rm -rf _output/local/go

FROM golang:1.12.1-stretch AS build_aks-engine
ARG AKSENGINE_VERSION=v0.32.3
WORKDIR /go/src/github.com/Azure/aks-engine
RUN git clone https://github.com/Azure/aks-engine . \
    && git checkout -b temp $AKSENGINE_VERSION \
    && make

FROM buildpack-deps:stretch-scm
# libxml-simple-perl, liblwp-useragent-determined-perl by k8s-azure
RUN apt update && apt install -y --no-install-recommends libxml-simple-perl liblwp-useragent-determined-perl
RUN curl -L https://packages.microsoft.com/keys/microsoft.asc | apt-key add -
RUN echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ stretch main" > \
        /etc/apt/sources.list.d/azure-cli.list \
    && apt-get install apt-transport-https \
    && apt-get update \
    && apt-get install azure-cli

ENV GOPATH="/home/k8s-ci/go"

COPY --from=build_kubernetes /go/bin $GOPATH/bin
COPY --from=build_kubernetes /go/src/k8s.io/kubernetes/cluster $GOPATH/src/k8s.io/kubernetes/cluster
COPY --from=build_kubernetes /go/src/k8s.io/kubernetes/hack $GOPATH/src/k8s.io/kubernetes/hack
COPY --from=build_kubernetes /go/src/k8s.io/kubernetes/_output $GOPATH/src/k8s.io/kubernetes/_output
COPY --from=build_kubernetes /go/src/k8s.io/kubernetes/e2e $GOPATH/src/k8s.io/kubernetes/e2e
COPY --from=build_aks-engine /go/src/github.com/Azure/aks-engine/bin/aks-engine /usr/local/bin/
COPY k8s-azure skip.txt /opt/k8s-azure/
COPY manifest /opt/k8s-azure/manifest
RUN ["/bin/bash", "-c", "ln -s /opt/k8s-azure/k8s-azure /usr/local/bin"]
ENV K8S_AZURE_MANIFEST_DIR='/opt/k8s-azure/manifest'
ENV K8S_AZURE_GLOBAL_SKIP_FILES='/opt/k8s-azure/skip.txt'
RUN useradd -m k8s-ci

COPY init.sh /usr/local/bin
ENTRYPOINT [ "init.sh" ]
