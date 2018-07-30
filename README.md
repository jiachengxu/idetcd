# idetcd
[![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/jiachengxu/idetcd/idetcd)
[![Build Status](https://img.shields.io/travis/jiachengxu/idetcd/master.svg?label=build)](https://travis-ci.org/jiachengxu/idetcd)
[![Code Coverage](https://img.shields.io/codecov/c/github/jiachengxu/idetcd/master.svg)](https://codecov.io/github/jiachengxu/idetcd?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/jiachengxu/idetcd)](https://goreportcard.com/report/jiachengxu/idetcd)
idetcd is a etcd-based [CoreDNS](https://coredns.io/) plugin used for identifying nodes in a cluster without domain name collsion.

## Motivation
In distributed TensorFlow, identifying the nodes without domain name collision is a big [challenge](https://groups.google.com/a/tensorflow.org/forum/#!msg/developers/s8MJ2vqQ1z0/mWoVaAMvCwAJ;context-place=forum/developers). CoreDNS is a really lightweight, flexible and extandable plugin.

