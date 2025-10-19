#!/bin/bash

# test script that uses the spacectl CLI to test the API

make build

./bin/spacectl auth login --email $TEST_EMAIL --password $TEST_PASSWORD

./bin/spacectl org list

# create an organization
./bin/spacectl org create test --description "Test organization"

# list projects
./bin/spacectl project list

# create a project
./bin/spacectl project create test --description "Test project" --org-name test

# delete the project
./bin/spacectl project delete test

# delete the organization
./bin/spacectl org delete test