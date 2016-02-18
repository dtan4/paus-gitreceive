#!/bin/bash

if [ ! -d /home/git/.ssh ]; then
  mkdir /home/git/.ssh
  chown -R git /home/git/.ssh
fi

if [ -n "$PAUS_BASE_DOMAIN" ]; then
  echo $PAUS_BASE_DOMAIN > /base-domain
else
  echo "example.com" > /base-domain
fi

exec $@
