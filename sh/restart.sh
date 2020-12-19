#!/bin/sh

tar -xzvf output.tar.gz
pkill chain
pkill entity
pkill gravity
pkill idasc
pkill idgen
pkill oauth
pkill pay
pkill randgen
pkill team
pkill wallet
#pkill gateway