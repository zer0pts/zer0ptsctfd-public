# scoreserver (backend)

Go製。スケールすると思って書いた。MySQL, RedisをDBとして使っている

## launch (for developping / debuggging)

色々環境変数で渡す必要があるけど、dockerで提供できるもの（MySQL, Redis, ...）に関してはMakefileの中にもう書いてある

```
$ make up
$ export WEHBHOOK=https://hooks.slack.com/services/XXXXXX
$ export EMAIL=smtp.gmail.com:587/XXXXXXXX@gmail.com/XXXXXXXX
$ make run
```

## initialize DB (for developping / debugging)

```
$ make reset
```

`database/reset.sql` `database/init.sql` を実行しているのと Redisの中身を吹き飛ばしている

## register challenges

```
$ make challenge-registerer
```

とすると `./bin/challenge-registerer` というバイナリが生成される。こいつはGitリポジトリとかファイルシステムから問題データを読んできてDBに放り込む。

```
$ make set-challenges
```

は `../challenges/` 以下の問題を追加する
