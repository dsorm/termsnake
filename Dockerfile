# use ubuntu focal as base image
# builder stage
FROM ubuntu:focal AS builder

# make sure we're root
USER root

# get build dependencies
# get go toolchain
WORKDIR /tmp
RUN apt-get update && apt-get install wget unzip -y && \
wget https://dl.google.com/go/go1.17.3.linux-amd64.tar.gz -O /tmp/go.linux-amd64.tar.gz && \
tar -C /usr/local -xzf go.linux-amd64.tar.gz && \
rm /tmp/go.linux-amd64.tar.gz

WORKDIR /root/go/src/github.com/dsorm/termsnake/

# copy source files
COPY . .

# get dependencies and compile
RUN /usr/local/go/bin/go install github.com/dsorm/termsnake

# final image stage
FROM ubuntu:focal

# copy artefacts and needed files
RUN mkdir /app && mkdir /app/html
COPY --from=builder /root/go/bin/termsnake /app/termsnake

# run
WORKDIR /app
CMD ["./termsnake"]