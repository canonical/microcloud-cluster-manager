#!/bin/bash
#
# Spin up single-node MicroCloud clusters in LXD containers and enroll each of them with
# the MicroCloud Cluster Manager that runs on this machine.
#
# For every container the script:
#   1. installs the LXD and MicroCloud snaps,
#   2. initializes a one node MicroCloud cluster with `microcloud preseed`,
#   3. configures its LXD with the OIDC settings of the Cluster Manager,
#   4. creates a remote cluster join token with the Cluster Manager CLI,
#   5. joins the cluster to the Cluster Manager,
#   6. enables the cluster manager reverse tunnel.
#
# The Cluster Manager development environment must be running (`make dev`) and the
# Cluster Manager binary must be built (`make build`). The environment that `make dev`
# sets up is loaded from `scripts/setup-env.sh`, because the Cluster Manager CLI reads
# the same configuration as the services.

set -uo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
CLI_BIN="$(cd -- "${SCRIPT_DIR}/.." && pwd)/cmd/app"

# The join token advertises the Cluster Connector address that the Cluster Manager runs
# behind. The Cluster Manager CLI reads it from the environment, so the script uses the
# same values and falls back to the development environment defaults. They are set in
# main(), after the environment is loaded.
CLUSTER_CONNECTOR_DOMAIN=""
CLUSTER_CONNECTOR_PORT=""

readonly PROJECT="microcloud-mcm"

# Containers are named "${INSTANCE_PREFIX}-01", "${INSTANCE_PREFIX}-02", ...
# LXD also uses the container name as its hostname and MicroCloud cluster name.
INSTANCE_PREFIX="microcloud"

# Name of the identity provider group that is mapped to the LXD group with the server
# admin entitlement. The Cluster Manager only accepts users in the OIDC group "admins",
# and an admin on the cluster has to be able to use it, so the same group name is used on
# both sides.
readonly IDP_ADMIN_GROUP="admins"

COUNT=2
LXD_CHANNEL="6/stable"
LXD_CHANNEL_EXPLICIT=0
MICROCLOUD_CHANNEL="3/stable"
TIMEOUT=600
VERBOSE=0
NUKE=0

WORK_DIR=""

# ==============================================================================
# Output helpers

log() {
  echo "==> $*"
}

log_error() {
  echo "ERROR: $*" >&2
}

print_usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [options]

Creates single-node MicroCloud clusters in LXD containers and enrolls each of them with
the Cluster Manager running on this machine.

Options:
  -n, --count NUM              Number of containers to create or delete.
                               Each container is a single-node MicroCloud cluster.
                               (default: ${COUNT})
      --instance-prefix PREFIX
               Prefix for container and cluster names. (default: ${INSTANCE_PREFIX})
      --nuke                   Force-delete the selected containers instead of setting them up.
      --microcloud-channel CHANNEL
                               Channel of the MicroCloud snap. (default: ${MICROCLOUD_CHANNEL})
      --lxd-channel CHANNEL    Channel of the LXD snap. (default: ${LXD_CHANNEL})
  -v, --verbose                Print the commands run in the containers.
  -h, --help                   Show this help message.

The containers are named ${INSTANCE_PREFIX}-01, ${INSTANCE_PREFIX}-02, ... in the ${PROJECT} project.
An existing container is reused, and clusters that are already enrolled are left as is.
Use --nuke with the same --instance-prefix and --count to remove the containers, including their data.
Missing containers are skipped.
The project and the Cluster Manager records are kept.
EOF
}

# ==============================================================================
# Argument parsing

