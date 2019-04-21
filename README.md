# Gogios
Simple system to check important services are on remote machines

# Install on Ubuntu/Debian
```
sudo apt install golang-go apt-transport-https nginx
wget -q https://packages.microsoft.com/config/ubuntu/16.04/packages-microsoft-prod.deb
sudo dpkg -i packages-microsoft-prod.deb
sudo apt update
sudo apt install dotnet-sdk-2.2
```

And then download and install the latest release deb file. To start the services, run:

```
sudo systemctl start gogios
sudo systemctl start gogios-web
```

The file that checks are pulled from is in /etc/gingertechengine, as well as an example nginx website config file.
