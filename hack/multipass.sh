#!/bin/bash

NAME=faasd
export OPENFAAS_URL="http://${NAME}.local:8080"

# replace ssh-key
sed -i "/ssh-rsa/c\  - $(cat ~/.ssh/id_rsa.pub)" cloud-config.txt

# build VM
multipass launch --cloud-init cloud-config.txt  --name "${NAME}"
notify-send "Multipass ${NAME} complete"

# get basic-auth-pw
ssh "ubuntu@${NAME}.local" "sudo cat /var/lib/faasd/secrets/basic-auth-password" > basic-auth-password

# wait for faasd svc to start
until curl -s "http://${NAME}.local:8080/ui/" > /dev/null
do
	echo "Waiting for faasd.."
	sleep 1
done

# log in
cat ./basic-auth-password | /usr/local/bin/faas-cli login --password-stdin

echo "# Gateway:"
echo "# export OPENFAAS_URL=http://${NAME}.local:8080"