parse_args() {
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -n|--count)
        if [ "$#" -lt 2 ]; then
          log_error "$1 requires a value"
          exit 1
        fi
        COUNT="${2:-}"
        shift 2
        ;;
      --instance-prefix)
        if [ "$#" -lt 2 ]; then
          log_error "$1 requires a value"
          exit 1
        fi
        INSTANCE_PREFIX="${2:-}"
        shift 2
        ;;
      --microcloud-channel)
        if [ "$#" -lt 2 ]; then
          log_error "$1 requires a value"
          exit 1
        fi
        MICROCLOUD_CHANNEL="${2:-}"
        shift 2
        ;;
      --lxd-channel)
        if [ "$#" -lt 2 ]; then
          log_error "$1 requires a value"
          exit 1
        fi
        LXD_CHANNEL="${2:-}"
        LXD_CHANNEL_EXPLICIT=1
        shift 2
        ;;
      --nuke)
        NUKE=1
        shift
        ;;
      -v|--verbose)
        VERBOSE=1
        shift
        ;;
      -h|--help)
        print_usage
        exit 0
        ;;
      *)
        log_error "Unknown argument: $1"
        print_usage
        exit 1
        ;;
    esac
  done

  if ! [[ "${COUNT}" =~ ^[0-9]+$ ]] || [ "${COUNT}" -lt 1 ]; then
    log_error "--count must be a positive integer, got '${COUNT}'"
    exit 1
  fi

  COUNT=$((10#${COUNT}))

  if ! [[ "${INSTANCE_PREFIX}" =~ ^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?$ ]]; then
    log_error "--instance-prefix must start with a letter, end with a letter or digit, and contain only letters, digits, and hyphens"
    exit 1
  fi

  local longest_name
  longest_name="$(printf '%s-%02d' "${INSTANCE_PREFIX}" "${COUNT}")"
  if [[ "${#longest_name}" -gt 63 ]]; then
    log_error "--instance-prefix and --count must produce container names of at most 63 characters"
    exit 1
  fi
}

# ==============================================================================
# Development environment

# load_dev_env loads the environment that the Cluster Manager CLI needs to create a join
# token. The CLI reads the same configuration as the services, including the Cluster
# Connector certificate that the token is signed with, so it has to run with the
# environment that `make dev` sets up.
#
# `scripts/setup-env.sh` is the source of that environment. It is sourced in a subshell
# with its side effects stubbed out, so that it only exports the variables and does not
# start the Postgres and Prometheus containers again.
load_dev_env() {
  local setup_env="${SCRIPT_DIR}/setup-env.sh"
  if [ ! -f "${setup_env}" ]; then
    log_error "Cannot find ${setup_env}"
    exit 1
  fi

  local exports
  exports="$(
    (
      docker() { :; }
      make() { :; }
      # shellcheck source=setup-env.sh
      source "${setup_env}" >/dev/null 2>&1
      # setup-env.sh enables tracing, which would otherwise end up in the output.
      set +x
      export -p
    ) 2>/dev/null
  )"

  # Keep only the variable declarations, so that nothing else can end up in the eval.
  # `declare -x` is turned into `export`, because `declare` inside a function would
  # create variables that are local to this function.
  exports="$(echo "${exports}" | grep '^declare -x ' | sed 's/^declare -x /export /')"
  if ! echo "${exports}" | grep -q "OIDC_CLIENT_ID"; then
    log_error "Failed to load the development environment from ${setup_env}"
    exit 1
  fi

  eval "${exports}"
}

# ==============================================================================
# Container helpers

# run_in_instance runs a command inside the container. This is the only place that calls
# `lxc exec`, so that every command gets the same treatment.
#
# Connectors are set up so that the caller can pipe data into the command, as when the
# preseed is piped to `microcloud preseed`. When the caller passes something on stdin,
# it is forwarded. Otherwise stdin is the terminal the script was started from, and it
# is disabled, because `lxc exec` forwards stdin to the command. Reading the terminal
# from a process in a background process group stops that process with SIGTTIN and
# hangs the script, and `timeout` runs the command in its own process group. The
# pseudo-terminal is disabled as well, so that a terminal on stdin cannot make `lxc exec`
# switch to interactive mode.
run_in_instance() {
  local name="$1"
  shift

  if [ "${VERBOSE}" -eq 1 ]; then
    # Values of configuration keys can hold a secret, such as oidc.client.secret, and the
    # output is also written to a log file. Only the keys are shown.
    local shown=()
    local arg
    for arg in "$@"; do
      shown+=("${arg%%=*}=<redacted>")
      if [[ "${arg}" != *=* ]]; then
        shown[-1]="${arg}"
      fi
    done
    # Written to stderr, so that the trace does not end up in output that the caller
    # captures, as when the subnet of the container is read.
    echo "+ ${name}: ${shown[*]}" >&2
  fi

  local stdin_flag=()
  if [ -t 0 ]; then
    stdin_flag=(--disable-stdin)
  fi

  timeout "${TIMEOUT}" lxc --project "${PROJECT}" exec --force-noninteractive "${stdin_flag[@]}" "${name}" "$@"
}

