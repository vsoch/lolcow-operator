FROM ubuntu:latest

# build the lolcow container
# docker build -f docker/Dockerfile -t ghcr.io/vsoch/lolcow-operator .

ENV LC_ALL=C
ENV PATH=/usr/games:$PATH

RUN apt-get -y update && \
    apt-get -y install fortune cowsay lolcat

COPY ./docker/entrypoint.sh /entrypoint.sh
ENTRYPOINT ["/bin/bash", "/entrypoint.sh"]
