#!/bin/bash

# for line in $(cat secrets.toml | sed 's/ //g' | sed "s/\"/'/g"); do
for line in $(cat secrets.toml | sed 's/[ "]//g'); do
    echo $line
    export $(echo $line)
done
go run .