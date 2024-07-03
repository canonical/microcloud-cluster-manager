#!/usr/bin/env bash
set -e

ENTRIES=200

# Instance and member statuses templates
INSTANCE_STATUSES_TEMPLATE='[{"status": "running", "count": %d}]'
MEMBER_STATUSES_TEMPLATE='[{"status": "active", "count": %d}]'

# Prepare bulk insert statements for core_sites
CORE_SITES_INSERT="INSERT INTO core_sites (name, site_certificate) VALUES "
CORE_SITES_VALUES=()

for i in $(seq 1 $ENTRIES); do
    SITE_NAME="site_$i"
    SITE_CERTIFICATE="cert_$i"
    CORE_SITES_VALUES+=("('$SITE_NAME', '$SITE_CERTIFICATE')")
done

# Combine the values into a single insert statement
CORE_SITES_INSERT+=$(IFS=,; echo "${CORE_SITES_VALUES[*]}")";"

# Prepare bulk insert statements for site_details
SITE_DETAILS_INSERT="INSERT INTO site_details (core_site_id, status, instance_statuses, member_statuses) VALUES "
SITE_DETAILS_VALUES=()

for i in $(seq 1 $ENTRIES); do
    CORE_SITE_ID=$(( (i % $ENTRIES) + 1 ))
    STATUS="PENDING_APPROVAL"
    if [ $((i % 2)) -eq 0 ]; then
        STATUS="ACTIVE"
    fi
    INSTANCE_COUNT=$(( i ))
    MEMBER_COUNT=$(( i ))
    INSTANCE_STATUSES=$(printf "$INSTANCE_STATUSES_TEMPLATE" "$INSTANCE_COUNT")
    MEMBER_STATUSES=$(printf "$MEMBER_STATUSES_TEMPLATE" "$MEMBER_COUNT")
    SITE_DETAILS_VALUES+=("($CORE_SITE_ID, '$STATUS', '$INSTANCE_STATUSES', '$MEMBER_STATUSES')")
done

# Combine the values into a single insert statement
SITE_DETAILS_INSERT+=$(IFS=,; echo "${SITE_DETAILS_VALUES[*]}")";"

# Execute the combined SQL commands
go run ./cmd/lxd-site-mgr --state-dir ./state/dir1 sql "
    $CORE_SITES_INSERT
    $SITE_DETAILS_INSERT
"
