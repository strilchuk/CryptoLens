docker network create oa_monitor_network
docker-compose up --build
docker run -it --rm orangeadmin_monitor-oa_monitor_runner sh