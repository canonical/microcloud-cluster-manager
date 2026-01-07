package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/canonical/microcloud-cluster-manager/internal/pkg/api/models/v1"
	"github.com/canonical/microcloud-cluster-manager/test/helpers"
)

func testRemoteClusterSuccess(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join and status updates under normal conditions", func(t *testing.T) {
		remoteClusterName := helpers.GetRandomName("remote_cluster_e2e")
		var condition string
		var err error
		var tokenData models.RemoteClusterTokenBody

		{
			condition = "Should be able to create token with valid data"
			tokenData, err = helpers.CreateAndReturnRemoteClusterJoinToken(env, remoteClusterName, time.Time{})
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.ServerName == "" {
				err = fmt.Errorf("invalid server_name")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Fingerprint == "" {
				err = fmt.Errorf("invalid fingerprint")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Addresses[0] != env.ClusterConnectorHostPort() {
				err = fmt.Errorf("invalid address")
				helpers.LogTestOutcome(t, condition, err)
			}

			if tokenData.Secret == "" {
				err = fmt.Errorf("invalid secret")
				helpers.LogTestOutcome(t, condition, err)
			}

			if time.Time.Equal(tokenData.ExpiresAt, time.Time{}) {
				err = fmt.Errorf("invalid expiry")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to receive a join request"
			err = helpers.SendJoinRequest(env, tokenData)
			helpers.LogTestOutcome(t, condition, err)
		}

		{
			condition = "Should be able to get remote cluster with ACTIVE status"
			remoteCluster, err := helpers.FindRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.Status != string(models.ACTIVE) {
				err = fmt.Errorf("invalid remote cluster status")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should have deleted the remote cluster join token after receiving join request"
			token, err := helpers.FindToken(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if token != (models.RemoteClusterToken{}) {
				err = fmt.Errorf("token not deleted")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to receive a status update"
			input := helpers.CreateStatusPostData()
			response, err := helpers.SendStatusUpdate(env, tokenData, input)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			expected := env.ClusterConnectorHostPort()

			if !reflect.DeepEqual(response.ClusterManagerAddress, expected) {
				fmt.Println(response.ClusterManagerAddress)
				fmt.Println(expected)
				err = fmt.Errorf("invalid Cluster Manager address")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		{
			condition = "Should be able to get remote cluster status"
			remoteCluster, err := helpers.FindRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.CPUTotalCount != 8 {
				err = fmt.Errorf("invalid CPU total count")
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.MemoryTotalAmount != 1024 {
				err = fmt.Errorf("invalid memory total amount")
				helpers.LogTestOutcome(t, condition, err)
			}

			if remoteCluster.DiskTotalSize != 1024 {
				err = fmt.Errorf("invalid disk total size")
				helpers.LogTestOutcome(t, condition, err)
			}

			if !reflect.DeepEqual(remoteCluster.InstanceStatuses, []models.StatusDistribution{
				{Status: "running", Count: 1},
				{Status: "stopped", Count: 2},
			}) {
				err = fmt.Errorf("invalid instance statuses")
				helpers.LogTestOutcome(t, condition, err)
			}

			if !reflect.DeepEqual(remoteCluster.MemberStatuses, []models.StatusDistribution{
				{Status: "active", Count: 1},
				{Status: "inactive", Count: 2},
			}) {
				err = fmt.Errorf("invalid member statuses")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}

		env.RemoveRemoteClusterToken(remoteClusterName)
		env.RemoveRemoteCluster(remoteClusterName)
	}
}

func testRemoteClusterSuccessWithMetrics(env *helpers.Environment) (testName string, testFunc func(t *testing.T)) {
	return "lxd remote cluster join and status updates with metrics from multiple servers under normal conditions", func(t *testing.T) {
		remoteClusterName := helpers.GetRandomName("remote_cluster_e2e_with_metrics")
		var condition string
		{
			condition = "Should be able to receive a status update with metrics from multiple servers"

			tokenData, err := helpers.RegisterRemoteCluster(env, remoteClusterName)
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}
			instanceMember1 := helpers.GetRandomName("e2e-random-instance-member-1")
			metricsMember1 := models.ServerMetrics{
				Member:  "member1",
				Service: "LXD",
				Metrics: getServerMetricsWithInstance(instanceMember1),
			}

			instanceMember2 := helpers.GetRandomName("e2e-random-instance-member-2")
			metricsMember2 := models.ServerMetrics{
				Member:  "member2",
				Service: "LXD",
				Metrics: getServerMetricsWithInstance(instanceMember2),
			}

			input := helpers.CreateStatusPostData()
			input.ServerMetrics = []models.ServerMetrics{
				metricsMember1,
				metricsMember2,
			}

			_, err = helpers.SendStatusUpdate(env, *tokenData, input)

			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}
			// Allow time for Prometheus to ingest metrics
			time.Sleep(2 * time.Second)
			prometheusResponse, err := helpers.QueryPrometheus(env, "lxd_cpu_seconds_total")
			if err != nil {
				helpers.LogTestOutcome(t, condition, err)
			}

			if !strings.Contains(prometheusResponse, instanceMember1) {
				err = fmt.Errorf("instance from cluster member1 not found in Prometheus")
				helpers.LogTestOutcome(t, condition, err)
			}

			if !strings.Contains(prometheusResponse, instanceMember2) {
				err = fmt.Errorf("instance from cluster member2 not found in Prometheus")
				helpers.LogTestOutcome(t, condition, err)
			}

			helpers.LogTestOutcome(t, condition, nil)
		}
	}
}

func getServerMetricsWithInstance(instanceName string) string {
	metricsText := fmt.Sprintf(`# HELP lxd_cpu_seconds_total Total CPU time used
			# TYPE lxd_cpu_seconds_total gauge
			lxd_cpu_seconds_total{instance="inst1",project="default"} 100.0
			lxd_cpu_seconds_total{instance="inst2",project="default"} 200.0
			lxd_cpu_seconds_total{instance="%s",project="other"} 300.0
			# HELP http_requests_total Total HTTP requests
			# TYPE http_requests_total counter
			http_requests_total{method="GET",status="200"} 1024
			# HELP lxd_memory_bytes Memory usage
			# TYPE lxd_memory_bytes gauge
			lxd_memory_bytes{instance="inst1"} 2048.0
			`, instanceName)
	return metricsText
}
