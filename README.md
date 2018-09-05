[![Throughput Graph](https://graphs.waffle.io/GenesisKernel/go-genesis/throughput.svg)](https://waffle.io/GenesisKernel/go-genesis/metrics/throughput)

[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](http://makeapullrequest.com)
[![Go Report Card](https://goreportcard.com/badge/github.com/GenesisKernel/go-genesis)](https://goreportcard.com/report/github.com/GenesisKernel/go-genesis)
[![Build Status](https://travis-ci.org/GenesisKernel/go-genesis.svg?branch=master)](https://travis-ci.org/GenesisKernel/go-genesis)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg?style=flat)](http://genesiskernel.readthedocs.io/en/latest/)
[![](https://tokei.rs/b1/github/GenesisKernel/go-genesis)](https://github.com/GenesisKernel/go-genesis)
![](https://reposs.herokuapp.com/?path=GenesisKernel/go-genesis&style=flat)
[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/GenesisKernel/go-genesis)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/go-genesis?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
[![Slack Status](https://slack.apla.io/badge.svg)](https://slack.apla.io)

# Genesis Getting Started Guide

## Table of contents

   * [Overview](#overview)
   * [Backend Install](#backend-install)
      * [Backend Install for Debian](#backend-install-deb)
        * [Backend Software Prerequisites](#backend-software-prerequisites-deb)
        * [First Node Deployment](#first-node-deployment-deb)
        * [Other Nodes Deployment](#other-nodes-deployment-deb)
      * [Backend Install for Windows](#backend-install-win)
        * [Backend Software Prerequisites](#backend-software-prerequisites-win)
        * [First Node Deployment](#first-node-deployment-win)
        * [Other Nodes Deployment](#other-nodes-deployment-win)
   * [Frontend Install](#frontend-install)
      * [Frontend Install for Debian](#frontend-install-deb)
        * [Frontend Software Prerequisites](#frontend-software-prerequisites-deb)
        * [Build Molis App](#build-molis-app-deb)
      * [Frontend Install for Windows](#frontend-install-win)
        * [Frontend Software Prerequisites](#frontend-software-prerequisites-win)
        * [Build Molis App](#build-molis-app-win)
   * [Launching](#launching)

## Overview <a name="overview"></a>

 Genesis is a platform which was developed for building digital ecosystems. Go-genesis is a backend for Genesis blockchain platform.

Genesis Blockchain Platform consists of two main components:

- Backend

  Contains:
  - Centrifugo notification service
  - Go-Genesis kernel service (includes Apla TCP and API servers)
  - PostgreSQL database
  
- Frontend
  
  Molis client is the frontend for Genesis. It can be built as a native OS application or a web-application.
  
In a production environment, each of these components (backend and frontend) can be deployed on different hosts and operating systems.

In this guide we will deploy Genesis Blockchain Platform based on three nodes with the same OS on the test ICT-infrastructure and build Molis client. 

As a node OS we will use:
 - Debian 9 (Stretch) 64-bit [official distributive](https://www.debian.org/CD/http-ftp/#stable)
   - with installed GNOME GUI in a case of building Molis client on your Debian host
   - minimal server core installation in a case of deployment only backend components
 - Windows Server 2012R2/2016

For testing purposes, all of these hosts are connected to each other in a simple network. In the table below, there are network settings for each node component that we will deploy through this guide:

|Node Number| Component | IP and Port |
|:---------:|-----------|-------------|
| 1 | PostgreSQL | 127.0.0.1:5432|
| 1 | Centrifugo | 10.10.99.1:8000|
| 1 | Go-Genesis (TCP-server) | 10.10.99.1:7078|
| 1 | Go-Genesis (API-server) | 10.10.99.1:7079|
| 2 | PostgreSQL | 127.0.0.1:5432|
| 2 | Centrifugo | 10.10.99.2:8000|
| 2 | Go-Genesis (TCP-server) | 10.10.99.2:7078|
| 2 | Go-Genesis (API-server) | 10.10.99.2:7079|
| 3 | PostgreSQL | 127.0.0.1:5432|
| 3 | Centrifugo | 10.10.99.3:8000|
| 3 | Go-Genesis (TCP-server) | 10.10.99.3:7078|
| 3 | Go-Genesis (API-server) | 10.10.99.3:7079|

## ***Backend Install*** <a name="backend-install"></a>

In  this section we will deploy Genesis Backend components.

Genesis Blockchain Platform backend consists of three main components:

1) **PostgreSQL database system**

Each Genesis node use PostgreSQL database system for store its current state database. 

In a testing environment, you can deploy just one instance of PostgreSQL database system for all nodes. In this case, you must create PostgreSQL databases for each node. All nodes will connect to their databases located on one PostgreSQL instance.

In a production environment, it is not recommended to have one PostgreSQL database system for all nodes. Each Genesis node must have its own instance of PostgreSQL and should connect only to it. It is not necessary to deploy this instance on the same host with other backend components.

For testing purposes, in this guide, we will deploy PostgreSQL on each Genesis node.

2) **Centrifugo notification server**

Centrifugo is a notification service which receives notifications from Go-Genesis TCP-server and sends them to the frontend (Molis client), so that users can see status of their transactions.

Centrifugo is a unified notification service for all nodes in Genesis Blockchain Platform. When Molis client connects to Go-Genesis API-service, it receives the IP-address of Centrifugo host and connects to it via a websocket.

In a testing environment, you can deploy centrifugo service on the same host with other backend components. It can be a single centrifugo service for all nodes, or each node may connect to its own centrifugo instance.

In a production environment, you must have at least several dedicated centrifugo hosts.

For testing purposes, in this guide, we will deploy Centrifugo service on each Genesis node.

3) **Go-Genesis**

Go-Genesis is the kernel of a Genesis node. It consists of two services: TCP-server and API-server.

TCP-server is responsible for Genesis nodes interconnection.

API-server is responsible for connections with Molis clients.

In a testing environment, you can deploy Go-Genesis service with other backend components on one host.

In a production environment, you must deploy Go-Genesis services on dedicated hosts.

For testing purposes, in this guide, we will deploy Go-Genesis services on the same host with other backend components.


## Backend Install for Debian OS <a name="backend-install-deb"></a>

### Backend Software Prerequisites <a name="backend-software-prerequisites-deb"></a>

Before installing Genesis Backend components, you need to install additional software.

#### Install sudo

All commands for Debian 9 must be run as a non-root user. But some system commands need superuser privileges to be executed. By default, sudo is not installed on Debian 9, and first, you should install it.

1) Become the root superuser:

```bash
su - 
```

2) Upgrade your system:

```bash
apt update -y && apt upgrade -y && apt dist-upgrade -y
```

3) Install sudo:

```bash
apt install sudo -y
```

4) Add your user to the sudo group:

```bash
usermod -a -G sudo user
```

5) After the reboot, the changes take effect.

#### Install common software

Some of used packages can be downloaded from the official Debian repository. Install packages:
```
$ sudo apt install -y git curl apt-transport-https build-essential
```

#### Create a directory for Genesis

For Debian 9 OS, it is recommended to store all software used by Genesis Blockchain Platform in a special directory. 

In this guide, we will use /opt/genesis directory, but you can change it to your own.

1) Make directory and go to it:

```bash
sudo mkdir /opt/genesis && cd /opt/genesis
```
2) Make your user owner of this directory:

```bash
sudo chown user /opt/genesis/
```

#### Install Go Language

1) Download the latest stable version of Go (1.10) from the official site or via the command line:

```bash
wget https://dl.google.com/go/go1.10.1.linux-amd64.tar.gz
```

2) Install Go:

```bash
tar -xvf go1.10.1.linux-amd64.tar.gz && sudo mv go /usr/local/
```

3) Export Go environment variables:

```bash
export GOROOT=/usr/local/go && export GOPATH=/opt/genesis/ && export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
```

4) Remove the temporary file:

```bash
rm go1.10.1.linux-amd64.tar.gz
```


#### Install Python packages

These packages must installed only on the first node because it must execute special scripts.

1) Install Python3-pip:

```bash
sudo apt install -y python3-pip
```

2) Download and install required python packages:

```bash
sudo wget https://raw.githubusercontent.com/GenesisKernel/genesis-tests/master/requirements.txt && sudo pip3 install -r requirements.txt
```


#### OS Firewall Requirements

By default, after installing Debian 9, there are no firewall rules. But if you want to design more secure system with a firewall, following incoming connections should be allowed:

-	7078/TCP - Node's TCP-server
-	7079/TCP - Node's API-server
-	8000/TCP - Centrifugo server

### First Node Deployment <a name="first-node-deployment-deb"></a>

#### Install PostgreSQL <a name="install-postgres-deb"></a>

1) Install PostgreSQL:

```bash
sudo apt install -y postgresql
```

2) Change user's password postgres to Genesis' default. You can set your own password, but then you also must change it in the node configuration file config.toml.

```bash
sudo -u postgres psql -c "ALTER USER postgres WITH PASSWORD 'genesis'"
```

3) Create a node current state database, for example 'genesisdb':
```
$ sudo -u postgres psql -c "CREATE DATABASE genesisdb"
```

#### Install Centrifugo <a name="install-centrifugo-deb"></a>

1) Download Centrifugo version 1.7.9 from [GitHub](https://github.com/centrifugal/centrifugo/releases/) or via command line:

```bash
wget https://github.com/centrifugal/centrifugo/releases/download/v1.7.9/centrifugo-1.7.9-linux-amd64.zip && unzip centrifugo-1.7.9-linux-amd64.zip && mkdir centrifugo && mv centrifugo-1.7.9-linux-amd64/* centrifugo/
```

2) Remove temporary files:

```bash
$ rm -R centrifugo-1.7.9-linux-amd64 && rm centrifugo-1.7.9-linux-amd64.zip
```

3) Create Centrifugo configuration file:

```bash
$ echo '{"secret":"CENT_SECRET"}' > centrifugo/config.json
```

You can set your own "secret", but then you also must change it in node configuration file config.toml.

#### Install Go-Genesis

1) Create go-genesis and node1 directories:
```
$ mkdir go-genesis && cd go-genesis && mkdir node1
```
2) Download and build the latest release of Go-Genesis from [GitHub](https://github.com/GenesisKernel/go-genesis/releases), and then copy it into the go-genesis directory:

```bash
go get -v github.com/GenesisKernel/go-genesis && cp $GOPATH/bin/go-genesis /opt/genesis
```


3) Create Node 1 configuration file, all used network settings (IP-adresses and ports) are described in [Overview](#overview):

```bash
./go-genesis config --dataDir=/opt/genesis/go-genesis/node1 --dbName=genesisdb --privateBlockchain=true --centSecret="CENT_SECRET" --centUrl=http://10.10.99.1:8000 --httpHost=10.10.99.1 --tcpHost=10.10.99.1
```
Where:

- --dbName - database name 'genesisdb' that was created in section [Install PostgreSQL](#install-postgres-deb)
- --centSecret - Centrifugo secret 'CENT_SECRET' that was created in section [Install Centrifugo](#install-centrifugo-deb)
- --centUrl=ht&#8203;tp://10.10.99.1:8000 - used IP address and port of Centrifugo of Node 1
- --httpHost=10.10.99.1 - used IP address and port of API-server of Node 1
- --tcpHost=10.10.99.1 - used IP address and port of TCP-server of Node 1
- Other usage and flags of go-genesis are described in [documentation](http://genesiskernel.readthedocs.io/en/latest/)

4) Generate Node 1 keys:

```bash
$ ./go-genesis generateKeys --config=node1/config.toml
```

5) Generate the first block:

```bash
$ ./go-genesis generateFirstBlock --config=node1/config.toml
```

6) Initialize the database:

```bash
./go-genesis initDatabase --config=node1/config.toml
```


#### Create services for backend components

This section is under development.


#### Start First Node

To start the first node you must start two services:

-	centrifugo
-	go-genesis

If you did not create these services, you can just execute binary files from their directories in different consoles.

1) Execute centrifugo file:

```bash
$ cd /opt/genesis/centrifugo && ./centrifugo -a 10.10.99.1 --config=config.json
```

Where:

 - 10.10.99.1 - IP-address of Node 1
 - --config=config.json - path to centrifugo configuration file 'config.json'
 
2) Execute go-genesis file in another console:

```bash
$ cd /opt/genesis/go-genesis/ && ./go-genesis start --config=node1/config.toml
```

Now, you can connect to your node via Molis client.

### Other Nodes Deployment <a name="other-nodes-deployment-deb"></a>

Deployment of the second node and others is similar to the first node, but has some differences in creation of go-genesis config.toml file.

For each other node deployment you must repeat the following steps:

- Install Backend Software Prerequisites
- Install PostgreSQL
- Install Centrifugo 
- Install Go-Genesis

#### Other Nodes Configuration

In this example we will configure Node 2. Other Nodes can be configured in the same way. All used network settings (IP-adresses and ports) are described in [Overview](#overview).

1) Copy file of the first block to Node 2. For example, you can do it via scp on Node 2:

```bash
scp user@10.10.99.1:/opt/genesis/go-genesis/node1/first /opt/genesis/go-genesis/node2/
```

2) Create Node 2 configuration file:

```bash

./go-genesis config --dataDir=/opt/genesis/go-genesis/node2 --dbName=genesisdb  --privateBlockchain=true --centSecret="CENT_SECRET" --centUrl=http://10.10.99.2:8000 --httpHost=10.10.99.2 --tcpHost=10.10.99.2 --nodesAddr=10.10.99.1
```

Where:

- --dbName - database name 'genesisdb' that was created in section [Install PostgreSQL](#install-postgres-deb)
- --centSecret - Centrifugo secret 'CENT_SECRET' that was created in section [Install Centrifugo](#install-centrifugo-deb)
- --centUrl=ht&#8203;tp://10.10.99.2:8000 - used IP address and port of Centrifugo of Node 2
- --httpHost=10.10.99.2 - used IP address and port of API-server of Node 2
- --tcpHost=10.10.99.2 - used IP address and port of TCP-server of Node 2
- --nodesAddr=10.10.99.1 - IP-address of Node 1
- Other usage and flags of go-genesis are described in [documentation](http://genesiskernel.readthedocs.io/en/latest/)

3) Generate Node 2 keys:

```bash
./go-genesis generateKeys --config=node2/config.toml
```

4) Initialize the database:

```bash
./go-genesis initDatabase --config=node2/config.toml
```

5) Start Node 2:

```bash
./go-genesis start --config=node2/config.toml 
```

Ignore the showed errors. If you start node with log level "INFO", you'll see that node starts downloading blocks.


#### Adding keys

Errors that occurred above are caused by untrusted relationships between nodes. To fix it, add the second node public key to the first node.

To add keys, download this script [updateKeys.py](https://github.com/GenesisKernel/genesis-tests/blob/master/scripts/updateKeys.py). All information that is needed for  script execution is located in node's directory 'nodeN'. This scipt must be executed on the first node with founder's privileges. Execute the script with the following arguments:

```bash
python3 updateKeys.py PrivateKey1 Host1 Port1 KeyID2 PublicKey2 balance
```

Where:
-	PrivateKey1 - founder private key, located in the file PrivateKey of the first node
-	Host1 - IP-addres or DNS-name of the first node
-	Port1 - the first node API-server port
-	KeyID2 - content of file KeyID of the second node
-	PublicKey2 - content of file PublicKey of the second node
-	balance - set wallet balance of the second node

**Example**: 

```bash
python3 updatekeys.py bda1c45d3298cb7bece1f76a81d8016d33cdec18c925297c7748621c502a23f2 10.10.99.1 7079 -5910245696104921893 1812246837170b6df8609fd9d846a0984f4e5b3ee9037717e39dc38c82ea1a8e528c9e6f6acdc06b2a33f228c4d2649005bde47af857f3f756aaf64d3f1648dd 1000000000000000000000
```

All used network settings (IP-adresses and ports) are described in [Overview](#overview).

This script will create a contract that adds the second node public key to the table 'keys' of the database.

#### Create connection between nodes

Next, you must create connection between nodes. For this, you should download this script [newValToFullNodes.py](https://github.com/GenesisKernel/genesis-tests/blob/master/scripts/newValToFullNodes.py). All information that is needed for  script execution is located in node's directory 'nodeN'.

Execute the script with the following arguments:

```bash
python3 newValToFullNodes.py PrivateKey1 Host1 Port1 'NewValue'
```

Where:
-	PrivateKey1 - founder private key, located in the file PrivateKey of the first node
-	Host1 - IP-addres or DNS-name of the first node
-	Port1 - the first node API-server port
-	NewValue - new value of Full_Nodes parameter

Argument **NewValue** must be written in json format:

```json
[
 {
  "tcp_address":"Host1:tcpPort1", 
  "api_address":"http://Host1:httpPort1", 
  "key_id":"KeyID1", 
  "public_key":"NodePubKey1"
 },
 {
  "tcp_address":"Host2:tcpPort2", 
  "api_address":"http://Host2:httpPort2", 
  "key_id":"KeyID2", 
  "public_key":"NodePubKey2"
 },
 {
  "tcp_address":"HostN:tcpPortN", 
  "api_address":"http://HostN:httpPortN", 
  "key_id":"KeyIDN", 
  "public_key":"NodePubKeyN"
 }
]
```

Where:
-	Host1 - IP-addres or DNS-name of the first node
-	tcpPort1 - the first node TCP-server port
-	httpPort1 - the first node API-server port
-	KeyID1 - content of file KeyID of the first node
-	NodePubKey1 - content of file NodePublicKey of the first node
-	Host2 - IP-addres or DNS-name of the second node
-	tcpPort2 - the second node TCP-server port
-	httpPort2 - the second node API-server port
-	KeyID2 - content of file KeyID of the second node
-	NodePubKey2 - content of file NodePublicKey of the second node
-	HostN - IP-addres or DNS-name of node N
-	tcpPortN - node N TCP-server port
-	httpPortN - node N API-server port
- KeyIDN - content of file KeyID of node N
-	NodePubKeyN - content of file NodePublicKey of node N

**Example:**

```bash 
python3 newValToFullNodes.py bda1c45d3298cb7bece1f76a81d8016d33cdec18c925297c7748621c502a23f2 10.10.99.1 7079 '[{"tcp_address":"10.10.99.1:7078","api_address":"http://10.10.99.1:7079","key_id":"5541394763743537703","public_key":"d26824d0e94894bae9e983e7a386a1c9e4f609990d4b635b6926b52c831d6ec28b95f75acf0c9d10ee96afc0dd02617f08fea225706f0e502d5fe26587023e3b"},{"tcp_address":"10.10.99.2:7078","api_address":"http://10.10.99.2:7079","key_id":"6404048169476933259","public_key":"afd9ed260ec65a2a294794285ad40c5edc219e3be2455a044e2444111b8525815b224fdb369aa17307434d0e6aca8f9c959f823756baeb9ccb105f96f996bf11" }, {"tcp_address":"10.10.99.3:7078","api_address":"http://10.10.99.3:7079","key_id":"-5910245696104921893","public_key":"254c38cd6d9f47ffc42a8d178bb47f9a0cbc46ec6ef4d972c05146bfe87a8da03cb3450b71b2a724fdb2184163ae91023931c9fe5f148f0bdceeeefc5a16fe58"}]'
```
All used network settings (IP-adresses and ports) are described in [Overview](#overview).

Now, all nodes are connected to each other.

### ***Backend Install for Windows Server OS*** <a name="backend-install-win"></a>

### Backend Software Prerequisites <a name="backend-software-prerequisites-win"></a>

Before installing Genesis Backend components, you need to install additional software. To do this, you must have administrators privileges.

#### Install Go Language

1) Download Go latest stable version 1.10 for Windows from the [official site](https://golang.org/dl/).

2) Install Go without any specific settings.

#### Install Git

1) Download the latest 64-bit Git for Windows from the [official site](https://git-scm.com/download/win).

2) Install Git without any specific settings.

#### Install MinGW

You must install MinGW software only if you want to build Genesis backend from source code.

1) Download the latest MinGW-W64 from its [site](https://sourceforge.net/projects/mingw-w64/).

2) During the installation process, you must specify the following setup settings:

- Version: from the drop list, select the latest version
- Architecture: from the drop list select "x86_64"
- Threads: from the drop list select "win32"

Leave default values for other settings.

3) Add absolute path of directory "mingw64/bin" to the system environment variable PATH via command line, for example "C:\mingw64\bin":

```
> setx PATH “C:\mingw64\bin”
```

4) For Windows Server 2016, you must restart your system.


#### Install Python 3

1) Python 3 must be installed only on the first node because special scripts must be executed on this node.

2) Download latest Python 3 Release from the [official site](https://www.python.org/downloads/).

3) During installation process, select "Add python.exe to Path" in features tree. Leave default values for other settings.

4) Download python packages list "requirements.txt" from [GitHub](https://raw.githubusercontent.com/GenesisKernel/genesis-tests/master/requirements.txt).

5) For script execution, install additional packages via "requirements.txt":

```
> py -m pip install -r requirements.txt
```

#### OS Firewall Requirements

In Windows Server firewall settings, you must allow the following incoming connections:

-	7078/TCP - Node's TCP-server
-	7079/TCP - Node's API-server
-	8000/TCP - Centrifugo server

### First Node Deployment <a name="first-node-deployment-win"></a>

#### Install PostgreSQL <a name="install-postgres-win"></a>

1) Download PostgreSQL 10.4 installer for Windows x86-64 from the [official site](https://www.enterprisedb.com/downloads/postgres-postgresql-downloads).

2) During the installation process, you must:

- specify a default installation directory
- specify all selected components
- specify default data directory
- set a password for the database superuser (postgres), for example 'genesis'
- specify the default port 5432 the server must listen on (you can set your own port, but also you must change it in node configuration file config.toml)
- select a default locale to be used by the new database cluster
- after the setup wizard is completed, don’t launch stack builder

3) Run pgAdmin4 app and create current state database for the node, for example 'genesisdb'

