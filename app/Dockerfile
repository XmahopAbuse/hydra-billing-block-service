FROM golang:1.18-buster as builder

WORKDIR /app

# Копируем go.* файлы для установки зависимостей
COPY go.* ./
RUN go mod download

# Копируем код
COPY . ./

# Собираем бинарник
RUN go build -v -o app

# Копируем из билдера в отдельный контейнер
FROM debian:buster-slim

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y \
         ca-certificates

RUN apt-get install alien libaio1 wget -y  --no-install-recommends
RUN wget https://download.oracle.com/otn_software/linux/instantclient/185000/oracle-instantclient18.5-basiclite-18.5.0.0.0-3.x86_64.rpm && \
    wget https://download.oracle.com/otn_software/linux/instantclient/185000/oracle-instantclient18.5-devel-18.5.0.0.0-3.x86_64.rpm
RUN alien -i oracle-instantclient18.5-basiclite-18.5.0.0.0-3.x86_64.rpm && \
    alien -i oracle-instantclient18.5-devel-18.5.0.0.0-3.x86_64.rpm
ENV LD_LIBRARY_PATH="/usr/lib/oracle/18.5/client64/lib:${LD_LIBRARY_PATH}"


RUN mkdir /app
WORKDIR /app
# Копируем бинарник и папки со статикой из билдера
COPY --from=builder /app/app /app/
COPY ./templates/ /app/templates
COPY config.yml /app/config.yml
COPY ./migrations /app/migrations
COPY ./static /app/static

ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# Run the web service on container startup.
CMD ["/app/app"]