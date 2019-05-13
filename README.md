# Gogios
Simple system to check important services are on remote machines

# Install on Ubuntu/Debian
Download and install the latest release deb file and then start the service. For example:

```
wget https://github.com/BKasin/Gogios/releases/download/v1.3/gogios-1.3.deb
sudo dpkg -i gogios-1.3.deb
sudo systemctl start gogios
sudo systemctl enable gogios
```

You will need to install a webserver and have it point to

```
/opt/gingertechengine
```

as its webroot.

The file that checks are pulled from is in /etc/gingertechengine, as well as an example nginx website config file.

After installing Gogios, you can configure it at /etc/gingertechengine/gogios.toml.