#### Install Centrifugo <a name="install-centrifugo-win"></a>

1) Download Centrifugo-1.7.9-windows-amd64.zip from [GitHub](https://github.com/centrifugal/centrifugo/releases/).

2) Unzip archive to your Centrifugo folder inside Genesis directory

3) By any text editor, create Centrifugo configuration file config.json in the centrifugo directory. Add the following line to config.json file:

```json
{"secret":"CENT_SECRET"}
```

You can set your own "secret", but also you must change it in node configuration file config.toml.

#### Install Go-Genesis

1) In Genesis directory, create go-genesis directory and node folder inside it.

2) Download Go-Genesis from [GitHub](https://github.com/GenesisKernel/go-genesis/releases) or build latest release by command line:
```
> cd C:\Genesis\go-genesis
> go get -v github.com/GenesisKernel/go-genesis
> go build github.com/GenesisKernel/go-genesis
```

After that, go-genesis.exe file will appear in go-genesis directory.

Usage and flags of go-apla.exe file are described in [documentation](http://genesiskernel.readthedocs.io/en/latest/).

3) Create Node 1 config.toml configuration file. All used network settings (IP-adresses and ports) are described in [Overview](#overview).

```
> go-genesis.exe config --dataDir=C:\Genesis\go-genesis\node --dbName=genesisdb --privateBlockchain=true --centSecret="CENT_SECRET" --centUrl=http://10.10.99.1:8000 --httpHost=10.10.99.1 --tcpHost=10.10.99.1
```

Where:

- --dbName - database name 'genesisdb' that was created in section [Install PostgreSQL](#install-postgres-win)
- --centSecret - Centrifugo secret 'CENT_SECRET' that was created in section [Install Centrifugo](#install-centrifugo-win)
- --centUrl=ht&#8203;tp://10.10.99.1:8000 - used IP address and port of Centrifugo of Node 1
- --httpHost=10.10.99.1 - used IP address and port of API-server of Node 1
- --tcpHost=10.10.99.1 - used IP address and port of TCP-server of Node 1
- Other usage and flags of go-genesis are described in [documentation](http://genesiskernel.readthedocs.io/en/latest/)

4) Generate Node 1 keys:

```
> go-genesis.exe generateKeys --config=node\config.toml
```

5) Generate the first block:

```
> go-genesis.exe generateFirstBlock --config=node\config.toml
```

6) Initialize the database:

```
> go-genesis.exe initDatabase --config=node\config.toml
```

#### Create services for backend components

This section is under development.


#### Start the First Node

To start the first node you must start two services:

- centrifugo
- go-genesis

If you did not create these services, you can just execute .exe files from its directories in different command prompts.

1) Run centrifugo.exe:

```
> centrifugo.exe -a 10.10.99.1 --config=config.json
```

Where:

 - 10.10.99.1 - IP-address of Node 1
 - --config=config.json - path to centrifugo configuration file 'config.json'

2) Run go-genesis.exe:

