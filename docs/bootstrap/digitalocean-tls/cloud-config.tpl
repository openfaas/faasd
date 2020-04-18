#cloud-config
ssh_authorized_keys:
  - ${ssh_key}

groups:
  - caddy

users:
  - name: caddy
    gecos: Caddy web server
    primary_group: caddy
    groups: caddy
    shell: /usr/sbin/nologin
    homedir: /var/lib/caddy

write_files:
- content: |
    {
      email ${letsencrypt_email}
    }

    ${faasd_domain_name} {
      reverse_proxy 127.0.0.1:8080
    }

  path: /etc/caddy/Caddyfile

package_update: true

packages:
 - runc

runcmd:
- curl -sLSf https://github.com/containerd/containerd/releases/download/v1.3.2/containerd-1.3.2.linux-amd64.tar.gz > /tmp/containerd.tar.gz && tar -xvf /tmp/containerd.tar.gz -C /usr/local/bin/ --strip-components=1
- curl -SLfs https://raw.githubusercontent.com/containerd/containerd/v1.3.2/containerd.service | tee /etc/systemd/system/containerd.service
- systemctl daemon-reload && systemctl start containerd
- /sbin/sysctl -w net.ipv4.conf.all.forwarding=1
- mkdir -p /opt/cni/bin
- curl -sSL https://github.com/containernetworking/plugins/releases/download/v0.8.5/cni-plugins-linux-amd64-v0.8.5.tgz | tar -xz -C /opt/cni/bin
- mkdir -p /go/src/github.com/openfaas/
- mkdir -p /var/lib/faasd/secrets/
- echo ${gw_password} > /var/lib/faasd/secrets/basic-auth-password
- echo admin > /var/lib/faasd/secrets/basic-auth-user
- cd /go/src/github.com/openfaas/ && git clone https://github.com/openfaas/faasd
- curl -fSLs "https://github.com/openfaas/faasd/releases/download/0.8.1/faasd" --output "/usr/local/bin/faasd" && chmod a+x "/usr/local/bin/faasd"
- cd /go/src/github.com/openfaas/faasd/ && /usr/local/bin/faasd install
- systemctl status -l containerd --no-pager
- journalctl -u faasd-provider --no-pager
- systemctl status -l faasd-provider --no-pager
- systemctl status -l faasd --no-pager
- curl -sSLf https://cli.openfaas.com | sh
- sleep 5 && journalctl -u faasd --no-pager
- wget https://github.com/caddyserver/caddy/releases/download/v2.0.0-rc.2/caddy_2.0.0-rc.2_linux_amd64.tar.gz -O /tmp/caddy.tar.gz && tar -zxvf /tmp/caddy.tar.gz -C /usr/bin/ caddy
- wget https://raw.githubusercontent.com/caddyserver/dist/master/init/caddy.service -O /etc/systemd/system/caddy.service
- systemctl daemon-reload
- systemctl enable caddy
- systemctl start caddy
