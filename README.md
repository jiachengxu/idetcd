# idetcd
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jiachengxu/idetcd/idetcd)
[![Build Status](https://img.shields.io/travis/jiachengxu/idetcd/master.svg?label=build)](https://travis-ci.org/jiachengxu/idetcd)
[![Code Coverage](https://img.shields.io/codecov/c/github/jiachengxu/idetcd/master.svg)](https://codecov.io/github/jiachengxu/idetcd?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jiachengxu/idetcd)](https://goreportcard.com/report/jiachengxu/idetcd)

*idetcd* is a etcd-based [CoreDNS](https://coredns.io/) plugin used for identifying nodes in a cluster without domain name collsion.

## Motivation
In distributed TensorFlow, identifying the nodes without domain name collision is a big [challenge](https://groups.google.com/a/tensorflow.org/forum/#!msg/developers/s8MJ2vqQ1z0/mWoVaAMvCwAJ;context-place=forum/developers). CoreDNS has a plugin-based architecture and it is a really lightweight, flexible and extandable DNS server which can easily enable custumized plugin. For solving this issue, we can set up the CoreDNS plus customized plugin on every node in the TensorFlow cluster, and using the plugin to write/read DNS records in a distributed key-value store, like zookeeper and etcd. This is what *idetcd* does.

## How it works
![deploy](https://github.com/jiachengxu/idetcd/blob/master/fig/deploy.png)

## Usage
You can get this project by:
```
$ go get -u github.com/jiachengxu/idetcd
```

Then you need to add a Corefile which specifys the configuration of the CoreDNS server in the same directory of `main.go`, an simple Corefile example is as follows, please go to [CoreDNS github repo](https://github.com/coredns/coredns) for more details. And for syntax of idetcd plugin, you can check in the [idetcd folder](https://github.com/jiachengxu/idetcd/tree/master/idetcd#idetcd).
 ~~~ corefile
 . {
     idetcd {
         endpoint http://localhost:2379
         limit 10
         pattern worker{{.ID}}.tf.local.
     }
     whoami
 }
 ~~~

And then you can generate binary file by:
```
$ go build -o coredns
```

Then run it by:
```
$ ./coredns
```

After that, the node in the cluster is trying to find a free slot in the etcd to expost itself, once it is successed(for example, it takes the worker4.tf.local. domain name), you can get the domain name of this node on every node in the same cluster by:
```
$ dig +short worker4.tf.local @localhost
```
Also ipv6 is supported:
```
$ dig +short worker4.tf.local AAAA @localhost
```
