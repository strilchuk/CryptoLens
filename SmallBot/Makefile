up:
	docker compose up -d
down:
	docker compose down
restart:
	docker compose restart cl_s_app
build:
	docker compose build
rebuild:
	docker compose build --no-cache
force-recreate:
	docker compose up -d --force-recreate
app:
	docker compose exec -it cl_s_app bash
logs:
	docker compose logs -f cl_s_app
app_logs:
	docker compose exec -it cl_s_app tail -f logs/app.log