```
> go-genesis.exe start --config=node\config.toml
```

Now, you can connect to your node via Molis client.

### Other Nodes Deployment <a name="other-nodes-deployment-win"></a>

Deployment of the second node and others is similar to the first node, but has some differences in creation of go-genesis config.toml file.

For each other node deployment you should repeat the following steps:

- Install Backend Software Prerequisites
- Install PostgreSQL
- Install Centrifugo 
- Install Go-Apla

#### Other Nodes Configuration

In this example we will configure Node 2. Other Nodes can be configured in the same way. All used network settings (IP-adresses and ports) are described in [Overview](#overview).

1) Copy file of the first block to Node 2 in the same directory. You can see the default location of the first block file in config.toml file.

2) Create Node 2 config.toml configuration file:

```
> go-genesis.exe config --dataDir=C:\Genesis\go-genesis\node --dbName=genesisdb --privateBlockchain=true --centSecret="CENT_SECRET" --centUrl=http://10.10.99.2:8000 --httpHost=10.10.99.2 --tcpHost=10.10.99.2 --nodesAddr=10.10.99.1
```

Where:

- --dbName - database name 'genesisdb' that was created in section [Install PostgreSQL](#install-postgres-win)
- --centSecret - Centrifugo secret 'CENT_SECRET' that was created in section [Install Centrifugo](#install-centrifugo-win)
- --centUrl=ht&#8203;tp://10.10.99.2:8000 - used IP address and port of Centrifugo of Node 2
- --httpHost=10.10.99.2 - used IP address and port of API-server of Node 2
- --tcpHost=10.10.99.2 - used IP address and port of TCP-server of Node 2
- --nodesAddr=10.10.99.1 - IP-address of Node 1
- Other usage and flags of go-genesis are described in [documentation](http://genesiskernel.readthedocs.io/en/latest/)

3) Generate Node 2 keys:
```
> go-genesis.exe generateKeys --config=node\config.toml
```

4) Initialize the database:
```
> go-genesis.exe initDatabase --config=node\config.toml
```

5) Start Node 2 services:
```
> centrifugo.exe -a 10.10.99.2 --config=config.json 
> go-genesis.exe start --config=node\config.toml
```

