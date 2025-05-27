up:
	docker compose up -d
down:
	docker compose down
restart:
	docker compose restart cl_app
build:
	docker compose build
rebuild:
	docker compose build --no-cache
force-recreate-all:
	docker compose up -d --force-recreate
force-recreate-app:
	docker compose up -d --force-recreate cl_app
app:
	docker compose exec -it cl_app bash
logs:
	docker compose logs -f cl_app
app_logs:
	docker compose exec -it cl_app tail -f logs/app.log

create-network:
	docker network create cl-network