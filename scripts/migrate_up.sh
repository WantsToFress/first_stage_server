#!/usr/bin/env bash
migrate -path assets/migrations -database "postgres://postgres@localhost:5432/hack?sslmode=disable" up