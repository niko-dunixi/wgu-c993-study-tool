FROM golang:1.14 AS ORACLE_DOWNLOADER_BUILDER
WORKDIR /workdir
COPY oracle-downloader /workdir
RUN go build .

FROM chromedp/headless-shell:latest AS ORACLE_DOWNLOADER
ARG ORACLE_USERNAME
ARG ORACLE_PASSWORD
ARG ORACLE_AGREE_TO_TERMS_OF_SERVICE
WORKDIR /workdir
COPY --from=ORACLE_DOWNLOADER_BUILDER /workdir/oracle-downloader /workdir/oracle-downloader
RUN ./oracle-downloader

FROM docker:stable-dind AS ORACLE_BUILDER
RUN apk add bash git
WORKDIR /workdir
RUN git clone https://github.com/oracle/docker-images.git --depth=1
COPY --from=ORACLE_DOWNLOADER /workdir/linuxx64_12201_database.zip /workdir/docker-images/OracleDatabase/SingleInstance/dockerfiles/12.2.0.1
WORKDIR /workdir/docker-images/OracleDatabase/SingleInstance/dockerfiles/
CMD ./buildDockerImage.sh -v 12.2.0.1 -e
