run:
	docker compose run --rm cl-app bash

up:
	docker compose -f docker-compose.yml up -d

down:
	docker compose -f docker-compose.yml down

restart:
	docker compose restart cl-app

force-recreate-all:
	docker compose -f docker-compose.yml up -d --force-recreate

force-recreate-app:
	docker compose -f docker-compose.yml up -d --force-recreate cl-app

app-bash:
	docker compose -f docker-compose.yml exec -it cl_app bash

build:
	docker compose -f docker-compose.yml build

rebuild:
	docker compose -f docker-compose.yml build --no-cache

create-network:
	docker network create cl-network