Ignore the showed errors. If you start node with log level "INFO", you'll see that node starts downloading blocks.


#### Adding keys

Errors that occurred above are caused by untrusted relationships between nodes. To fix it, add the second node public key to the first node.

To add keys, download this script [updateKeys.py](https://github.com/GenesisKernel/genesis-tests/blob/master/scripts/updateKeys.py). All information that is needed for  script execution is located in node's directory. This scipt must be executed on the first node with founder's privileges. 

Execute the script with the following arguments:

```
> py updateKeys.py PrivateKey1 Host1 Port1 KeyID2 PublicKey2 balance
```

Where:
-	PrivateKey1 - founder private key, located in the file PrivateKey of the first node
-	Host1 - IP-addres or DNS-name of the first node
-	Port1 - the first node API-server port
-	KeyID2 - content of file KeyID of the second node
-	PublicKey2 - content of file PublicKey of the second node
-	balance - set wallet balance of the second node

**Example:**

```
>py updateKeys.py 0f1aaf0c76716f189a295a0edbeed05ae760c4cd0009bd337f19aea6a0d37d89 10.10.99.1 7079 839301472950762263 2ce37e8a3fbeadd3862e962267fa29c43c02b6d2fbab9360f7d2e988e1477c333aa06fd6d85a8999779f1314063bf2bd2a298ea1284d0284b1c1ea69870d3ba 1000000000000000000000
```

All used network settings (IP-adresses and ports) are described in [Overview](#overview).

This script will create a contract, which add the second node public key to the table 'keys' of the database.

#### Create connection between nodes

Next, you must create connection between nodes. For this, you should download this script [newValToFullNodes.py](https://github.com/GenesisKernel/genesis-tests/blob/master/scripts/newValToFullNodes.py). All information that is needed for  script execution is located in node's directory. This script must be executed on the first node with founder's privileges. 

Execute the script with the following arguments:

```
> py newValToFullNodes.py PrivateKey1 Host1 Port1 "NewValue"
```

Where:
-	PrivateKey1 - founder private key, located in the file PrivateKey of the first node
-	Host1 - IP-addres or DNS-name of the first node
-	Port1 - the first node API-server port
-	NewValue - new value of Full_Nodes parameter

Argument **NewValue** must be written in json format:

```json
[
 {
  "tcp_address":"Host1:tcpPort1", 
  "api_address":"http://Host1:httpPort1", 
  "key_id":"KeyID1", 
  "public_key":"NodePubKey1"
 },
 {
  "tcp_address":"Host2:tcpPort2", 
  "api_address":"http://Host2:httpPort2", 
  "key_id":"KeyID2", 
  "public_key":"NodePubKey2"
 },
 {
  "tcp_address":"HostN:tcpPortN", 
  "api_address":"http://HostN:httpPortN", 
  "key_id":"KeyIDN", 
  "public_key":"NodePubKeyN"
 }
]
```

Where:

- Host1 - IP-addres or DNS-name of the first node
- tcpPort1 - the first node TCP-server port
- httpPort1 - the first node API-server port
- KeyID1 - content of file KeyID of the first node
- NodePubKey1 - content of file NodePublicKey of the first node
- Host2 - IP-addres or DNS-name of the second node
- tcpPort2 - the second node TCP-server port
- httpPort2 - the second node API-server port
- KeyID2 - content of file KeyID of the second node
- NodePubKey2 - content of file NodePublicKey of the second node
- HostN - IP-addres or DNS-name of node N
- tcpPortN - node N TCP-server port
- httpPortN - node N API-server port
- KeyIDN - content of file KeyID of node N
- NodePubKeyN - content of file NodePublicKey of node N

**Example:**

```
>py updateFullNode.py 0f1aaf0c76716f189a295a0edbeed05ae760c4cd0009bd337f19aea6a0d37d89 10.10.99.1 7079 "[{\"tcp_address\":\"10.10.99.1:7078\",\"api_address\":\"http://10.10.99.1:7079\",\"key_id\":\"4053339477525839986\",\"public_key\":\"d708bda3734e17822245d6477810ff28c150380abb9ae0271c5a49eca05be92fa7d80d0043e476ef936971288dd5df08b83370488182f524d789b919b398e70b\"},{\"tcp_address\":\"10.10.99.2:7078\",\"api_address\":\"http://10.10.99.2:7079\",\"key_id\":\"1647233376862283221\",\"public_key\":\"ed1bd2a29f607ca5529b353e8d6a9d998ebaca30a129a775b738b1bb0da4ff8afde9c6abdf208fd41fa4541cec471417705c7786ab69c359d36d822e8840815a\"},{\"tcp_address\":\"10.10.99.3:7078\",\"api_address\":\"http://10.10.99.3:7079\",\"key_id\":\"5700268145718545990\",\"public_key\":\"f653dd110e19abc259865397f14b1d866215fc4ea9abcaa246e620e750a3e26a46c5565bd6f2d19b4f18af10ccd8088bc62b7b2687b869bd542e91f2203ec164\"}]"
```

All used network settings (IP-adresses and ports) are described in [Overview](#overview).

Now, all nodes are connected to each other.

## ***Frontend Install*** <a name="frontend-install"></a>

Molis client is the frontend for Apla Blockchain Platform.

First, to work with the system, you must build Molis client, then you can deploy it to your users.

Molis client can be built in three technical implementations:

-	Desktop Application
-	Web Application
-	Mobile Application

## Frontend Install for Debian <a name="frontend-install-deb"></a>

Molis client can be build by the yarn package manager only on Debian 9 (Stretch) 64-bit [official distributive](https://www.debian.org/CD/http-ftp/#stable) with **installed GNOME GUI**.

### Frontend Software Prerequisites <a name="frontend-software-prerequisites-deb"></a>

#### Install Node.js

1) Download Node.js LTS version 8.11 from the [official site](https://nodejs.org/en/download/) or via the command line:

```bash
curl -sL https://deb.nodesource.com/setup_8.x | sudo -E bash
```

2) Install Node.js:

```bash
sudo apt install -y nodejs
```


#### Install Yarn

1) Download Yarn version 1.7.0 from [GitHub](https://github.com/yarnpkg/yarn/releases) or via command line:

```bash
$ cd /opt/apla &&  wget https://github.com/yarnpkg/yarn/releases/download/v1.7.0/yarn_1.7.0_all.deb
```

2) Install Yarn:

```bash
$ sudo dpkg -i yarn_1.7.0_all.deb && rm yarn_1.7.0_all.deb
```

### Build Molis App <a name="build-molis-app-deb"></a>

1) Download latest release of Molis from [GitHub](https://github.com/GenesisKernel/genesis-front) via git:

```bash
git clone https://github.com/GenesisKernel/genesis-front.git
```

2) Install Apla-Front dependencies via Yarn:

```bash
cd genesis-front/ && yarn install
```


#### Build Molis Desktop App

1) Create settings.json file that contains connections information about full nodes:

```bash
cp public/settings.json.dist public/settings.json
```

2) Edit settings.json file in any text editor and add required settings in this format:

