all: clean build play

build: build_network build_restful build_checkproxy build_db build_cache build_mq build_config build_pushmsg build_save build_scanproxy build_scanport build_proxyscan
play: play_mq play_db play_cache play_config play_pushmsg play_save play_restful play_checkproxy
clean:
	docker stop mq-1 \
					restful-1 \
    				db-1 \
    				cache-1 \
    				scanproxy-1 \
    				scanproxy-2 \
    				scanproxy-3 \
					scanproxy-4 \
    				scanproxy-5 \
					scanproxy-6 \
                    scanproxy-7 \
                    scanproxy-8 \
                    scanproxy-9 \
                    scanproxy-0 \
                    scanport-1 \
                    scanport-2 \
                    scanport-3 \
                    scanport-4 \
                    scanport-5 \
                    scanport-6 \
                    scanport-7 \
                    scanport-8 \
                    scanport-9 \
                    scanport-0 \
                    proxyscan-1 \
                    proxyscan-2 \
                    proxyscan-3 \
                    proxyscan-4 \
                    proxyscan-5 \
                    proxyscan-6 \
                    proxyscan-7 \
                    proxyscan-8 \
                    proxyscan-9 \
                    proxyscan-0 \
    				save-1 \
    				config-1 \
    				pushmsg-1 | xargs docker rm || exit 0

build_network:
	docker network create -d bridge searchproxy || exit 0
build_db:
	cd db;docker build -t searchdb .;cd -;
build_cache:
	cd cache;docker build -t searchcache .;cd -;
build_mq:
	cd mq;docker build -t searchmq .;cd -;
build_restful:
	cp -rf ../bin/restful restful/
	cd restful;docker build -t restful .;cd -;
	rm -rf restful/restful
build_checkproxy:
	cp -rf ../bin/checkproxy checkproxy/
	cd checkproxy;docker build -t checkproxy .;cd -;
	rm -rf checkproxy/checkproxy
build_config:
	cp -rf ../bin/config config/
	cd config;docker build -t config .;cd -;
	rm -rf config/config
build_pushmsg:
	cp -rf ../bin/pushmsg pushmsg/
	cp -rf ../bin/config pushmsg/
	cd pushmsg;docker build -t pushmsg .;cd -;
	rm -rf pushmsg/pushmsg pushmsg/config.json
build_scanproxy:
	cp -rf ../bin/task scanproxy/
	cp -rf ../bin/scanproxy scanproxy/
	cp -rf ../bin/config.json scanproxy/
	cd scanproxy;docker build -t scanproxy .;cd -;
	rm -rf scanproxy/scanproxy scanproxy/task scanproxy/config.json
build_scanport:
	cp -rf ../bin/task scanport/
	cp -rf ../bin/scanport scanport/
	cp -rf ../bin/config.json scanport/
	cd scanport;docker build -t scanport .;cd -;
	rm -rf scanport/scanport scanport/task scanport/config.json
build_proxyscan:
	cp -rf ../bin/task proxyscan/
	cp -rf ../bin/proxyscan proxyscan/
	cp -rf ../bin/config.json proxyscan/
	cd proxyscan;docker build -t proxyscan .;cd -;
	rm -rf proxyscan/proxyscan proxyscan/task proxyscan/config.json
build_save:
	cp -rf ../bin/task save/
	cp -rf ../bin/save save/
	cp -rf ../bin/config.json save/
	cd save;docker build -t save .;cd -;
	rm -rf save/task save/save save/config.json
play_config:
	docker run --network searchproxy --name config-1 --restart=always -d config || exit 0
play_db:
	docker run --network searchproxy --name db-1 --restart=always -d searchdb || exit 0
play_mq:
	docker run --network searchproxy --name mq-1 --restart=always -p 15672:15672 -p 5672:5672 -d searchmq || exit 0
play_cache:
	docker run --network searchproxy --name cache-1 --restart=always -d searchcache || exit 0
play_restful:
	docker run --network searchproxy --name restful-1 -p 8080:8080 --restart=always -d restful || exit 0
play_checkproxy:
	docker run --network searchproxy --name checkproxy-1 --restart=always -d checkproxy || exit 0
play_scanport:
	docker run --network searchproxy --name scanport-1 --restart=always -d scanport || exit 0
	docker run --network searchproxy --name scanport-2 --restart=always -d scanport || exit 0
	docker run --network searchproxy --name scanport-3 --restart=always -d scanport || exit 0
play_proxyscan:
	docker run --network searchproxy --name proxyscan-1 --restart=always -d proxyscan || exit 0
	docker run --network searchproxy --name proxyscan-2 --restart=always -d proxyscan || exit 0
	docker run --network searchproxy --name proxyscan-3 --restart=always -d proxyscan || exit 0
play_scanproxy:
	docker run --network searchproxy --name scanproxy-1 --restart=always -d scanproxy || exit 0
	docker run --network searchproxy --name scanproxy-2 --restart=always -d scanproxy || exit 0
	docker run --network searchproxy --name scanproxy-3 --restart=always -d scanproxy || exit 0
play_localscan:
	docker run --name scanproxy-1 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-2 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-3 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-4 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-5 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-6 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-7 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-8 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-9 --restart=always -e MOD=0 -d scanproxy || exit 0
	docker run --name scanproxy-0 --restart=always -e MOD=0 -d scanproxy || exit 0
play_localport:
	docker run --name scanport-1 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-2 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-3 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-4 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-5 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-6 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-7 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-8 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-9 --restart=always -e MOD=0 -d scanport || exit 0
	docker run --name scanport-0 --restart=always -e MOD=0 -d scanport || exit 0
play_localproxy:
	docker run --name proxyscan-1 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-2 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-3 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-4 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-5 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-6 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-7 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-8 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-9 --restart=always -e MOD=0 -d proxyscan || exit 0
	docker run --name proxyscan-0 --restart=always -e MOD=0 -d proxyscan || exit 0
play_save:
	docker run --network searchproxy --name save-1 --restart=always -d save || exit 0
play_pushmsg:
	docker run --network searchproxy --name pushmsg-1 --restart=always -d pushmsg || exit 0