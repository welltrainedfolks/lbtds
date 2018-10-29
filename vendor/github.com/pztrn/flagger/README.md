[![GoDoc](https://godoc.org/github.com/pztrn/flagger?status.svg)](https://godoc.org/github.com/pztrn/flagger) [![Build Status](https://travis-ci.org/pztrn/flagger.svg?branch=master)](https://travis-ci.org/pztrn/flagger)

# Flagger

Flagger is an arbitrary CLI flags parser, like argparse in Python.
Flagger is able to parse boolean, integer and string flags.

# Installation

```
go get -u -v lab.pztrn.name/golibs/flagger
```

# Usage

Flagger requires logging interface to be passed on initialization.
See ``loggerinterface.go`` for required logging functions.
It is able to run with standart log package, in that case
initialize flagger like:

```
flgr = flagger.New(flagger.LoggerInterface(log.New(os.Stdout, "testing logger: ", log.Lshortfile)))
flgr.Initialize()
```

Adding a flag is easy, just fill ``Flag`` structure and pass to ``AddFlag()`` call:

```
flag_bool := Flag{
    Name: "boolflag",
    Description: "Boolean flag",
    Type: "bool",
    DefaultValue: true,
}
err := flgr.AddFlag(&flag_bool)
if err != nil {
    ...
}
```

After adding all neccessary flags you should issue ``Parse()`` call to get
them parsed:

```
flgr.Parse()
```

After parsed they can be obtained everywhere you want, like:

```
val, err := flgr.GetBoolValue("boolflag")
if err != nil {
    ...
}
```

For more examples take a look at ``flagger_test.go`` file or [at GoDoc](https://github.com/pztrn/flagger).
