#!/bin/sh

OUTPUT="./bin"
mkdir -p $OUTPUT

build()
{
    echo ">>" $1
    go build -o $OUTPUT code.51shunyou.com/shunyou/bingo/src/$1
}

if [ "$1" != "" ]; then
    build $1
    exit $?
fi

for d in `find ./src -maxdepth 1 -type d`
do
MODULE=${d##*/}
if [ "$MODULE" != "src" -a "$MODULE" != "def" ]; then
    build $MODULE
fi
done
