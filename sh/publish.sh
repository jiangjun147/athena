#!/bin/sh

#shunyou@192.168.199.125:/home/shunyou/bingo
git pull
rm -rf bin
./build.sh
tar -czvf output.tar.gz bin conf stay.sh restart.sh
scp ./output.tar.gz root@119.29.153.100:/opt/glc &
scp ./output.tar.gz root@119.29.111.54:/opt/glc &
#scp ./output.tar.gz root@106.55.246.11:/opt/glc &
#scp ./output.tar.gz root@42.194.162.240:/opt/glc &
wait