```
http://Node_IP-address:Node_HTTP-Port
```

**Example** settings.json file for three nodes:

```json
{
    "fullNodes": [
        "http://10.10.99.1:7079",
        "http://10.10.99.2:7079",
        "http://10.10.99.3:7079"
    ]
}
```
All used network settings (IP-adresses and ports) are described in [Overview](#overview).

3) Build the desktop app with Yarn:

```bash
$ cd /opt/apla/apla-front && yarn build-desktop
```

4) Then desktop app must be packed to the AppImage:

```bash
$ yarn release --publish never -l
```

After that, your application will be ready to use, but its connection settings can not be changed in the future. If these settings will change, you must build a new version of the application.


#### Build Molis Web App

1) Create settings.json file as it is described in Build Molis Desktop App section.


2) Build the web app:

```bash
$ cd /opt/apla/apla-front/ && yarn build
```

After building, redistributable files will be placed to the '/build' directory. You can serve it with any web-server of your choice. Settings.json file must be also placed there. Note that you do not need to build your application again if your connection settings will change. Instead, edit the settings.json file and restart web-server.

2a) For development or testing purposes you can build Yarn's web-server:

```bash
$ sudo yarn global add serve && serve -s build
```

After this, your Molis Web App will be available at: ``http://localhost:5000``

