#! /bin/bash

mkdir -p /var/lib/faasd/secrets/
echo ${gw_password} > /var/lib/faasd/secrets/basic-auth-password

export FAASD_DOMAIN=${faasd_domain_name}
export LETSENCRYPT_EMAIL=${letsencrypt_email}

curl -sfL https://raw.githubusercontent.com/openfaas/faasd/master/hack/install.sh | bash -s -
