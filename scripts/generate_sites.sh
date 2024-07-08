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
SITE_DETAILS_INSERT="INSERT INTO site_details (core_site_id, status, cpu_total_count, cpu_load_1, cpu_load_5, cpu_load_15, memory_total_amount, memory_usage, disk_total_size, disk_usage, instance_count, instance_statuses, member_count, member_statuses, updated_at) VALUES "
SITE_DETAILS_VALUES=()

for i in $(seq 1 $ENTRIES); do
    SEED=$((i * 12345))
    CORE_SITE_ID=$(( (i % $ENTRIES) + 1 ))
    STATUS="PENDING_APPROVAL"
    if [ $((i % 2)) -eq 0 ]; then
        STATUS="ACTIVE"
    fi
    INSTANCE_COUNT=200
    MEMBER_COUNT=5
    CPU_TOTAL_COUNT=10
    CPU_LOAD_1="0.25"
    CPU_LOAD_5="0.50"
    CPU_LOAD_15="0.75"
    MEMORY_TOTAL_AMOUNT=16000000000 #16GB
    MEMORY_USAGE=$(( (SEED * 12345) % (MEMORY_TOTAL_AMOUNT + 1) ))  # 0 to MEMORY_TOTAL_AMOUNT
    DISK_TOTAL_SIZE=2000000000000 #2TB
    DISK_USAGE=$(( (SEED * 12345678) % (DISK_TOTAL_SIZE + 1) ))  # 0 to DISK_TOTAL_SIZE
    # Generates a random date and time between date.now and 10 minutes ago.
    LAST_UPDATED_AT=$(date -u -d "@$(( $(date -u -d '10 minutes ago' +%s) + SEED % ($(date -u +%s) - $(date -u -d '10 minutes ago' +%s)) ))" +'%Y-%m-%d %H:%M:%S')
    INSTANCE_STATUSES=$(printf "$INSTANCE_STATUSES_TEMPLATE" $((SEED % 200 + 1)))
    MEMBER_STATUSES=$(printf "$MEMBER_STATUSES_TEMPLATE" $((SEED % 4)))
    SITE_DETAILS_VALUES+=("($CORE_SITE_ID, '$STATUS', '$CPU_TOTAL_COUNT', '$CPU_LOAD_1', '$CPU_LOAD_5', '$CPU_LOAD_15', '$MEMORY_TOTAL_AMOUNT', '$MEMORY_USAGE', '$DISK_TOTAL_SIZE', '$DISK_USAGE', '$INSTANCE_COUNT', '$INSTANCE_STATUSES', '$MEMBER_COUNT', '$MEMBER_STATUSES', '$LAST_UPDATED_AT')")
done

# Combine the values into a single insert statement
SITE_DETAILS_INSERT+=$(IFS=,; echo "${SITE_DETAILS_VALUES[*]}")";"

# if flag --local exists then use go run to run site-mgr
if [ "$1" == "--local" ]; then
    COMMAND="go run ./cmd/lxd-site-mgr"
else
    COMMAND="./lxd-site-mgr"
fi

# Execute the combined SQL commands
$COMMAND --state-dir ./state/dir1 sql "
    $CORE_SITES_INSERT
    $SITE_DETAILS_INSERT
"
