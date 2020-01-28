SHELL := /bin/bash
MULTIPASS_IP := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))/.multipass-ip

NAME := faasd
DOMAIN := $(shell . $(MULTIPASS_IP) ; fetch-multipass-ip $(NAME))

##########################################################
##@ Run
##########################################################
.PHONY: multipass-run mulitpass-clean

multipass-run:	## run latest FaaSd in multipass VM
	sh hack/multipass.sh

multipass-clean:	## clean Multipass VM
	ssh-keygen -f "/home/gabeduke/.ssh/known_hosts" -R "$(DOMAIN)"
	multipass stop faasd
	multipass delete faasd

##########################################################
##@ DEV
##########################################################
.PHONY: dev-sync

dev-sync:	## sync repository to multipass VM
	ssh ubuntu@$(DOMAIN) mkdir -p /home/ubuntu/go/src/github.com/alexellis/faasd
	rsync -avz . ubuntu@$(DOMAIN):/home/ubuntu/go/src/github.com/alexellis/faasd

dev: dev-sync	## sync and install repository to multipass VM
	ssh ubuntu@$(DOMAIN) 'cd /home/ubuntu/go/src/github.com/alexellis/faasd ; make install'

ssh: dev-sync ## ssh to multipass VM
	ssh -t ubuntu@$(DOMAIN) 'cd /home/ubuntu/go/src/github.com/alexellis/faasd ; /bin/bash'

login: export OPENFAAS_URL = $(DOMAIN):8080
login: ## faas authenticate to multipass VM
	ssh "ubuntu@$(DOMAIN)" "sudo cat /var/lib/faasd/secrets/basic-auth-password" > basic-auth-password
	cat ./basic-auth-password | /usr/local/bin/faas-cli login --password-stdin