## Frontend Install for Windows <a name="frontend-install-win"></a>

### Frontend Software Prerequisites <a name="frontend-software-prerequisites-win"></a>

#### Install Node.js

1) Download Node.js LTS version 8.11 for Windows from the [official site](https://nodejs.org/en/download/).

2) Install Node.js without any specific settings. All required environment variables will be installed during installation process.

#### Install Yarn

1) Download Yarn version 1.7.0 msi package from the [official site](https://yarnpkg.com/lang/en/docs/install/#windows-stable).

2) Install Yarn msi package without any specific settings. All required environment variables will be installed during the installation process.

### Build Molis <a name="build-molis-app-win"></a>

1) Download the latest release of Molis from [GitHub](https://github.com/GenesisKernel/genesis-front) via git:

```
> git clone https://github.com/GenesisKernel/genesis-front.git
```

2) Install Molis dependencies via Yarn:

```
> cd genesis-front
> yarn install
```

#### Build Molis Desktop App

1) Create a settings.json file in apla-front/public directory, which contains connections information about full nodes.

2) Edit settings.json file in any text editor and add the required settings in this format:

```
http://Node_IP-address:Node_HTTP-Port
```

**Example** 'settings.json' file for three nodes:

```json
{
    "fullNodes": [
        "http://10.10.99.1:7079",
        "http://10.10.99.2:7079",
        "http://10.10.99.3:7079"
    ]
}
```
All used network settings (IP-adresses and ports) are described in [Overview](#overview).

3) Build desktop app with Yarn:

