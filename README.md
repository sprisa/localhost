# localhost

Local development certs for LAN https services.

## Install

### With Go
```sh
go install github.com/sprisa/localhost@latest
```

### With Homebrew
```sh
TODO
```

## Usage

Serve the service at port `3000`
```sh
localhost 3000
```

Serve on all interfaces
```sh
localhost 3000 -a
```

Change the proxy port. Defaults to port `5050`
```sh
localhost 3000 -p 3001
```


Show Help
```sh
localhost --help
```



## Inspiration

- localtls - https://github.com/Corollarium/localtls
- sslip.io - https://sslip.io/
