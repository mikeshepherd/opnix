default:
    @just --list

gomod2nix:
    gomod2nix generate

build:
    just gomod2nix
    nix build

run:
    just build
    ./result/bin/opnix
