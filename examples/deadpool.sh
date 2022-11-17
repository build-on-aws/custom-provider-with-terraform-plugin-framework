#!/bin/bash

curl -X POST http://localhost:9200/buildonaws/_doc \
    -H 'Content-Type: application/json' \
    -d @deadpool.json
