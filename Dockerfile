FROM ubuntu:14.04
MAINTAINER Daisuke Fujita <dtanshi45@gmail.com> (@dtan4)

ENV DOCKER_COMPOSE_VERSION 1.6.0
ENV GITRECEIVE_COMMIT d152fd28e9dba9fcd0af5366cf188fc89ce8385f

RUN apt-get update && \
    apt-get install -y git openssh-server wget && \
    rm -rf /var/lib/apt/lists/*

RUN wget -O /usr/local/bin/docker-compose https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-Linux-x86_64 && \
    chmod +x /usr/local/bin/docker-compose

RUN wget -O /usr/local/bin/gitreceive https://raw.githubusercontent.com/progrium/gitreceive/$GITRECEIVE_COMMIT/gitreceive && \
    chmod +x /usr/local/bin/gitreceive

RUN mkdir /var/run/sshd
COPY files/sshd_config /etc/ssh/

RUN gitreceive init
RUN echo "git:passwd" | chpasswd
COPY files/receiver /home/git/

COPY files/upload-key /usr/local/bin/

COPY entrypoint.sh /

VOLUME /home/git
EXPOSE 22

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/sbin/sshd", "-D", "-e"]
