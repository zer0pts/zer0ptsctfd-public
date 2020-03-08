#!/bin/bash
FLAG="ctf4y{JUST_DO_IT}"
password=$(cat /dev/urandom | tr -d -C A-Za-z0-9 | head -c 32)
rm -rf distfiles challenge
cp -r base distfiles
cp -r base challenge
sed -i challenge/html/index.php -e "s/<flag>/${FLAG}/"
sed -i challenge/sql/1_init.sql -e "s/<password>/${password}/" 
