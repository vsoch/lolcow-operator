FROM ubuntu:latest

# build the lolcow container
# docker build -f docker/Dockerfile -t ghcr.io/vsoch/lolcow-operator .
# docker run -it -p 8080:8080 ghcr.io/vsoch/lolcow-operator

ENV LC_ALL=C
ENV PATH=/usr/games:$PATH

RUN apt-get -y update && \
    apt-get -y install fortune cowsay lolcat python3-pip && \
    pip install flask

WORKDIR /code
COPY ./docker/ /code
COPY ./docker/entrypoint.sh /entrypoint.sh
EXPOSE 8080
ENTRYPOINT ["/bin/bash", "/entrypoint.sh"]
