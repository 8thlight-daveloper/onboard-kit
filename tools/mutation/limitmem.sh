#!/usr/bin/env bash
ulimit -v "${MUTATION_MEMORY_KB:-3145728}" 2> /dev/null || true
exec "$@"
