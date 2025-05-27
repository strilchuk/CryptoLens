.PHONY: help
help:
	@grep -E '^[1-9a-zA-Z_-]+:.*?## .*$$|(^#--)' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[32m %-43s\033[0m %s\n", $$1, $$2}' \
    | sed -e 's/\[32m #-- /[33m/'

#-- Application
update: ## Update everything, including sources from git
	make dc_down
	git fetch
	git pull
	make dc_build
	make dc_up

#-- Docker
dc_ps: ## Apps list
	docker compose ps
dc_up: ## Start Apps
	docker compose up -d
dc_down: ## Stop Apps
	docker compose down
dc_restart: ## Restart Apps
	make dc_down
	make dc_up
dc_build: ## Rebuild Apps
	docker compose build

#-- Database - oa_monitor_db
db_up: ## Database start
	docker compose up -d oa_monitor_db
db_stop: ## Database stop
	docker compose stop oa_monitor_db
db_rm: ## Delete database container
	docker compose rm oa_monitor_db
db_restart: ## Database restart
	make -s db_stop
	make -s db_rm
	make -s db_up
db_logs: ## Database logs
	docker compose logs -f oa_monitor_db

#-- Grafana - oa_monitor_grafana
gf_up: ## Grafana start
	docker compose up -d oa_monitor_grafana
gf_stop: ## Grafana stop
	docker compose stop oa_monitor_grafana
gf_rm: ## Deleta Grafana container
	docker compose rm oa_monitor_grafana
gf_restart: ## Grafana restart
	make -s gf_stop
	make -s gf_rm
	make -s gf_up
gf_logs: ## Grafana logs
	docker compose logs -f oa_monitor_grafana

#-- Prometheus - oa_monitor_prometheus
pt_up: ## Prometheus start
	docker compose up -d oa_monitor_prometheus
pt_stop: ## Prometheus stop
	docker compose stop oa_monitor_prometheus
pt_rm: ## Deleta Prometheus container
	docker compose rm oa_monitor_prometheus
pt_restart: ## Prometheus restart
	make -s gf_stop
	make -s gf_rm
	make -s gf_up
pt_logs: ## Prometheus logs
	docker compose logs -f oa_monitor_prometheus

#-- Runner - oa_monitor_runner
rr_up: ## Runner start
	docker compose up -d oa_monitor_runner
rr_stop: ## Runner stop
	docker compose stop oa_monitor_runner
rr_rm: ## Delete Runner container
	docker compose rm oa_monitor_runner
rr_restart: ## Runner restart
	make -s rr_stop
	make -s rr_rm
	make -s rr_up
rr_logs: ## Runner logs
	docker compose logs -f oa_monitor_runner

create-network:
	docker network create cl-network