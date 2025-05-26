#!/bin/bash

chmod -R o+rw bootstrap storage

chown -R www-data:www-data \
        /var/www/storage \
        /var/www/bootstrap/cache

composer install
php artisan migrate
php artisan db:seed
php-fpm