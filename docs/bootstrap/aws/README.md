# Provision faasd on AWS

1) [Sign up to AWS](https://portal.aws.amazon.com/billing/signup)
2) [Download Terraform](https://www.terraform.io)
3) Clone this gist using the URL from the address bar
4) Run `terraform init`
5) Configure terraform variables as needed by creating a `main.tfvars` file:

   | Variable     | Description         | Default         |
   | ------------ | ------------------- | --------------- |
   | `aws_access_key` | AWS access token | None |
   | `aws_secret_key` | AWS secret token | None |
   | `aws_region` | Region to deploy to | `eu-west-2`
   | `aws_ec2_instance_type` | Instance type for the EC2 instance | `t2.micro` |
   | `ssh_key_file` | Path to public SSH key file |`~/.ssh/id_rsa.pub` |

> Environment variables can also be used to set terraform variables when running the `terraform apply` command using the format `TF_VAR_name`.

6) Run `terraform apply`
   1) Add `-var-file=main.tfvars` if you have set the variables in `main.tfvars`.
   2) OR [use environment variables](https://www.terraform.io/docs/commands/environment-variables.html#tf_var_name) for setting the terraform variables when running the `apply` command

7) View the output for the login command and gateway URL i.e.

```
instance_ip = 178.128.39.201
gateway_url = http://178.128.39.201/
login_cmd = faas-cli login -g http://178.128.39.201/ -p rvIU49CEcFcHmqxj
password = rvIU49CEcFcHmqxj
```
8) Use your browser to access the OpenFaaS interface

Note that the user-data may take a couple of minutes to come up since it will be pulling in various components and preparing the machine. 
Also take into consideration the DNS propagation time for the new DNS record.

A single host with 1GB of RAM will be deployed for you, to remove at a later date simply use `terraform destroy`.
