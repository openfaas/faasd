FROM golang:1.13-buster AS faasd
ADD . /go/src/github.com/openfaas/faasd
WORKDIR  /go/src/github.com/openfaas/faasd
RUN make

FROM debian:10
RUN apt-get update && \
  apt-get install -y --no-install-recommends \
  ca-certificates \
  curl \
  iptables \
  runc \
  systemd \
  systemd-sysv
# Install containerd
ARG CONTAINERD_VERSION=1.3.3
RUN mkdir -p /opt/cni/bin && \
  curl -sLSf https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}.linux-amd64.tar.gz \
  | tar xzvC /usr/local/bin/ --strip-components=1
ADD https://raw.githubusercontent.com/containerd/containerd/v${CONTAINERD_VERSION}/containerd.service /etc/systemd/system/
VOLUME /var/lib/containerd
# Install CNI
ARG CNI_VERSION=0.8.5
RUN curl -sLSf https://github.com/containernetworking/plugins/releases/download/v${CNI_VERSION}/cni-plugins-linux-amd64-v${CNI_VERSION}.tgz \
  | tar xzvC /opt/cni/bin
# Install faas-cli
RUN curl -sSLf https://cli.openfaas.com | sh
# Install faasd
COPY --from=faasd /go/src/github.com/openfaas/faasd/bin/faasd /usr/local/bin/
EXPOSE 8080
# Set up workdir for `faasd install` (executed in docker-2ndboot.sh)
COPY --from=faasd /go/src/github.com/openfaas/faasd /go/src/github.com/openfaas/faasd
WORKDIR /go/src/github.com/openfaas/faasd
# Install entrypoint for systemd
ADD https://raw.githubusercontent.com/AkihiroSuda/containerized-systemd/b9f98826563f12b36ab5474f172b5082f17d420c/docker-entrypoint.sh /
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT ["/docker-entrypoint.sh"]
# Install 2ndboot
ADD docker-2ndboot.sh /
CMD ["/docker-2ndboot.sh"]