# instance_gateway returns the address of the host as seen from inside the container. The
# network is not configured the moment `lxc launch` returns, so wait for the default route.
instance_gateway() {
  run_in_instance "$1" -- sh -c '
    while ! ip -4 route show default | grep -q .; do sleep 1; done
    ip -4 route show default | awk "{print \$3; exit}"
  ' 2>/dev/null
}

# ==============================================================================
# Setup steps

ensure_project() {
  if lxc project show "${PROJECT}" &>/dev/null; then
    return 0
  fi

  log "Creating LXD project ${PROJECT}"
  # Share the host's images, networks, and configured default profile.
  if ! lxc project create "${PROJECT}" \
    -c features.images=false -c features.profiles=false -c features.networks=false; then
    log_error "Failed to create LXD project ${PROJECT}"
    return 1
  fi
}

nuke_instances() {
  local existing_instances
  if ! existing_instances="$(lxc --project "${PROJECT}" list --format csv --columns n)"; then
    log_error "Failed to list containers in LXD project ${PROJECT}"
    return 1
  fi

  local selected=()
  local name
  for name in "$@"; do
    if grep -Fxq -- "${name}" <<<"${existing_instances}"; then
      selected+=("${name}")
    else
      log "Container ${name} does not exist in LXD project ${PROJECT}, skipping"
    fi
  done

  if [[ "${#selected[@]}" -eq 0 ]]; then
    return 0
  fi

  log "Deleting containers in LXD project ${PROJECT}: ${selected[*]}"
  if ! lxc --project "${PROJECT}" delete --force "${selected[@]}"; then
    log_error "Failed to delete containers in LXD project ${PROJECT}"
    return 1
  fi
}

# create_instance creates the container, or starts it again when it exists. An existing
# container is kept so that running the script again does not discard a cluster that is
# already enrolled.
create_instance() {
  local name="$1"

  if lxc --project "${PROJECT}" info "${name}" &>/dev/null; then
    log "Container ${name} already exists, reusing it"
    # Fails when the container is already running, which is fine.
    lxc --project "${PROJECT}" start "${name}" 2>/dev/null || true

    return 0
  fi

  log "Creating container ${name}"
  # MicroCloud runs a nested LXD, which needs security.nesting.
  if ! lxc --project "${PROJECT}" launch ubuntu:24.04 "${name}" -c security.nesting=true; then
    log_error "Failed to create container ${name}"
    return 1
  fi
}

# setup_instance_hosts points the Cluster Connector domain at the host address as seen
# from inside the container. The join token advertises this domain, so it has to resolve
# to the Cluster Manager on every container.
setup_instance_hosts() {
  local name="$1"
  local host_address
  host_address="$(instance_gateway "${name}")"
  if [ -z "${host_address}" ]; then
    log_error "Failed to determine the host address as seen from container ${name}"
    return 1
  fi

  # The entry is rewritten on every run, because the host address can change when the
  # container is recreated or the host joins another network.
  local entry="${host_address} ${CLUSTER_CONNECTOR_DOMAIN} # microcloud-cluster-manager"
  if run_in_instance "${name}" -- grep -qxF "${entry}" /etc/hosts; then
    log "Container ${name} already points ${CLUSTER_CONNECTOR_DOMAIN} to ${host_address}"
    return 0
  fi

  log "Pointing ${CLUSTER_CONNECTOR_DOMAIN} to ${host_address} inside ${name}"
  if ! run_in_instance "${name}" -- sh -c "
    sed -i '/# microcloud-cluster-manager$/d' /etc/hosts
    echo '${entry}' >> /etc/hosts
  "; then
    log_error "Failed to configure the Cluster Manager host entry in ${name}"
    return 1
  fi
}

