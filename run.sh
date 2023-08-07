#!/bin/sh

mpirun -np 4 go run cmd/main.go -mpi --redisURL="redis://pchpc-redis-1:6379" -debug $1