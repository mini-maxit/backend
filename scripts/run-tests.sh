#!/usr/bin/env bash

POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=test-maxit
POSTGRES_CONTAINER_NAME=maxit-testdb
BROKER_CONTAINER_NAME=maxit-testbroker
COVERAGE_FILE="coverage.out"
COVERAGE_HTML="coverage.html"

# Function to check if Docker is running
check_docker() {
  if ! docker info >/dev/null 2>&1; then
    echo -e "\033[31mError: Docker is not running. Please start Docker and try again.\033[0m"
    exit 1
  fi
}

# Function to setup the container
setup_container() {
  check_docker
  docker rm -f $POSTGRES_CONTAINER_NAME 2>/dev/null || true
  docker run -d --name $POSTGRES_CONTAINER_NAME \
    -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
    -e POSTGRES_USER=$POSTGRES_USER \
    -e POSTGRES_DB=$POSTGRES_DB \
    -p 5432:5432 \
    postgres:17
    docker rm -f $BROKER_CONTAINER_NAME 2>/dev/null || true
  docker run -d --name $BROKER_CONTAINER_NAME \
    -p 5672:5672 \
    -p 15672:15672 \
    rabbitmq:3.13-management
  echo -e "\033[32mContainer setup complete.\033[0m"
}

# Function to run tests with coverage
generate_coverage() {
  echo -e "\033[32mRunning tests with coverage...\033[0m"
  go test -v ./... -cover -coverprofile=$COVERAGE_FILE
  if [ $? -eq 0 ]; then
    echo -e "\033[32mTests completed successfully. Excluding mockgen.go files from coverage...\033[0m"
    grep -v "mockgen.go" $COVERAGE_FILE > ${COVERAGE_FILE}.tmp && mv ${COVERAGE_FILE}.tmp $COVERAGE_FILE
    echo -e "\033[32mGenerating HTML coverage report: $COVERAGE_HTML\033[0m"
    go tool cover -html=$COVERAGE_FILE -o $COVERAGE_HTML
    echo -e "\033[32mCoverage report generated: $COVERAGE_HTML\033[0m"
  else
    echo -e "\033[31mTests failed. See output above for details.\033[0m"
    exit 1
  fi
}

# Function to clean up the container
cleanup_container() {
  check_docker
  docker rm -f $POSTGRES_CONTAINER_NAME 2>/dev/null || true
  docker rm -f $BROKER_CONTAINER_NAME 2>/dev/null || true
  echo -e "\033[32mContainer cleanup complete.\033[0m"
}

# Parse the first argument
case "$1" in
  # setup)
  #   setup_container
  #   exit 0
  #   ;;
  run)
    # check_docker
    # if ! docker ps --format '{{.Names}}' | grep -q "^$POSTGRES_CONTAINER_NAME$"; then
    #   echo -e "\033[31mError: Container $POSTGRES_CONTAINER_NAME is not running. Please setup the container first.\033[0m"
    #   exit 1
    # fi
    echo -e "\033[32mRunning tests...\033[0m"
    go test -v ./...
    exit 0
    ;;
  cover)
    generate_coverage
    exit 0
    ;;
  # cleanup)
  #   cleanup_container
  #   exit 0
  #   ;;
  # all)
  #   setup_container
  #   sleep 7
  #   generate_coverage
  #   cleanup_container
  #   exit 0
  #   ;;
  *)
    echo -e "\033[31mUsage: $0 {run|cover}\033[0m"
    exit 1
    ;;
esac
