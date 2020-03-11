#!/usr/bin/env bash
# go get github.com/dizzyfool/genna
genna model-named -c "postgres://postgres@localhost:5432/hack?sslmode=disable" -o "internal/model/model.go" -t "public.*" -f -s deleted_at
#genna search -c postgres://postgres@localhost:5432/postgres?sslmode=disable -o internal/model/search.go -t public.* -f
# genna validation -c postgres://postgres@localhost:5432/postgres?sslmode=disable -o ../internal/model/validation.go -t public.* -f
# fix error with uuid - use IsValidUUID() method, set ErrWrongValue