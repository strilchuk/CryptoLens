FROM php:8.3-fpm

RUN apt-get update && apt-get install -y \
    git \
    curl \
    libpng-dev \
    libonig-dev \
    libxml2-dev \
    zip \
    nano \
    unzip \
    locales \
    libmcrypt-dev \
    libpq-dev \
    libzip-dev \
    zlib1g-dev \
    libicu-dev \
    g++ \
    libmagickwand-dev \
    libxml2-dev \
    libzstd-dev \
    libssl-dev

# RUN apt-get update; \
#     apt-get install -y libmagickwand-dev; \
#     pecl install imagick; \
#     docker-php-ext-enable imagick;

RUN apt-get update; \
    apt-get install -y libmagickcore-dev; \
    apt-get install -y libmagickwand-dev; \
    pecl install imagick; \
    docker-php-ext-enable imagick;

RUN docker-php-ext-install zip

# Установка Redis
RUN pecl install redis && docker-php-ext-enable redis

RUN apt-get clean && rm -rf /var/lib/apt/lists/*

RUN pecl install xdebug && docker-php-ext-enable xdebug

RUN docker-php-ext-install pdo_mysql mbstring exif pcntl bcmath gd
RUN docker-php-ext-configure pgsql -with-pgsql=/usr/local/pgsql
RUN docker-php-ext-install pdo pdo_pgsql
RUN docker-php-ext-install opcache

COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

COPY ./docker/php_fpm/entrypoint.sh /etc/entrypoint.sh

WORKDIR /var/www

# Можно и так
#ENTRYPOINT ["/etc/entrypoint.sh"]
#CMD ["php-fpm"]

ENTRYPOINT ["bash", "/etc/entrypoint.sh"]
