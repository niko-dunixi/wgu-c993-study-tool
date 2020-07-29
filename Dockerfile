FROM golang:1.14 AS HEALTHCHECK_BUILDER
WORKDIR /workdir
COPY /healthcheck/go.mod /healthcheck/go.sum ./
RUN go mod download
ADD ./healthcheck /workdir
RUN go build .

FROM golang:1.14 AS GENERATOR_BUILDER
RUN go get -u github.com/shurcooL/vfsgen/cmd/vfsgendev
WORKDIR /workdir
COPY /data-generator/go.mod /data-generator/go.sum ./
RUN go mod download
ADD ./data-generator /workdir
RUN go generate
RUN go build .

FROM oracle/database:12.2.0.1-ee AS MY_ORACLE
ENV ORACLE_SID=mysid
ENV ORACLE_PDB=mypdb
ENV ORACLE_PWD=My1Simple2Password
ENV ADMIN_PASSWORD=My1Admin2Password3
ENV USER_PASSWORD=N01InsecureAtAll
USER root
RUN yum install -y sudo
RUN echo "oracle ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers
USER oracle
COPY --from=HEALTHCHECK_BUILDER /workdir/healthcheck /bin/healthcheck
HEALTHCHECK --interval=30s --timeout=30s --start-period=10m --retries=3 CMD [ "/bin/healthcheck" ]
COPY --from=GENERATOR_BUILDER /workdir/data-generator /bin/data-generator