install_snaps() {
  local name="$1"

  log "Waiting for snapd to finish seeding in ${name}"
  run_in_instance "${name}" -- snap wait system seed.loaded || true

  # The snaps are checked first, so that a run that reuses a container reports what it
  # skips instead of reinstalling the snaps.
  local snap
  for snap in lxd microcloud; do
    if run_in_instance "${name}" -- snap list "${snap}" &>/dev/null; then
      log "The ${snap} snap is already installed in ${name}"
      # An explicit channel also applies to a container that was created before, so that
      # a container with an LXD that MicroCloud rejects can be fixed by running the script
      # again with the channel that provides a supported LXD.
      if [ "${snap}" = "lxd" ] && [ "${LXD_CHANNEL_EXPLICIT}" -eq 1 ]; then
        log "Refreshing the lxd snap in ${name} to the ${LXD_CHANNEL} channel"
        if ! run_in_instance "${name}" -- snap refresh lxd --channel="${LXD_CHANNEL}"; then
          log_error "Failed to refresh the lxd snap in ${name}"
          return 1
        fi
      fi

      continue
    fi

    log "Installing the ${snap} snap in ${name}"
    local channel="${LXD_CHANNEL}"
    if [ "${snap}" = "microcloud" ]; then
      channel="${MICROCLOUD_CHANNEL}"
    fi

    if ! run_in_instance "${name}" -- snap install "${snap}" --channel="${channel}"; then
      log_error "Failed to install the ${snap} snap in ${name}"
      return 1
    fi
  done

  log "Waiting for LXD to be ready in ${name}"
  if ! run_in_instance "${name}" -- lxd waitready --timeout "${TIMEOUT}"; then
    log_error "LXD did not become ready in ${name}"
    return 1
  fi
}

# initialize_cluster runs the MicroCloud preseed for a single node cluster. The cluster
# member is named after the container, because LXD sets the container hostname to the
# container name.
initialize_cluster() {
  local name="$1"
  local log_file="${WORK_DIR}/${name}.preseed.log"

  log "Waiting for MicroCloud to be ready in ${name}"
  if ! run_in_instance "${name}" -- microcloud waitready --timeout "${TIMEOUT}"; then
    log_error "MicroCloud did not become ready in ${name}"
    return 1
  fi

  # MicroCloud needs the subnet of the container's own network to find the addresses of
  # the cluster members.
  local subnet
  subnet="$(run_in_instance "${name}" -- sh -c "
    iface=\$(ip -4 route show default | awk '{print \$5; exit}')
    ip -4 route show dev \"\${iface}\" proto kernel scope link | awk '{print \$1; exit}'
  ")"
  if [ -z "${subnet}" ]; then
    log_error "Failed to determine the subnet for container ${name}"
    return 1
  fi

  log "Initializing the MicroCloud cluster in ${name}"
  local status=0
  if run_in_instance "${name}" --env TEST_CONSOLE=0 -- microcloud preseed >"${log_file}" 2>&1 <<EOF
lookup_subnet: ${subnet}
initiator: ${name}
systems:
- name: ${name}
EOF
  then
    status=0
  else
    status=$?
  fi
  if [ "${status}" -ne 0 ]; then
    # A cluster that was created before is left untouched, so that running the script
    # again only enrolls the clusters that the Cluster Manager does not know yet.
    if grep -q "MicroCloud is already initialized" "${log_file}"; then
      log "The MicroCloud cluster in ${name} is already initialized, reusing it"
      return 0
    fi

    if [ "${status}" -eq 124 ]; then
      log_error "MicroCloud preseed timed out after ${TIMEOUT} seconds in ${name}" 2>>"${log_file}"
    else
      log_error "MicroCloud preseed exited with status ${status} in ${name}" 2>>"${log_file}"
    fi
    if ! run_in_instance "${name}" -- journalctl -u snap.microcloud.daemon.service --no-pager -n 40 >>"${log_file}" 2>&1; then
      log_error "Failed to collect MicroCloud daemon logs in ${name}" 2>>"${log_file}"
    fi
    log_error "MicroCloud initialization failed in ${name} (exit status ${status}), see ${log_file}"
    tail -n 20 "${log_file}" >&2
    return 1
  fi

  if ! grep -q "MicroCloud is ready" "${log_file}"; then
    log_error "MicroCloud did not report the cluster as ready in ${name}, see ${log_file}"
    tail -n 20 "${log_file}" >&2
    return 1
  fi
}

