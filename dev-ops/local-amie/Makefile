SHELL := /bin/bash
.PHONY: up down build logs ps wait

up:
	docker compose up -d --build
	$(MAKE) wait

down:
	docker compose down -v --remove-orphans

build:
	docker compose build

logs:
	docker compose logs -f --tail=100

ps:
	docker compose ps

# Blocks until both services report healthy. Times out at 90s.
wait:
	@deadline=$$(( $$(date +%s) + 90 )); \
	while true; do \
	  db_ok=$$(docker inspect -f '{{.State.Health.Status}}' custos_amie_test_db 2>/dev/null || echo starting); \
	  mock_ok=$$(docker inspect -f '{{.State.Health.Status}}' custos_amie_test_mock 2>/dev/null || echo starting); \
	  if [ "$$db_ok" = "healthy" ] && [ "$$mock_ok" = "healthy" ]; then echo "ready"; exit 0; fi; \
	  if [ "$$(date +%s)" -ge "$$deadline" ]; then \
	    echo "TIMED OUT waiting for healthy: db=$$db_ok mock=$$mock_ok" >&2; \
	    docker compose ps; exit 1; \
	  fi; \
	  sleep 2; \
	done
