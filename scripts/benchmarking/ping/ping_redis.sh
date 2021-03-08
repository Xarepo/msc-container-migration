#!/bin/sh

redis-server --save '' --appendonly no &

sh /ping.sh
