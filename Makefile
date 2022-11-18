default: docker
docker:
	docker build -t francoismichel/build-speedtest-webtransport-go -f ./Dockerfile.build .
	docker run -t francoismichel/build-speedtest-webtransport-go /bin/true
	docker cp `docker ps -a -q -f "ancestor=francoismichel/build-speedtest-webtransport-go"`:/server-speedtest-webtransport-go .
	docker cp `docker ps -a -q -f "ancestor=francoismichel/build-speedtest-webtransport-go"`:/client-speedtest-webtransport-go .
	docker container rm `docker ps -a -q -f "ancestor=francoismichel/build-speedtest-webtransport-go"`
	chmod 0755 ./server-speedtest-webtransport-go ./client-speedtest-webtransport-go
	docker build --rm -t francoismichel/speedtest-webtransport-go -f ./Dockerfile.static .
