#!/usr/bin/env bash
set -e

ENTRIES=200

# Instance and member statuses templates
INSTANCE_STATUSES_TEMPLATE='[{"status": "Running", "count": %d}, {"status": "Stopped", "count": %d}, {"status": "Frozen", "count": %d}, {"status": "Error", "count": %d}]'
MEMBER_STATUSES_TEMPLATE='[{"status": "Online", "count": %d}, {"status": "Offline", "count": %d}, {"status": "Evacuated", "count": %d}, {"status": "Blocked", "count": %d}]'

# Prepare bulk insert statements for core_remote_clusters
CORE_REMOTE_CLUSTERS_INSERT="INSERT INTO core_remote_clusters (name, cluster_certificate) VALUES "
CORE_REMOTE_CLUSTERS_VALUES=()

for i in $(seq 1 $ENTRIES); do
    REMOTE_CLUSTER_NAME="cluster_$i"
    CLUSTER_CERTIFICATE="cert_$i"
    CORE_REMOTE_CLUSTERS_VALUES+=("('$REMOTE_CLUSTER_NAME', '$CLUSTER_CERTIFICATE')")
done

# Combine the values into a single insert statement
CORE_REMOTE_CLUSTERS_INSERT+=$(IFS=,; echo "${CORE_REMOTE_CLUSTERS_VALUES[*]}")";"

# Prepare bulk insert statements for remote_cluster_details
REMOTE_CLUSTER_DETAILS_INSERT="INSERT INTO remote_cluster_details (core_remote_cluster_id, status, cpu_total_count, cpu_load_1, cpu_load_5, cpu_load_15, memory_total_amount, memory_usage, disk_total_size, disk_usage, instance_count, instance_statuses, member_count, member_statuses, updated_at) VALUES "
REMOTE_CLUSTER_DETAILS_VALUES=()

for i in $(seq 1 $ENTRIES); do
    SEED=$((i * 12345))
    CORE_REMOTE_CLUSTER_ID=$(( (i % $ENTRIES) + 1 ))
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
    # Generates a random date and time between date.now and 30 minutes ago.
    LAST_UPDATED_AT=$(date -u -d "@$(( $(date -u -d '30 minutes ago' +%s) + SEED % ($(date -u +%s) - $(date -u -d '30 minutes ago' +%s)) ))" +'%Y-%m-%d %H:%M:%S')
    INSTANCE_STATUSES=$(printf "$INSTANCE_STATUSES_TEMPLATE" $((SEED % 200 + 1)) $((SEED % 150 + 1)) $((SEED % 100 + 1)) $((SEED % 50 + 1)))
    MEMBER_STATUSES=$(printf "$MEMBER_STATUSES_TEMPLATE" $((SEED % 4 + 1)) $((SEED % 3 + 1)) $((SEED % 2 + 1)) $((SEED % 1 + 1)))
    REMOTE_CLUSTER_DETAILS_VALUES+=("($CORE_REMOTE_CLUSTER_ID, '$STATUS', '$CPU_TOTAL_COUNT', '$CPU_LOAD_1', '$CPU_LOAD_5', '$CPU_LOAD_15', '$MEMORY_TOTAL_AMOUNT', '$MEMORY_USAGE', '$DISK_TOTAL_SIZE', '$DISK_USAGE', '$INSTANCE_COUNT', '$INSTANCE_STATUSES', '$MEMBER_COUNT', '$MEMBER_STATUSES', '$LAST_UPDATED_AT')")
done

# Combine the values into a single insert statement
REMOTE_CLUSTER_DETAILS_INSERT+=$(IFS=,; echo "${REMOTE_CLUSTER_DETAILS_VALUES[*]}")";"

# if flag --local exists then use go run to run lxd-cluster-mgr
if [ "$1" == "--local" ]; then
    COMMAND="go run ./cmd/lxd-cluster-mgr"
else
    COMMAND="./lxd-cluster-mgr"
fi

# Execute the combined SQL commands
$COMMAND --state-dir ./state/dir1 sql "
    $CORE_REMOTE_CLUSTERS_INSERT
    $REMOTE_CLUSTER_DETAILS_INSERT
"
