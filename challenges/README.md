# challenges

問題ファイルのサンプル。

## rules

- `distfiles/` 以下のファイル、 `distarchive.tar.gz` が配布ファイルとして扱われる
- 問題のディレクトリごと問題サーバにコピーされて、 `docker-compose up -d --build` が実行される

## challenge.json

host/port以外はrequired

- `is_dynamic`: falseにすると静的配点になる
- `is_questionary`: true にするとこの問題の提出時刻は最終提出時刻にならなくなる
- `difficulty`: 文字列


```yaml
servers:
  ?: &pwn_host localhost
  ?: &crypt_host localhost
  ?: &web_host localhost
challenges:
  "rsa":
    description: 'rsa challenge!!!! <pre>nc {{.Host}} {{.Port}}</pre>'
    flag: zer0pts{c0mm0n_m0dulu5_4tt4ck}
    category: crypto
    tags: [cma]
    author: yoshiking
    base_score: 1000
    difficulty: "easy"
    is_dynamic: true
    is_questionary: false
    host: *crypt_host
    port: 11000

  "Just Login":
    description: '<a href="http://{{.Host}}:{{.Port}}">DO IT</a>'
    flag: zer0pts{JUST_DO_IT}
    category: web
    tags: []
    author: theoremoon
    base_score: 1000
    difficulty: "medium"
    is_dynamic: true
    is_questionary: false
    host: *web_host
    port: 11000

  "Just Get Password":
    description: '<a href="http://{{.Host}}:{{.Port}}">Just get the password for admin</a>'
    flag: ctf4y{KEEN_YOUR_SQL_SKILL}
    category: web
    tags: []
    author: theoremoon
    base_score: 1000
    difficulty: "hard"
    is_dynamic: true
    is_questionary: false
    host: *web_host
    port: 11000

  "Questionary":
    description: "the flag is zer0pts{thank_u_4_playing}"
    flag: "zer0pts{thank_u_4_playing}"
    category: questionary
    tags: []
    author: zer0pts
    base_score: 1000
    difficulty: "questionary"
    is_dynamic: true
    is_questionary: true
```
