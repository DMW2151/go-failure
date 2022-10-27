#! /bin/bash

# provision server w. src && prometheus
export SERVER_HOSTNAME=$(terraform -chdir=./demo/infra/ output -json | jq -r '.server_hostname.value')
export DIGITALOCEAN_SSH_KEYNAME=$(terraform -chdir=./demo/infra/ output -json | jq -r '.digitalocean_ssh_keyname.value')

# build phi-server image
scp -r -i ~/.ssh/$DIGITALOCEAN_SSH_KEYNAME ./src/ root@$SERVER_HOSTNAME:/root/src

ssh -i ~/.ssh/$DIGITALOCEAN_SSH_KEYNAME root@$SERVER_HOSTNAME "mkdir -p /root/demo"


ssh -i ~/.ssh/$DIGITALOCEAN_SSH_KEYNAME root@$SERVER_HOSTNAME "mkdir -p /root/demo"
scp -r -i ~/.ssh/$DIGITALOCEAN_SSH_KEYNAME ./demo/server/ root@$SERVER_HOSTNAME:/root/demo/server