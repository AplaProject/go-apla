[![Throughput Graph](https://graphs.waffle.io/AplaProject/go-apla/throughput.svg)](https://waffle.io/AplaProject/go-apla/metrics/throughput)

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/AplaProject/go-apla)](https://goreportcard.com/report/github.com/AplaProject/go-apla)
[![Build Status](https://travis-ci.org/AplaProject/go-apla.svg?branch=master)](https://travis-ci.org/AplaProject/go-apla)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg?style=flat)](http://apla.readthedocs.io/en/latest/)
[![](https://tokei.rs/b1/github/AplaProject/go-apla)](https://github.com/AplaProject/go-apla)
![](https://reposs.herokuapp.com/?path=AplaProject/go-apla&style=flat)
[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/AplaProject/go-apla)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-apla?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Slack Status](https://slack.apla.io/badge.svg)](https://slack.apla.io)

# About Apla

Apla blockchain platform is a secure, simple and compliant blockchain infrastructure for the fast-growing global collaborative economy. It was developed for building digital ecosystems. The platform includes an integrated application development environment with a multi-level system of access rights to data, interfaces and smart contracts.

For more information about Apla, visit [Apla website](https://apla.io).

We are open to new ideas and contributions and will be happy to see you among our active contributors to the source code, documentation, or whatever part you find inspiring in Apla. See our [Contribution Guide](https://github.com/AplaProject/go-apla/blob/master/CONTRIBUTING.md) for more information.

# Getting started

You can get started with Apla in several ways.

## Apla Testnet

**Apla Testnet** is the network for testing purposes. You can explore Apla features, build apps from scratch and test your apps in the real network environment.

You can explore Apla testnet from your browser. You don't need to install anything to do so. Just visit https://testapla0.apla.io/.

If you want to install Apla frontend (Molis) on your computer:

1. Download the latest [apla-front release for testnet](https://github.com/AplaProject/apla-front/releases).

2. Follow the instructions in [apla-front](https://github.com/AplaProject/apla-front) repository README. 


## Apla Quickstart

Apla Quickstart is a compact software package that you can use to deploy the Apla blockchain network on a local computer. Quickstart installs 1 to 5 nodes alongside the platform’s client software.

Quickstart is aimed at providing end users with an idea of how Apla blockchain works and includes usage examples of graphical interface elements as well as smart-contracts. 

Quickstart is available for computers running MacOS and Linux.

[Apla Quickstart for Linux and MacOS](https://github.com/AplaProject/quick-start)


## Deploying the Apla blockchain platform

Ready to deploy your own network? You can find out how to do that using our [Apla blockchain network deployment guide](https://apla.readthedocs.io/en/latest/howtos/deployment.html).


# About the backend components

Apla's backend has the following components:

  - go-apla service

    - TCP server
    - API server

  - PostgreSQL database system

  - Centrifugo notification service

## PostgreSQL database system

Each Apla node use PostgreSQL database system to store its current state database.

Testing and production environment considerations:

- *Testing environment*. You can deploy a single instance of PostgreSQL database system for all nodes. In this case, you must create PostgreSQL databases for each node. All nodes will connect to their databases located on one PostgreSQL instance.

- *Production environment*. It is recommended to use a separate instance of PostgreSQL database system for each node. Each node must connect only to its own PostgreSQL database instance. It is not requred to deploy this instance on the same host with other backend components.

## Centrifugo notification server

Centrifugo is a notification service that receives notifications from go-apla TCP-server and sends them to the frontend (Molis client) so that users can see status of their transactions.

Centrifugo is a unified notification service for all nodes in an Apla blockchain platform. When Molis client connects to a go-apla API service, it receives the IP-address of Centrifugo host and connects to it via a websocket.

Testing and production environment considerations:

- *Testing environment*. You can deploy centrifugo service on the same host with other backend components. It can be a single centrifugo service for all nodes, or each node may connect to its own centrifugo instance.

- *Production environment*. You must have at least several dedicated centrifugo hosts.

## Go-apla

Go-apla is the kernel of an Apla node. It consists of two services: TCP-server and API-server.

- TCP-server supports the interaction between Apla nodes.
- API-server supports the interaction between Molis clients and Apla nodes.

Testing and production environment considerations:

- *Testing environment*. You can deploy go-apla service with other backend components on one host.

- *Production environment*. You must deploy go-apla services on dedicated hosts.

# Installation instructions

> For a detailed guide, see [Apla blockchain network deployment guide](https://apla.readthedocs.io/en/latest/howtos/deployment.html).

## Directories

In this example, backend components are locatesd in the following directories:

* `/opt/apla/go-apla` go-apla.
* `/opt/apla/go-apla/node1` node data directory.
* `/opt/apla/centrifugo` centrifugo.

## Prerequisites and dependencies

- Go versions 1.10.x and above
- Centrifugo version 1.8
- Postgresql versions 10 and above

## Postgres database

1. Change user's password postgres to Apla's default password. 

```bash
    sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'apla'"
```

2. Create a node current state database.

```bash
    sudo -u postgres psql -c "CREATE DATABASE apladb"
```

## Centrifugo configuration

1. Specify Centrifugo secret in the Centrifugo configuration file. 

```bash
    echo '{"secret":"CENT_SECRET"}' > config.json
```

## Installing go-apla

1. Download and build the latest release:

```bash
    go get -v github.com/AplaProject/go-apla
```

2. Copy the go-apla binary from the Go workspace to the destination directory (`/opt/apla/go-apla` in this example).

```bash
    cp $HOME/go/bin/go-apla /opt/apla/go-apla
```

## Configure the node

1. Create the node configuration file:

```bash
    /opt/apla/go-apla/go-apla config \
        --dataDir=/opt/apla/go-apla/node1 \
        --dbName=apladb \
        --centSecret="CENT_SECRET" --centUrl=http://10.10.99.1:8000 \
        --httpHost=10.10.99.1 \
        --httpPort=7079 \
        --tcpHost=10.10.99.1 \
        --tcpPort=7078
```

2. Generate node keys:

```bash
    /opt/apla/go-apla/go-apla generateKeys \
        --config=/opt/apla/go-apla/node1/config.toml
```

3. Genereate the first block. If you are creating your own blockchain network. you must use the `--test=true` option. Otherwise you will not be able to create new accounts.

```bash
    /opt/apla/go-apla/go-apla generateFirstBlock \
        --config=/opt/apla/go-apla/node1/config.toml \
        --test=true
```

4. Initialize the database.

```bash

    /opt/apla/go-apla/go-apla initDatabase \
        --config=/opt/apla/go-apla/node1/config.toml
```

# Starting go-apla

To start the first node backend, you must start two services: centrifugo and go-apla.

1. Run centrifugo:

```bash
    /opt/apla/centrifugo/centrifugo \
        -a 10.10.99.1 -p 8000 \
        --config /opt/apla/centrifugo/config.json
```

2. Run go-apla:

```bash
    /opt/apla/go-apla/go-apla start \
        --config=/opt/apla/go-apla/node1/config.toml
```
