## Instructions for testing a patch for faasd

### Launch a virtual machine

You can use any kind of Linux virtual machine, Ubuntu 20.04 is recommended.

Launch a cloud VM or use [Multipass](https://multipass.run), which is free to use an can be run locally. A Raspberry Pi 3 or 4 could also be used, but will need you to run `make dist` to cross compile a valid binary.

### Copy over your SSH key

Your SSH key will be used, so that you can copy a new faasd binary over to the host.

```bash
multipass launch \
  --mem 4G \
  -c 2 \
  -n faasd

# Then access its shell
multipass shell faasd

# Edit .ssh/authorized_keys

# Add .ssh/id_rsa.pub from your host and save the file
```

### Install faasd on the VM

You start off with the upstream version of faasd on the host, then add the new version over the top later on.

```bash
cd /tmp/
git clone https://github.com/openfaas/faasd --depth=1
cd faasd/hack
./install.sh

# Run the login command given to you at the end of the script
```

Get the multipass IP address:

```bash
export IP=$(multipass info faasd --format json| jq -r '.info.faasd.ipv4[0]')
```

### Build a new faasd binary with the patch

Check out faasd on your local computer

```bash
git clone https://github.com/openfaas/faasd
cd faasd

gh pr checkout #PR_NUMBER_HERE

GOOS=linux go build

# You can also run "make dist" which is slower, but includes
# a version and binaries for other platforms such as the Raspberry Pi
```

### Copy it over to the VM

Now build a new faasd binary and copy it to the VM:

```bash
scp faasd ubuntu@$IP:~/
```

Now deploy the new version on the VM:

```bash
killall -9 faasd-linux; killall -9 faasd-linux ; mv ./faasd-linux /usr/local/bin/faasd
```

### Check it worked and test that patch

Now run a command with `faas-cli` such as:

* `faas-cli list`
* `faas-cli version`

See the testing instructions on the PR and run through those steps.

Post your results on GitHub to assist the creator of the pull request.

You can see how to get the logs for various components using the [eBook Serverless For Everyone Else](https://gumroad.com/l/serverless-for-everyone-else), or by consulting the [DEV.md](/docs/DEV.md) guide.