```
> yarn build-desktop
```

4) Release your build for Windows OS:

```
> yarn release --publish never -w
```

After that, your application will be ready to execute at genesis-front\releases directory, but its connection settings can not be changed in the future. If these settings will change, you must build a new version of the application.

#### Build Molis Web App

1) Create settings.json file as it is described in Build Molis Desktop App section.

2) Build the web app:

```
> yarn build
```

After building, redistributable files will be placed to the '/build' directory. You can serve it with any web-server of your choice. Settings.json file must be also placed there. It is worth noting that you do not need to build your application again if your connection settings will change. Just edit settings.json file and restart web-server.

2')For development or testing purposes you can build Yarn's web-server:

```
> yarn global add serve
> serve -s build
```

After this, your Molis Web App will be available at http://localhost:3000

## Launching <a name="launching"></a>

After building Molis, you can obtain access to the system by the selected user.

### Login as Founder

To obtain system administrator rights on your ecosystem, you must login as ecosystem founder (Node 1 founder). To do this, first, you must obtain the private key of the founder that was generated during the installation of the node. This key is stored in the 'PrivateKey' file, located in node configuration directory.

Next, in the Molis client, in account options you must choose 'Import existing key' item.

In the next window, Import account, you should copy your founder's private key in field of account seed and set new founder password.

Now, in accounts list you can see your founder's account.

### Create a wallet

To create a wallet, in the Molis client, in account options you must choose item 'Generate new key'.

Invent a new account seed phrase or generate it and set a new user password.

Now, in accounts list you can see your user's account.
