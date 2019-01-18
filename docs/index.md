---
layout: default
title: REVA Documentation
---

# REVA Documentation

REVA software is organized in the following folders:

- [github.com/owncloud/revaold/revad](https://github.com/owncloud/revaold/tree/master/revad/): the core gRPC server that handles sharing and storage operations among others.

- [github.com/owncloud/revaold/ocproxy](https://github.com/owncloud/revaold/tree/master/ocproxy/): a proxy server that translates ownCloud  operations (WebDAV, OCS and apps) to the the revad server over gRPC.

- [github.com/cernbox/reva-cli](https://github.com/owncloud/revaold/tree/mater/reva-cli): a cli-tool to interact with the revad server.

# Installation

REVA is developed in the Go programming language, so make sure you have it installed on your platform, you can follow the guide at the official Go site: [https://golang.org/doc/install](https://golang.org/doc/install).
Also, be sure you have correctly configured the $GOPATH and $PATH variables, like for example:

```
# add to .bashrc
export GOPATH=/home/labkode/go/
export PATH=$GOPATH/bin
```

To get the software, just do:

```
go get github.com/owncloud/revaold/...
```

Run `reva-cli --help` to be sure that the software has been correclty deployed and available. 
