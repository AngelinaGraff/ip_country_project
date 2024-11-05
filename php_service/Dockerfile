FROM php:8.1-apache

RUN apt-get update && apt-get install -y \
    libzip-dev \
    unzip \
    && docker-php-ext-install zip

RUN pecl install redis && docker-php-ext-enable redis

COPY --from=composer:2 /usr/bin/composer /usr/bin/composer

COPY composer.json /var/www/html/
COPY composer.lock /var/www/html/
RUN composer install --no-dev --prefer-dist --optimize-autoloader

COPY . /var/www/html/

RUN chown -R www-data:www-data /var/www/html

RUN a2enmod rewrite

RUN echo "log_errors = On" >> /usr/local/etc/php/conf.d/docker-php-logging.ini \
    && echo "error_log = /dev/stderr" >> /usr/local/etc/php/conf.d/docker-php-logging.ini

EXPOSE 80

CMD ["apache2-foreground"]
