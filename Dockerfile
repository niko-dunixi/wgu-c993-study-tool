FROM golang:1.14 AS TOOL_SUITE_BUILDER
WORKDIR /workdir
COPY /tool-suite/go.mod /tool-suite/go.sum ./
RUN go mod download
COPY /tool-suite /workdir
RUN go generate ./...
RUN go install ./...

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
COPY --from=TOOL_SUITE_BUILDER /go/bin/healthcheck /bin/healthcheck
COPY --from=TOOL_SUITE_BUILDER /go/bin/database-hydrator /bin/database-hydrator
HEALTHCHECK --interval=30s --timeout=30s --start-period=10m --retries=3 CMD [ "/bin/healthcheck" ]