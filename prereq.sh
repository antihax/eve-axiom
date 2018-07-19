#!/bin/bash
mkdir prereq
cd prereq
wget https://github.com/antihax/libdogma/releases/download/latest/libdogma-latest.tar.xz
wget https://phoenixnap.dl.sourceforge.net/project/judy/judy/Judy-1.0.5/Judy-1.0.5.tar.gz

tar -xf libdogma-latest.tar.xz
tar -xf Judy-1.0.5.tar.gz

cd judy-1.0.5
./configure
make
sudo make install

cd ..
cd libdogma-1.2.1
./configure
make
sudo make install

cd ..
cd ..
