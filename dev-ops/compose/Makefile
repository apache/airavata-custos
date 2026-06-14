SHELL := /bin/bash
.PHONY: base up down build cli test test-integration smoke lint keys logs

up: base
	docker compose up -d --build

down:
	docker compose down -v
