FROM golang:1.14 AS HEALTHCHECK_BUILDER
WORKDIR /workdir
COPY go.mod go.sum ./
RUN go mod download
ADD ./healthcheck /workdir
RUN go build .

FROM oracle/database:12.2.0.1-ee AS MY_ORACLE
ENV ORACLE_SID=mysid
ENV ORACLE_PDB=mypdb
ENV ORACLE_PWD=My1Simple2Password
USER root
RUN yum install -y sudo
RUN echo "oracle ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
USER oracle
COPY --from=HEALTHCHECK_BUILDER /workdir/healthcheck /bin/healthcheck
HEALTHCHECK --interval=30s --timeout=30s --start-period=5m --retries=3 CMD [ "/bin/healthcheck" ]
