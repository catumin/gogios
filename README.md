# Gogios

Simple system to check important services are on remote machines.

## Build Instructions

I'm working on moving package creation to an Open Build Service instance, so I can do it all at once
and have a way easier time. I also like to think that I have gotten slightly less foolish over time as I
progress through the development of this project.

Because of this, I'm rebuilding this repo to be more OBS compliant and a bit more usable in general.

### New Method of Building

I use go-dep to track project dependencies. They should all be included in the vendor/ folder, but if not
you can use:

```bash
dep ensure
```

to collect them all (in theory, go get -d ./... will do the same thing).

Once you have those, you can build the project with:

```bash
make build
```

Which will create a bin folder and put all the binaries in it. To install:

```bash
make install
```

This will make folders and put files where they need to be.

## Installing from Binary Packages

### Ubuntu/Debian

Download and install the latest release deb file and then start the service. For example:

```bash
wget https://github.com/BKasin/Gogios/releases/download/VERSION/gogios-VERSION.deb
sudo dpkg -i gogios-VERSION.deb
sudo systemctl start gogios
sudo systemctl enable gogios
```

### Arch

I have made an AUR package that can be installed with something like yay using:

```bash
yay -S gogios-bin
```

And then start and enable the service.

### All OSes

You will need to install a webserver and have it point to

```bash
/opt/gingertechengine
```

as its webroot.

The file that checks are pulled from is in /etc/gingertechengine, as well as an example nginx website config file.

After installing Gogios, you can configure it at /etc/gingertechengine/gogios.toml.
