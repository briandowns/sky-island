# Sky Island

<p align="left">
  <a href="https://godoc.org/github.com/briandowns/sky-island"><img src="https://godoc.org/github.com/briandowns/sky-island?status.svg?" alt="GoDoc"></a>
  <a href="https://opensource.org/licenses/BSD-3-Clause"><img src="https://img.shields.io/badge/License-BSD%203--Clause-orange.svg?" alt="License"></a>
  <a href="https://github.com/briandowns/sky-island/releases"><img src="https://img.shields.io/badge/version-0.0.0-green.svg?" alt="Version"></a>
</p>

**Experimental** / *ALPHA stage* 

## 

Sky Island is a FaaS/serverless platform built for FreeBSD, jail driven, on ZFS, for running Go functions, with interaction through a REST API.

## How It Works

A request comes in to run a function. The request contains a Github URL to a Go repository containing the function. The request also contains the "call".  The call is what will be run including the arguments necessary to run the function.

Upon successfully accepting the inbound request, the server will check if the repo has already been cloned and if not, it will clone it. From there, it will generate a "main.go" file and compile a binary in the "build" jail. The "build" jail holds all of the cloned repositories and will be reused on each request unless otherwise told not to.  Once a binary is created, an execution jail is created, the binary is copied into it, and is subsequently executed. The binary's output is then returned to the caller via an HTTP response to the original request.

### Examples

Simple Call
```
curl --silent -XPOST http://demo.skyisland.io:3281/api/v1/function -d '{"url": "github.com/mmcloughlin/geohash", "call": "Encode(100.1, 80.9)"}'
```

Cache Bust Call
```
curl --silent -XPOST http://demo.skyisland.io:3281/api/v1/function -d '{"url": "github.com/mmcloughlin/geohash", "call": "Encode(100.1, 80.9)", "cache_bust": true}'
```

Result
```
{"timestamp":1513717061,"data":"jcc92ytsf8kn"}
```

## Requirements

* lib32.txz installed
* ZFS
* Go version >= 1.9 
* Make sure that `jails_enabled="YES"` is present in the "/etc/rc.conf" file

## System Initialization

Initialzing the system does a number of things to make running Sky Island easier.  Sky Island will check to see if the base system packages and Go tarball have already been downloaded and if they have, they'll use those. 

* Create a ZFS dataset to work from
* Download the base package for the version of FreeBSD you have installed
* Extract those packages to the dataset where the base jail will be kept
* Update the base jail with `freebsd-update`
* Set some basic jail configuration
* Install Go and create a workspace
* Create a ZFS snapshot of the base jail
* Create `build` jail

This is accomplished by running: 

`sky-island -c config.json -i`

## Installation

`go install` will install the Sky Island binary into the Go bin directory in the GOPATH.  

The above can be adequate however for some folks, you might want to have Sky Island controlled through the RC system. An RC script is included as well as a target in the Makefile to install it.  `make install`

## Running Sky Island

To run Sky Island, run the command below.

`sky-island -c config.json` 

## IP Address Management

 The Sky Island config file has an IP4 section to configure how it handles jails IP addressing.  If a request is received that indicates a jail needs an IP address, Sky Island checks to see if there is an available address and returns one to be assigned to the execution jail. Use the admin API, described below, to manage the IP pool and to see which jail is associated with which IP and visa versa.

 The subnet that Sky Island exists on should have DHCP turned off or at a minimum, make sure that the IP pools aren't overlapping.

 There will be a future effort to support multiple IP4 pools.

## Caching

Sky Island tries it's best to respond to API requests as quickly as possible.  To achieve some level of speed, a number of caching mechanisms has been implemented for binaries and repositories.  Upon receiving a request via the API, Sky Island will check to see if there's an associated binary that's already been compiled. If there is, that artifact is used.  If there's no binary, Sky Island checks to see if the repository has been seen before and if so, uses the repo on disk and compiles a binary from there.  The binary will be added to the binary cache for later use.

This cache can be busted however by including `cache_bust=true` in payload of a "function run" POST request. This will force Sky Island to clone the repo and build a new binary.

## API

The Sky Island API provides insight into the Sky Island system. The healthcheck endpoint is not protected by header auth however the admin endpoints are. This can be configured by fields in the `config.json` file by setting the 'admin_api_token' and 'admin_token_header' fields.

| Method | Resource                    | Description                                                            |
| :----- | :-------                    | :----------                                                            |
| GET    | /api/v1/healthcheck         | Verifies the service is up and running                                 | 
| POST   | /api/v1/function            | Endpoint that receives function run requests                           |
| GET    | /api/v1/admin/api-stats     | API statistics                                                         | 
| GET    | /api/v1/admin/jails         | Get a list of the running jails                                        |
| GET    | /api/v1/admin/jail/{id}     | Get the details for the given jail                                     |
| DELETE | /api/v1/admin/jail/{id}     | Kill the jail with the given ID                                        |
| DELETE | /api/v1/admin/jails         | Kill all jails                                                         |
| GET    | /api/v1/admin/ips           | Get a list of IP's filtered by param. `?state={available|unavailable}` |
| PUT    | /api/v1/admin/ips           | Update the state of a given IP                                         |

## Metrics

By default, Sky Island uses StatsD to write out metrics. Jail created/removed counts, requests times, etc are reported.

## Contact

Brian Downs [@bdowns328](http://twitter.com/bdowns328)

## License

Sky Island source code is available under the BSD 3 Clause [License](/LICENSE).

## 

![alt text](https://www.freebsd.org/gifs/powerani.gif)
