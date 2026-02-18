#!/bin/bash

set -e
set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source "$SCRIPT_DIR/setup-env.sh"

# build golang app
make build-coverage

# enable coverage
export GOCOVERAGE=true
export GOCOVERDIR="$SCRIPT_DIR/coverage"
rm -rf $GOCOVERDIR
mkdir -p $GOCOVERDIR

# run the cluster connector
export SERVICE=cluster-connector
export SERVER_PORT=9000
export STATUS_PORT=9009
nohup cmd/app-coverage > >(cat) 2> >(cat >&2) &
CLUSTER_CONNECTOR_PID=$!

# run the management api
export SERVICE=management-api
export SERVER_PORT=30000
export STATUS_PORT=30003
nohup cmd/app-coverage > >(cat) 2> >(cat >&2) &
MANAGEMENT_API_PID=$!

# wait for the apps to start
echo "waiting for services to start..."
sleep 10

# run golang e2e tests
echo "running golang e2e tests..."
go test -count=1 -v ./test/e2e

# run golang cli tests
echo "running golang cli tests..."
export SERVICE=cli
cmd/app-coverage enroll cluster-test-enroll
cmd/app-coverage enroll cluster-test-enroll-with-expire --expire 2042-05-23T17:00:00Z --description 'Here be dragons'

# run ui unit tests
echo "running ui unit tests..."
rm -rf ui/coverage
cd ui
yarn test-unit-coverage
cd ..

# run ui e2e tests
echo "running ui e2e tests..."
echo "OIDC_USER=cluster-manager-e2e-tests@example.org" >> ui/.env.local
echo "OIDC_PASSWORD=cluster-manager-e2e-password" >> ui/.env.local
cd ui
dotrun &
curl --head --fail --retry-delay 2 --retry 100 --retry-connrefused --insecure https://ma.lxd-cm.local:8414
npx playwright install --with-deps chromium
export COVERAGE=1 # ensure we run against dotrun ui, so the correct source maps and paths are used
yarn test-e2e-coverage
unset COVERAGE
cd ..
DOTRUN_CONTAINER_ID=$(docker ps | grep dotrun | awk '{print $1}' | tail -n1)
docker stop $DOTRUN_CONTAINER_ID

# combine ui coverage reports
echo "combining ui coverage reports..."
cd ui
yarn test-combine-coverage-reports
cd ..
cp ui/coverage/playwright-report/cobertura-coverage.xml test/coverage/coverage-ui.xml
cp -R ui/coverage/playwright-report test/coverage

# kill app processes
echo "stopping services..."
kill $CLUSTER_CONNECTOR_PID
kill $MANAGEMENT_API_PID

stop_postgres
stop_prometheus

# show coverage results
echo "coverage results:"
go tool covdata percent -i="${GOCOVERDIR}"

echo "convert coverage report to xml..."
go install github.com/boumenot/gocover-cobertura@latest
go tool covdata textfmt -i="${GOCOVERDIR}" -o "${GOCOVERDIR}"/coverage.out
gocover-cobertura < "${GOCOVERDIR}"/coverage.out > "${GOCOVERDIR}"/coverage-go.xml

# move coverage reports to .coverage folder for TICS
rm -rf .cover
mkdir -p .cover
cp test/coverage/coverage-go.xml .cover/coverage-go.xml
cp test/coverage/coverage-ui.xml .cover/coverage-ui.xml

echo "done."
