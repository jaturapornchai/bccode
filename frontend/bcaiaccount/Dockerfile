FROM ghcr.io/cirruslabs/flutter:3.10.5 as builder

WORKDIR /app

RUN flutter config --enable-web

COPY . /app

RUN flutter clean
RUN flutter pub get

RUN flutter build web -t lib/main_vfmerchant.dart --dart-define=ENVIRONMENT=PROD --dart-define=FLAVOR=vfpos

FROM nginx:1.25-alpine-slim

ADD nginx_conf/nginx.conf /etc/nginx/nginx.conf

COPY --from=builder /app/build/web /usr/share/nginx/html
EXPOSE 80