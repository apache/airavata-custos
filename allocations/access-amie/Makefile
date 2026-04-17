# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

BINARY := access-amie
DB_DSN ?= mysql://admin:admin@tcp(localhost:3306)/access_ci

.PHONY: build test test-short lint proto migrate-up migrate-down clean fmt tidy

build:
	go build -o bin/$(BINARY) .

test:
	go test ./... -v -count=1

test-short:
	go test ./... -v -short

lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

proto:
	protoc --go_out=proto/gen --go_opt=paths=source_relative \
	       proto/amie_packets.proto proto/internal_events.proto

migrate-up:
	migrate -path db/migrations -database "$(DB_DSN)" up

migrate-down:
	migrate -path db/migrations -database "$(DB_DSN)" down 1

clean:
	rm -rf bin/