# configure_instance_cluster_certificate binds the LXD API to the cluster address and
# adds that address to the SANs of the LXD cluster certificate.
#
# The requests that arrive over the reverse tunnel are proxied by microcloudd to the LXD of
# the cluster, and microcloudd verifies that connection against the cluster certificate.
# MicroCloud dials an address advertised by LXD, so the API must not listen on all
# interfaces and advertise addresses of nested bridges such as lxdfan0. LXD only puts
# the hostname and the loopback addresses in the SANs of its certificate, so verification
# fails with "certificate is valid for 127.0.0.1, ::1, not <address>". The certificate is
# regenerated with the address as an extra SAN to fix that.
configure_instance_cluster_certificate() {
  local name="$1"

  local endpoint
  if ! endpoint="$(run_in_instance "${name}" -- lxc --force-local config get cluster.https_address)"; then
    log_error "Failed to read the cluster address of ${name}"
    return 1
  fi

  local address
  address="${endpoint%%:*}"
  if [ -z "${address}" ]; then
    log_error "Failed to determine the cluster address of ${name}"
    return 1
  fi

  local api_endpoint
  if ! api_endpoint="$(run_in_instance "${name}" -- lxc --force-local config get core.https_address)"; then
    log_error "Failed to read the LXD API address of ${name}"
    return 1
  fi

  if [ "${api_endpoint}" != "${endpoint}" ]; then
    log "Binding the LXD API of ${name} to ${endpoint}"
    if ! run_in_instance "${name}" -- lxc --force-local config set "core.https_address=${endpoint}"; then
      log_error "Failed to bind the LXD API of ${name} to ${endpoint}"
      return 1
    fi
  fi

  if run_in_instance "${name}" -- sh -c "
    openssl x509 -in /var/snap/lxd/common/lxd/cluster.crt -noout -text |
      grep -A1 'Subject Alternative Name' |
      grep -qE 'IP Address:${address}([,[:space:]]|\$)'
  "; then
    log "The cluster certificate of ${name} already covers ${address}"
    return 0
  fi

  log "Adding ${address} to the SANs of the cluster certificate of ${name}"
  # The certificate is written to /root, because the lxc snap has a private /tmp and
  # `lxc cluster update-certificate` reads the files itself.
  local output
  if ! output="$(run_in_instance "${name}" -- sh -c "
    set -e
    openssl req -nodes -x509 -newkey ec -pkeyopt ec_paramgen_curve:secp384r1 -sha384 \
      -keyout /root/cluster.key -out /root/cluster.crt -days 3650 \
      -subj '/O=LXD/CN=root@${name}' \
      -addext 'subjectAltName=DNS:${name},IP:127.0.0.1,IP:::1,IP:${address}'
    lxc --force-local cluster update-certificate /root/cluster.crt /root/cluster.key
    rm -f /root/cluster.crt /root/cluster.key
  " 2>&1)"; then
    log_error "Failed to update the cluster certificate of ${name}: ${output}"
    return 1
  fi
}

# configure_instance_oidc points the LXD of the cluster at the same identity provider as
# the Cluster Manager, so that the tokens the Cluster Manager verifies are issued by a
# provider that LXD also trusts.
configure_instance_oidc() {
  local name="$1"

  # The values come from the environment that `make dev` sets up, which is what the
  # Cluster Manager itself runs with. The issuer and the client ID are required; the
  # audience is only used by providers that send one.
  local issuer="${OIDC_ISSUER}"
  local client_id="${OIDC_CLIENT_ID}"
  local client_secret="${OIDC_CLIENT_SECRET:-}"
  local audience="${OIDC_AUDIENCE:-}"

  if [ -z "${issuer}" ] || [ -z "${client_id}" ]; then
    log_error "OIDC_ISSUER and OIDC_CLIENT_ID are needed to configure the OIDC settings in ${name}"
    return 1
  fi

  # The Cluster Manager reads the groups from this claim, so LXD is told to read the same
  # one and the group names line up on both sides.
  local groups_claim="mcm-idp-groups"

  log "Configuring the OIDC settings in the LXD of ${name}"
  local settings=(
    "oidc.issuer=${issuer}"
    "oidc.client.id=${client_id}"
    "oidc.groups.claim=${groups_claim}"
  )

  # The identity provider is registered as a confidential client, which the Cluster
  # Manager authenticates as when it logs a user in. LXD talks to the same client when it
  # handles its own login callback, so it needs the secret to exchange the authorization
  # code for a token.
  if [ -n "${client_secret}" ]; then
    settings+=("oidc.client.secret=${client_secret}")
  fi

  if [ -n "${audience}" ]; then
    settings+=("oidc.audience=${audience}")
  fi

  local output
  if ! output="$(run_in_instance "${name}" -- lxc --force-local config set "${settings[@]}" 2>&1)"; then
    log_error "Failed to configure the OIDC settings in ${name}: ${output}"
    return 1
  fi
}

