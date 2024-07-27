# go binaries are taken from here
FROM golang:1.22 AS go-dependencies

# build image
FROM ubuntu:jammy AS build

# install dqlite
RUN apt-get update
RUN apt-get install -y software-properties-common python3-launchpadlib
RUN https_proxy="" http_proxy="" add-apt-repository ppa:dqlite/dev -y
RUN apt-get install --no-install-recommends -y libdqlite-dev build-essential make curl git

# install nodejs
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
RUN apt-get install -y nodejs
RUN npm install -g yarn

# import golang
COPY --from=go-dependencies /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# build binaries
WORKDIR /srv
COPY . .
RUN make

# demo base image
FROM ubuntu:jammy

# copy binaries and scripts
WORKDIR /srv
COPY --from=build /root/go/bin/lxd-cluster-mgr /srv/lxd-cluster-mgr
COPY --from=build /root/go/bin/lxd-cluster-mgrd /srv/lxd-cluster-mgrd
COPY --from=build /srv/entrypoint /srv/entrypoint
COPY --from=build /srv/scripts/generate_clusters.sh /srv/scripts/generate_clusters.sh
COPY --from=build /srv/ui/haproxy-demo.cfg /srv/ui/haproxy-demo.cfg
COPY --from=build /usr/lib/x86_64-linux-gnu/libdqlite.so.0 /usr/lib/x86_64-linux-gnu/libdqlite.so.0
COPY --from=build /usr/lib/x86_64-linux-gnu/libuv.so.1 /usr/lib/x86_64-linux-gnu/libuv.so.1
COPY --from=build /usr/lib/x86_64-linux-gnu/libraft.so.3 /usr/lib/x86_64-linux-gnu/libraft.so.3
COPY --from=build /usr/lib/x86_64-linux-gnu/libsqlite3.so.0 /usr/lib/x86_64-linux-gnu/libsqlite3.so.0

RUN apt-get update && \
    apt-get install haproxy -y && \
    apt-get install ca-certificates -y && update-ca-certificates

ENTRYPOINT ["./entrypoint"]
