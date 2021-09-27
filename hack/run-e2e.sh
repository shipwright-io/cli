#!/usr/bin/env bash

for i in test/e2e/*.bats; do
	./test/e2e/bats/core/bin/bats ${i}
done