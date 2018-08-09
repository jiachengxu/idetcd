#!/bin/bash
set -x
yum install -y docker
echo
chkconfig docker on
service docker start
echo
yum install -y git
git clone https://github.com/jiachengxu/idetcd.git /home/ec2-user/idetcd
cd /home/ec2-user/idetcd
docker run --rm -v $PWD:/go/src/github.com/jiachengxu/idetcd -w /go/src/github.com/jiachengxu/idetcd golang:1.10 go build -v -o coredns
./coredns
