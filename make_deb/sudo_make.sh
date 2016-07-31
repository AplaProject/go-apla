#! /bin/bash -e
ARCH0=""
ARCH1="32"
if [ $# -gt 0 ] && [ $1 = "amd64" ]
then
  ARCH0="64"
  ARCH1="64"
fi
cd make_deb
chown root dcoin$ARCH0/usr/share/dcoin/dcoin
chgrp root dcoin$ARCH0/usr/share/dcoin/dcoin
chmod 0777 dcoin$ARCH0/usr/share/dcoin/dcoin
dpkg-deb --build dcoin$ARCH0
zip -j dcoin_linux$ARCH1.zip dcoin$ARCH0/usr/share/dcoin/dcoin
mv dcoin$ARCH0.deb dcoin_linux$ARCH1.deb
rm -rf dcoin$ARCH0/usr/share/dcoin/dcoin