# configure_instance_admin_group maps the identity provider group that the Cluster Manager
# requires to the LXD group that has the server admin entitlement. Access is granted by
# the claim that the token carries, so without this an admin of the Cluster Manager could
# authenticate against the cluster's LXD but would not be authorized to do anything.
#
# LXD creates the "admins" group with the server admin entitlement itself, so only the
# mapping has to be added.
configure_instance_admin_group() {
  local name="$1"

  local existing
  existing="$(run_in_instance "${name}" -- lxc --force-local auth identity-provider-group list --format csv 2>/dev/null || true)"
  if echo "${existing}" | cut -d, -f1 | grep -qx "${IDP_ADMIN_GROUP}"; then
    log "The identity provider group ${IDP_ADMIN_GROUP} is already mapped to the LXD group ${IDP_ADMIN_GROUP} in ${name}"
    return 0
  fi

  log "Mapping the identity provider group ${IDP_ADMIN_GROUP} to the LXD group ${IDP_ADMIN_GROUP} in ${name}"
  local output
  if ! output="$(run_in_instance "${name}" -- lxc --force-local auth identity-provider-group create "${IDP_ADMIN_GROUP}" 2>&1)"; then
    log_error "Failed to create the identity provider group ${IDP_ADMIN_GROUP} in ${name}: ${output}"
    return 1
  fi

  if ! output="$(run_in_instance "${name}" -- lxc --force-local auth identity-provider-group group add "${IDP_ADMIN_GROUP}" "${IDP_ADMIN_GROUP}" 2>&1)"; then
    log_error "Failed to map the identity provider group ${IDP_ADMIN_GROUP} in ${name}: ${output}"
    return 1
  fi
}

# enroll_cluster joins the MicroCloud running in the container to the Cluster Manager and
# enables the reverse tunnel.
enroll_cluster() {
  local name="$1"
  local token
  local output

  if run_in_instance "${name}" --env TEST_CONSOLE=0 -- microcloud cluster-manager show &>/dev/null; then
    log "Cluster ${name} is already enrolled with the Cluster Manager"
    # An open tunnel retains the LXD address and certificate read when it connected.
    log "Restarting MicroCloud in ${name} to refresh the tunnel's LXD connection"
    if ! output="$(run_in_instance "${name}" -- snap restart microcloud.daemon 2>&1)"; then
      log_error "Failed to restart MicroCloud in ${name}: ${output}"
      return 1
    fi

    if ! output="$(run_in_instance "${name}" -- microcloud waitready --timeout "${TIMEOUT}" 2>&1)"; then
      log_error "MicroCloud did not become ready in ${name}: ${output}"
      return 1
    fi
  else
    log "Creating a join token for cluster ${name}"
    if ! token="$(create_join_token "${name}")"; then
      return 1
    fi

    log "Joining cluster ${name} to the Cluster Manager"
    if ! output="$(run_in_instance "${name}" --env TEST_CONSOLE=0 -- microcloud cluster-manager join "${token}" 2>&1)"; then
      log_error "Failed to join ${name} to the Cluster Manager: ${output}"
      return 1
    fi

    if ! echo "${output}" | grep -q "Successfully joined cluster manager"; then
      log_error "Unexpected output while joining ${name}: ${output}"
      return 1
    fi
  fi

  log "Enabling the reverse tunnel on cluster ${name}"
  if ! output="$(run_in_instance "${name}" --env TEST_CONSOLE=0 -- microcloud cluster-manager set reverse_tunnel true 2>&1)"; then
    log_error "Failed to enable the reverse tunnel on ${name}: ${output}"
    return 1
  fi

  local reverse_tunnel
  if ! reverse_tunnel="$(run_in_instance "${name}" --env TEST_CONSOLE=0 -- microcloud cluster-manager get reverse_tunnel)"; then
    log_error "Failed to read the reverse tunnel setting on ${name}"
    return 1
  fi

  if [ "${reverse_tunnel}" != "true" ]; then
    log_error "The reverse tunnel is not enabled on ${name}, got '${reverse_tunnel}'"
    return 1
  fi

  log "Cluster ${name} is enrolled with the reverse tunnel enabled"
}

