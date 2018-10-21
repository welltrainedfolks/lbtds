# LBTDS — Load Balancer That Doesn't Suck

At least we hope so.

## Our goals

Our goals is to provide as tiny as possible load balancer, compatible with "blue-green" deployment paradigm. Unlinke some other of load balancers, we have these features:

* Simple configuration
* Support for more than two colors (we support any color you want)
* Small run-time and binary size.

## Current status

This software is under development and in early stages. Some things may break, but mostly it is already good enough to try in production.

## A warning

Repository at Github is a mirror, all development takes place at https://source.hodakov.me/fat0troll/lbtds. Please, file bugs and PRs there.

About rationale of making this decision, please, read [here](https://source.hodakov.me/fat0troll/lbtds/src/branch/master/CONTRIBUTION.md).

## Installation

As simple as:

```
go get source.hodakov.me/fat0troll/lbtds
```

Do not try to go get from Github or you'll have plenty of errors!

## Configuration

Take a look at examples for generic configuration file. Documentation on that topic will be ready soon.

## Usage examples

See ``examples/`` folder of this repository.

## ToDo

* ACME (LetsEncrypt) support.
* TCP proxifying support.
* Tests and benchmarks.
* Statistics exporting (e.g. for Prometheus).
* ...maybe more, take a look at [issues page](https://source.hodakov.me/fat0troll/lbtds/issues).

## License

See LICENSE file for details.