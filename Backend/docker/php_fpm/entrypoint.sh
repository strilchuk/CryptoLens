#!/bin/bash

chmod -R o+rw bootstrap storage

chown -R www-data:www-data \
        /var/www/storage \
        /var/www/bootstrap/cache

# Создаем директорию для логов если её нет
mkdir -p /var/www/storage/logs
chmod -R 777 /var/www/storage/logs
touch /var/www/storage/logs/laravel.log
chmod 666 /var/www/storage/logs/laravel.log
touch /var/www/storage/logs/scheduler.log
chmod 666 /var/www/storage/logs/scheduler.log

# Проверяем наличие vendor директории
if [ ! -d "vendor" ]; then
    echo "Installing dependencies..."
    composer install
fi

# Проверяем, изменился ли composer.json
#if [ -f "composer.json" ] && [ -f "composer.lock" ]; then
#    if [ "$(stat -c %Y composer.json)" -gt "$(stat -c %Y vendor)" ]; then
#        echo "composer.json was modified, updating dependencies..."
#        composer install
#    fi
#fi

php artisan migrate
php artisan db:seed

# Если команда не передана, используем php-fpm по умолчанию
if [ $# -eq 0 ]; then
    php-fpm
else
    exec "$@"
fi