# create_join_token creates a remote cluster join token for the given cluster name with
# the Cluster Manager CLI and returns it on stdout.
create_join_token() {
  local cluster_name="$1"
  local output

  if ! output="$(SERVICE=cli "${CLI_BIN}" enroll "${cluster_name}" 2>&1)"; then
    # A join token created by a previous run but never used is still stored, and the
    # Cluster Manager only keeps one token per cluster name.
    if echo "${output}" | grep -q "already exists"; then
      log_error "Cluster ${cluster_name} already has a join token. Remove it in the Cluster Manager UI and run the script again."
      return 1
    fi

    if echo "${output}" | grep -q "Failed to load configuration"; then
      log_error "The Cluster Manager CLI could not load its configuration. Start the development environment with 'make dev' and run the script again."
      return 1
    fi

    log_error "Failed to create a join token for cluster ${cluster_name}: ${output}"
    return 1
  fi

  # The CLI prints the join token on the last line of its output.
  echo "${output}" | tail -n 1
}

# ==============================================================================
# Cluster Manager certificate

# tls_cert_fingerprint prints the SHA-256 fingerprint of a PEM certificate in the same
# format that MicroCloud reports, which is what the join token carries.
tls_cert_fingerprint() {
  openssl x509 -in "$1" -noout -fingerprint -sha256 2>/dev/null |
    sed 's/.*=//; s/://g' |
    tr 'A-F' 'a-f'
}

# check_connector_certificate compares the certificate the Cluster Connector serves with
# the one the join token is signed with. The token advertises the certificate fingerprint
# of the Cluster Connector certificate, and MicroCloud refuses to join when the connector
# answers with a different one. The two only differ when the connector runs from a
# different checkout than the CLI, which is easy to do with git worktrees, so this is
# checked before the containers are created instead of at the end of the run.
check_connector_certificate() {
  local cert_path="${CLUSTER_CONNECTOR_TLS_PATH}/tls.crt"
  if [ ! -f "${cert_path}" ]; then
    log_error "Cannot find the Cluster Connector certificate at ${cert_path}"
    exit 1
  fi

  local expected
  expected="$(tls_cert_fingerprint "${cert_path}")"

  local served
  served="$(openssl s_client -connect "${CLUSTER_CONNECTOR_DOMAIN}:${CLUSTER_CONNECTOR_PORT}" \
    -servername "${CLUSTER_CONNECTOR_DOMAIN}" </dev/null 2>/dev/null |
    openssl x509 -noout -fingerprint -sha256 2>/dev/null |
    sed 's/.*=//; s/://g' |
    tr 'A-F' 'a-f')"

  if [ "${expected}" = "${served}" ]; then
    return 0
  fi

  log_error "The Cluster Connector at ${CLUSTER_CONNECTOR_DOMAIN}:${CLUSTER_CONNECTOR_PORT} serves a different certificate than ${cert_path}."
  log_error "Expected fingerprint ${expected}, got ${served}."
  log_error "The running connector was probably started from another checkout. Restart it from this checkout with 'make dev'."
  exit 1
}

# ==============================================================================
# Work directory

