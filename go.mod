module github.com/openfaas/faasd

go 1.16

require (
	github.com/alexellis/go-execute v0.5.0
	github.com/alexellis/k3sup v0.0.0-20210726065733-9717ee3b75a0
	github.com/compose-spec/compose-go v0.0.0-20200528042322-36d8ce368e05
	github.com/containerd/containerd v1.5.8
	github.com/containerd/go-cni v1.0.2
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/docker/cli v0.0.0-20191105005515-99c5edceb48d
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/docker v17.12.0-ce-rc1.0.20191113042239-ea84732a7725+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.3 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/morikuni/aec v1.0.0
	github.com/opencontainers/runtime-spec v1.0.3-0.20210326190908-1c3f411f0417
	github.com/openfaas/faas-provider v0.18.6
	github.com/openfaas/faas/gateway v0.0.0-20210726163109-539f0a2c946e
	github.com/pkg/errors v0.9.1
	github.com/sethvargo/go-password v0.2.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852
	github.com/vishvananda/netns v0.0.0-20200728191858-db3c7e526aae
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1
	k8s.io/apimachinery v0.21.3
)
