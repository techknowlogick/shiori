There are several installation methods available :

- [Using precompiled binary](#using-precompiled-binary)
- [Building from source](#building-from-source)
- [Using Docker image](#using-docker-image)

## Using Precompiled Binary

Download the latest version of `shiori` from [the release page](https://github.com/techknowlogick/shiori/releases/latest), then put it in your `PATH`. 

On Linux or MacOS, you can do it by adding this line to your profile file (either `$HOME/.bash_profile` or `$HOME/.profile`):

```
export PATH=$PATH:/path/to/shiori
```

Note that this will not automatically update your path for the remainder of the session. To do this, you should run:

```
source $HOME/.bash_profile
or
source $HOME/.profile
```

On Windows, you can simply set the `PATH` by using the advanced system settings.

## Building From Source

Make sure you have `go >= 1.12` installed, then run :

```
go get -u -d src.techknowlogick.com/shiori
cd $GOPATH/src/src.techknowlogick.com/shiori
GO111MODULE=on make dep build
```

## Using Docker Image

To use Docker image, you can pull the latest automated build from Docker Hub :

```
docker pull techknowlogick/shiori
```

If you want to build the Docker image on your own, Shiori already has its [Dockerfile](https://github.com/techknowlogick/shiori/blob/master/Dockerfile), so you can build the Docker image by running :

```
docker build -t shiori .
```
