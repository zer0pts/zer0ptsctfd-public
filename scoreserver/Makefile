up:
	(cd dev; docker-compose up)

down:
	(cd dev;\
		docker-compose down)

build:
	go build

run: build
	FRONT="http://front.web.localhost:8080" REDIS='localhost:6379' DBDSN='zer0ptsuser:zer0ptspassword@tcp(localhost:13306)/zer0pts' ./scoreserver

test: reset build
	go test ./...


reset:
	(cd dev; docker-compose exec -T db mysql -u zer0ptsuser -pzer0ptspassword zer0pts < ../database/reset.sql || true)
	(cd dev; docker-compose exec -T db mysql -u zer0ptsuser -pzer0ptspassword zer0pts < ../database/schema.sql)
	(cd dev; docker-compose exec -T db mysql -u zer0ptsuser -pzer0ptspassword zer0pts < ../database/init.sql)
	(cd dev; docker-compose exec redis redis-cli flushall)

sql:
	(cd dev; docker-compose exec db mysql -u zer0ptsuser -pzer0ptspassword zer0pts)

challenge-registerer:
	(cd cmd/challenge-registerer; go build -a -tags netgo -installsuffix netgo -ldflags="-extldflags \"-static\"" -o ../../bin/challenge-registerer)

set-challeges:
	(cd bin; REDIS='localhost:6379' DBDSN='zer0ptsuser:zer0ptspassword@tcp(localhost:13306)/zer0pts' ./challenge-registerer --dir "../../challenges" -uploader "http://transfer:password@localhost:9999/")

production-build:
	go build -a -tags netgo -installsuffix netgo -ldflags="-extldflags \"-static\""


docker: production-build
	docker build . -t zer0ptsctfd
