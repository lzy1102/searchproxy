all: clean build play

build: build_network build_db build_cache build_mq build_config build_pushmsg build_save build_scanproxy
play: play_mq play_db play_cache play_config play_pushmsg play_scanproxy play_save
clean:
	docker stop mq-1 \
    				db-1 \
    				cache-1 \
    				scanproxy-1 \
    				scanproxy-2 \
    				scanproxy-3 \
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
build_config:
	cp -rf ../bin/config config/
	cd config;docker build -t config .;cd -;
	rm -rf config/config
build_pushmsg:
	cp -rf ../bin/pushmsg pushmsg/
	cd pushmsg;docker build -t pushmsg .;cd -;
	rm -rf pushmsg/pushmsg
build_scanproxy:
	cp -rf ../bin/task scanproxy/
	cp -rf ../bin/scanproxy scanproxy/
	cd scanproxy;docker build -t scanproxy .;cd -;
	rm -rf scanproxy/scanproxy
	rm -rf scanproxy/task
build_save:
	cp -rf ../bin/task save/
	cp -rf ../bin/save save/
	cd save;docker build -t save .;cd -;
	rm -rf save/task
	rm -rf save/save
play_config:
	docker run --network searchproxy --name config-1 --restart=always -d config || exit 0
play_db:
	docker run --network searchproxy --name db-1 --restart=always -p 27017:27017 -d searchdb || exit 0
play_mq:
	docker run --network searchproxy --name mq-1 --restart=always -p 5672:5672 -p 15672:15672 -d searchmq || exit 0
play_cache:
	docker run --network searchproxy --name cache-1 --restart=always -d searchcache || exit 0
play_scanproxy:
	docker run --network searchproxy --name scanproxy-1 --restart=always -d scanproxy || exit 0
	docker run --network searchproxy --name scanproxy-2 --restart=always -d scanproxy || exit 0
	docker run --network searchproxy --name scanproxy-3 --restart=always -d scanproxy || exit 0
play_save:
	docker run --network searchproxy --name save-1 --restart=always -d save || exit 0
play_pushmsg:
	docker run --network searchproxy --name pushmsg-1 --restart=always -d pushmsg || exit 0