# Makefile
SHELL := /bin/bash
.PHONY: base up down build cli test test-integration smoke lint keys logs

base:
	docker build -f docker/Dockerfile.base -t slurmrest/base:24.05 .

up: base
	docker compose up -d --build

down:
	docker compose down -v

build: base
	docker compose build

smoke:
	docker compose exec login sbatch --wrap 'hostname' -o /tmp/out.txt
	sleep 5
	docker compose exec login sacct -n -o JobID,State --starttime now-5minutes | tail -n 5
	docker compose exec login cat /tmp/out.txt

lint:
	cd cli && go vet ./...

keys:
	docker compose up init-keys

logs:
	docker compose logs -f --tail=100

token:
	docker compose exec login scontrol token