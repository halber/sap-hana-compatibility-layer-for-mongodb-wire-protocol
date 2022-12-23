FROM golang:1.20.0-bullseye as builder
WORKDIR /src

ENV CGO_ENABLED=1
ENV GORACE=halt_on_error=1,history_size=2

# download hanaclient https://tools.hana.ondemand.com/#hanatools
# and install SAP HANA Client 2.0
ENV PATH=$PATH:/usr/sap/hdbclient
ENV CGO_LDFLAGS=/usr/sap/hdbclient/libdbcapiHDB.so
ENV GO111MODULE=auto
ENV LD_LIBRARY_PATH=/usr/sap/hdbclient/
RUN mkdir driver \
    && curl 'https://tools.hana.ondemand.com/additional/hanaclient-2.15.19-linux-x64.tar.gz' \
       -H 'Cookie: eula_3_1_agreed=tools.hana.ondemand.com/developer-license-3_1.txt' \
       --output driver/hanaclient.tar.gz \
    && tar -xzvf driver/hanaclient.tar.gz -C driver \
    && driver/client/./hdbinst --batch --ignore=check_diskspace \
    && mv /usr/sap/hdbclient/golang/src/SAP /usr/local/go/src/ \
    && ( \
        cd /usr/sap/hdbclient/golang/src/ \
        && go install SAP/go-hdb/driver ) \
    && rm -rf driver


# build the Hana compatibility layer for mongoDB
ADD . .
RUN make init build-testcover

# cleanup driver directory for copying
WORKDIR /usr/sap/hdbclient
RUN rm -rf *.py *.tar.gz *.jar calcviewapi dotnetcore examples golang node ruby rtt rtt.sh sdk


###################

FROM buildpack-deps:bullseye-scm
ARG VERSION
ARG COMMIT

ENV PATH=$PATH:/usr/sap/hdbclient
ENV CGO_LDFLAGS=/usr/sap/hdbclient/libdbcapiHDB.so
ENV GO111MODULE=auto
ENV LD_LIBRARY_PATH=/usr/sap/hdbclient/

WORKDIR /usr/sap/hdbclient
COPY --from=builder /usr/sap/hdbclient ./

RUN groupadd -r appuser && useradd -r -g appuser appuser
USER appuser
WORKDIR /home/appuser
COPY --from=builder /src/bin/SAPHANAcompatibilitylayer-testcover ./

ENTRYPOINT ["./SAPHANAcompatibilitylayer-testcover", "-test.coverprofile=cover.txt", "-mode=normal", "-listen-addr=:27017"]
EXPOSE 27017

LABEL org.opencontainers.image.description="SAP HANA compatibility layer for MongoDB Wire Protocol"
LABEL org.opencontainers.image.revision="${COMMIT}"
LABEL org.opencontainers.image.source="https://github.com/SAP/sap-hana-compatibility-layer-for-mongodb-wire-protocol"
LABEL org.opencontainers.image.title="SAP HANA compatibility layer for MongoDB Wire Protocol"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL hanaclient-version="2.15.19-linux-x64"