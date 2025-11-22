FROM alpine:latest as builder
ARG TARGETPLATFORM
RUN echo "I'm building for $TARGETPLATFORM"

RUN apk add --no-cache gzip && \
    mkdir /clash-config && \
    wget -O /clash-config/geoip.metadb https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.metadb && \
    wget -O /clash-config/geosite.dat https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat && \
    wget -O /clash-config/geoip.dat https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.dat

COPY docker/file-name.sh /clash/file-name.sh
WORKDIR /clash
COPY bin/ bin/
RUN FILE_NAME=`sh file-name.sh` && echo $FILE_NAME && \
    FILE_NAME=`ls bin/ | egrep "$FILE_NAME.gz"|awk NR==1` && echo $FILE_NAME && \
    mv bin/$FILE_NAME clash.gz && gzip -d clash.gz && chmod +x clash && echo "$FILE_NAME" > /clash-config/test
FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/MetaCubeX/clash"

RUN apk add --no-cache ca-certificates tzdata iptables

VOLUME ["/root/.config/clash/"]

COPY --from=builder /clash-config/ /root/.config/clash/
COPY --from=builder /clash/clash /clash
ENTRYPOINT [ "/clash" ]
