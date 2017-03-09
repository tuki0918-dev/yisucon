#!/bin/bash

set -exu -o pipefail

CURRENT_DIR=$(cd $(dirname $0); pwd)

yum localinstall -y https://packages.chef.io/files/stable/chef/12.18.31/el/7/chef-12.18.31-1.el7.x86_64.rpm

cd $CURRENT_DIR
chef-client -z -c client.rb -E $1 -o "recipe[$2]"
