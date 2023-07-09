#!/bin/bash

for ((i = 1; i <= 10; i++)); do
	go test -bench=. -benchtime=30s >~/benchmark_$(date +%s)_sequential.log
done
