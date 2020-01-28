SHELL := /bin/bash
MULTIPASS := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))/.multipass

NAME := faasd
FAASD_IP = $(shell . $(MULTIPASS) ; fetch-multipass-ip $(NAME))

##########################################################
##@ Run
##########################################################
.PHONY: multipass-run mulitpass-clean

multipass-build: PUB=$(shell . $(MULTIPASS) ; inject-pub-key )
multipass-build:
	multipass launch --cloud-init cloud-config.txt --name $(NAME)

multipass-run: multipass-build	login ## run latest FaaSd in multipass VM

multipass-clean:	## clean Multipass VM
	ssh-keygen -f "$$HOME/.ssh/known_hosts" -R "$(FAASD_IP)"
	multipass stop $(NAME)
	multipass delete $(NAME)
	multipass purge

##########################################################
##@ DEV
##########################################################
.PHONY: dev-sync

dev-sync:	## sync repository to multipass VM
	ssh ubuntu@$(FAASD_IP) mkdir -p /home/ubuntu/go/src/github.com/alexellis/faasd
	rsync -avz . ubuntu@$(FAASD_IP):/home/ubuntu/go/src/github.com/alexellis/faasd

dev: dev-sync	## sync and install repository to multipass VM
	ssh ubuntu@$(FAASD_IP) 'cd /home/ubuntu/go/src/github.com/alexellis/faasd ; make install'

ssh: dev-sync ## ssh to multipass VM
	ssh -t ubuntu@$(FAASD_IP) 'cd /home/ubuntu/go/src/github.com/alexellis/faasd ; /bin/bash'

login: READY=$(shell . $(MULTIPASS) ; wait-for-faasd $(FAASD_IP) )
login: export OPENFAAS_URL = $(FAASD_IP):8080
login: ## faas authenticate to multipass VM
	ssh "ubuntu@$(FAASD_IP)" "sudo cat /var/lib/faasd/secrets/basic-auth-password" > basic-auth-password
	cat ./basic-auth-password | /usr/local/bin/faas-cli login --password-stdin
