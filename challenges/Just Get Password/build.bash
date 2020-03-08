#!/bin/bash
FLAG="ctf4y{KEEN_YOUR_SQL_SKILL}"
rm -rf distfiles challenge
cp -r base distfiles
cp -r base challenge
sed -i challenge/sql/1_init.sql -e "s/<flag>/${FLAG}/"
