# Docker
Build docker image for *keyhole*.

## Docker Build

```
$ docker build -t simagix/keyhole build

REPOSITORY                         TAG                 IMAGE ID            CREATED             SIZE
simagix/keyhole                    latest              1fc381cbafd7        14 minutes ago      9.75MB

$ docker push simagix/keyhole
```

Image `simagix/keyhole` is also available from [Docker hub](https://hub.docker.com/).

### Lightweight Dockerfile
The image file is less than 10MB.

```
FROM alpine
MAINTAINER Ken Chen <simagix@gmail.com>
ADD build/keyhole-linux-x64 /keyhole
CMD ["/keyhole", "--version"]
```

## Docker Commands
### Check Version

```
$ docker run simagix/keyhole
keyhole ver. master-20180528.1527529455
```

### Get Info
Connect to an instance on the Docker host.

```
docker run simagix/keyhole/keyhole --info mongodb://$(hostname -f):30000/ 
```

### Atlas Example
```
docker run -v /etc/ssl/certs:/etc/ssl/certs simagix/keyhole \
    /keyhole --info "mongodb+srv://root:happy123@argos-jgtm2.mongodb.net/test"
```
