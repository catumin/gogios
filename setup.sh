#!/bin/bash

cd CheckEngine
go build
mv CheckEngine ../Gogios/usr/bin/servicecheck

cd ..
dpkg-deb --build Gogios
mv Gogios.deb gogios.deb

cp gogios.deb ArchBuild/
cd ArchBuild

check=$(sha256sum gogios.deb | awk '{print $1}')
sed -i "/sha256sums/s/.*/sha256sums=(\'$check\')/" PKGBUILD
makepkg --printsrcinfo > .SRCINFO
makepkg
mv *.tar.xz ../
