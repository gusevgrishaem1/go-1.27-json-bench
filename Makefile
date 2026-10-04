.DEFAULT_GOAL := help

POWERSHELL ?= pwsh
BENCHTIME ?= 2s
COUNT ?= 5
PAYLOADS ?= small,medium,large,unicode,nested
VUS ?= 5
DURATION ?= 10m
TEXT_SIZE ?= 64
SLEEP ?= 0
ITEMS ?=
TARGETS ?=

.PHONY: help bench bench-smoke build up down load

help:
	@echo make bench [BENCHTIME=2s COUNT=5]
	@echo make bench-smoke
	@echo make build / make up / make down
	@echo make load [PAYLOADS=small,large VUS=5 DURATION=2m ITEMS=500 TEXT_SIZE=256]

bench:
ifeq ($(OS),Windows_NT)
	$(POWERSHELL) -NoProfile -File ./run-benchmarks.ps1 -Benchtime "$(BENCHTIME)" -Count "$(COUNT)"
else
	sh ./run-benchmarks.sh "$(BENCHTIME)" "$(COUNT)"
endif

bench-smoke:
ifeq ($(OS),Windows_NT)
	$(POWERSHELL) -NoProfile -File ./run-benchmarks.ps1 -Benchtime 1x -Count 1
else
	sh ./run-benchmarks.sh 1x 1
endif

build:
	docker compose build

up:
	docker compose up -d --build

down:
	docker compose down

# Start the stack with make up before running load.
# Omit optional variables so the script keeps its profile sizes and default URLs.
load:
	docker compose --profile load run --rm -e "PAYLOADS=$(PAYLOADS)" -e "VUS=$(VUS)" -e "DURATION=$(DURATION)" -e "TEXT_SIZE=$(TEXT_SIZE)" -e "SLEEP=$(SLEEP)" $(if $(strip $(ITEMS)),-e "ITEMS=$(ITEMS)") $(if $(strip $(TARGETS)),-e "TARGETS=$(TARGETS)") loadgen
