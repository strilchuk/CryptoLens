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

# Multi-arch build commands
build-amd64:
	docker buildx build --platform linux/amd64 -t cl_s_app:amd64 -f docker/app/Dockerfile .

build-arm7:
	docker buildx build --platform linux/arm/v7 -t cl_s_app:arm7 -f docker/app/Dockerfile .

build-arm64:
	docker buildx build --platform linux/arm64 -t cl_s_app:arm64 -f docker/app/Dockerfile .

build-all:
	docker buildx build --platform linux/amd64,linux/arm/v7,linux/arm64 -t cl_s_app:multi-arch -f docker/app/Dockerfile .