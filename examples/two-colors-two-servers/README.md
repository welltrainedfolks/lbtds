# Two colors and two servers

This example provides configuration for LBTDS for loadbalancing two colors with two servers, as well as example program which will start four servers on different ports. Each server will return string like:

```
Color: green, backend#1
```

## Execute

Start four HTTP servers:

```
go run main.go
```

Open another console and start LBTDS:

```
go run ../../main.go
```

Try it (in third console):

```
curl -H "Host: web2.host" http://127.0.0.1:8200
```

Change color:

```
curl -H "Content-Type: application/json; charset=UTF-8" -d '{"color": "blue"}' -X POST http://127.0.0.1:4800/api/v1/color
```

Change "blue" to "green" and retry curl to 8200.