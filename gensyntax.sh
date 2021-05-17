#!/usr/bin/env sh

# gensyntax.sh LANGUAGE
#
# Generate a tokenizer for a syntax language.
# This is useful for updating a language after editing the syntax rules.

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 LANGUAGE"
    exit 1
fi

(cd $(dirname $0)/syntax; \
 go run gen_tokenizers.go -language "$1" && \
 goimports -w -local "github.com/aretext" .)