# cleanup_work_dir removes the work directory of a successful run. The logs are kept when
# the run fails, because they are what the failure has to be debugged with.
cleanup_work_dir() {
  local status=$?

  if [ -z "${WORK_DIR}" ] || [ ! -d "${WORK_DIR}" ]; then
    return 0
  fi

  if [ "${status}" -eq 0 ]; then
    rm -rf "${WORK_DIR}"

    return 0
  fi

  echo "Setup logs are available in ${WORK_DIR}" >&2
}

# ==============================================================================
# Parallel execution helper

# run_parallel runs the given function for every given instance in the background. The
# output of every run is prefixed with the instance name, printed, and kept in the work
# directory.
run_parallel() {
  local what="$1"
  local func="$2"
  shift 2

  local pids=()
  local name
  for name in "$@"; do
    (
      set -eo pipefail
      "${func}" "${name}" 2>&1 | sed -u "s/^/[${name}] /" | tee "${WORK_DIR}/${name}.${func}.log"
    ) &
    pids+=("$!")
  done

  local pid
  local failed=0
  for pid in "${pids[@]}"; do
    if ! wait "${pid}"; then
      failed=1
    fi
  done

  if [ "${failed}" -ne 0 ]; then
    # cleanup_work_dir reports where the logs are kept.
    log_error "Failed to ${what}."
    exit 1
  fi
}

# ==============================================================================
# Orchestration

main() {
  trap cleanup_work_dir EXIT

  parse_args "$@"

  local instances=()
  local i
  for ((i = 1; i <= COUNT; i++)); do
    instances+=("$(printf '%s-%02d' "${INSTANCE_PREFIX}" "${i}")")
  done

  if [ "${NUKE}" -eq 1 ]; then
    nuke_instances "${instances[@]}"
    return $?
  fi

  if [ ! -x "${CLI_BIN}" ]; then
    log_error "Cluster Manager binary not found at ${CLI_BIN}. Build it with 'make build'."
    exit 1
  fi

  log "Loading the development environment from scripts/setup-env.sh"
  load_dev_env

  # The CLI advertises the Cluster Connector address from its own configuration, so the
  # script has to use the same address for the host entry and the reachability check.
  CLUSTER_CONNECTOR_DOMAIN="${CLUSTER_CONNECTOR_DOMAIN:-cc.lxd-cm.local}"
  CLUSTER_CONNECTOR_PORT="${CLUSTER_CONNECTOR_PORT:-9000}"

  local address="${CLUSTER_CONNECTOR_DOMAIN}:${CLUSTER_CONNECTOR_PORT}"
  log "Checking that the Cluster Connector is reachable at https://${address}"
  if ! curl --silent --insecure --max-time 5 --output /dev/null "https://${address}/"; then
    log_error "Cannot reach the Cluster Connector at https://${address}. Start the development environment with 'make dev'."
    exit 1
  fi

  log "Checking that the Cluster Connector uses the certificate of this checkout"
  check_connector_certificate

  if ! ensure_project; then
    exit 1
  fi

  WORK_DIR="$(mktemp -d -t lxd-container-setup-XXXXXX)"

  local name
  for name in "${instances[@]}"; do
    if ! create_instance "${name}"; then
      exit 1
    fi
  done

  for name in "${instances[@]}"; do
    if ! setup_instance_hosts "${name}"; then
      exit 1
    fi
  done

  run_parallel "install the snaps" install_snaps "${instances[@]}"
  run_parallel "initialize the MicroCloud clusters" initialize_cluster "${instances[@]}"
  run_parallel "configure the LXD API addresses and cluster certificates" configure_instance_cluster_certificate "${instances[@]}"
  run_parallel "configure the OIDC settings in the LXD of the clusters" configure_instance_oidc "${instances[@]}"
  run_parallel "map the identity provider group to the LXD admin group" configure_instance_admin_group "${instances[@]}"

  log "Enrolling the clusters with the Cluster Manager"

  # The join tokens are stored in the Cluster Manager database, so enroll one at a time.
  for name in "${instances[@]}"; do
    if ! enroll_cluster "${name}"; then
      exit 1
    fi
  done

  log "${#instances[@]} cluster(s) enrolled with the Cluster Manager:"
  for name in "${instances[@]}"; do
    echo "    - ${name}"
  done
}

main "$@"
