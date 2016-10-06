#! /bin/bash -e
ARCH0=""
ARCH1="32"
if [ $# -gt 0 ] && [ $1 = "amd64" ]
then
  ARCH0="64"
  ARCH1="64"
fi
cd make_deb
chown root daylight$ARCH0/usr/share/daylight/daylight
chgrp root daylight$ARCH0/usr/share/daylight/daylight
chmod 0777 daylight$ARCH0/usr/share/daylight/daylight
dpkg-deb --build daylight$ARCH0
zip -j daylight_linux$ARCH1.zip daylight$ARCH0/usr/share/daylight/daylight
mv daylight$ARCH0.deb daylight_linux$ARCH1.deb
rm -rf daylight$ARCH0/usr/share/daylight/daylight