#!/bin/bash

if [ ! -d /home/git/.ssh ]; then
  mkdir /home/git/.ssh
  chown -R git /home/git/.ssh
fi

if [ -f /var/run/docker.sock ]; then
  chown root:docker /var/run/docker.sock
fi

exec $@
