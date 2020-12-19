#!/bin/sh

idl_path="idl"

for p in `ls $idl_path/*.proto`
do
f=${p##*/}
name=${f%.*}
out=src/def
mkdir -p $out
echo ">>" $f
protoc --proto_path=$idl_path --go_out=plugins=grpc:$out $p
protoc-go-inject-tag -input=$out/$name.pb.go
done

echo ">>" usdt.abi
abigen --abi ./utils/chain/eth/usdt.abi --pkg eth --type USDT --out ./utils/chain/eth/usdt.go

# 去掉所有的omitempty，这个命令和平台相关
if [ "$(uname)" = "Darwin" ]; then
sed -i "" "s/,omitempty//g" ./src/def/*
else
sed -i "s/,omitempty//g" ./src/def/*
fi