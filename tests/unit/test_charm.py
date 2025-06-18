from charm import ClusterManagerCharm
from ops import testing


def test_pebble_layer_no_relations():
    ctx = testing.Context(ClusterManagerCharm)
    container = testing.Container(name='microcloud-cluster-manager', can_connect=True)
    state_in = testing.State(
        containers={container},
        leader=True,
    )
    state_out = ctx.run(ctx.on.pebble_ready(container), state_in)
    # Expected plan after Pebble ready with default config
    expected_plan = {}

    # Check that we have the plan we expected:
    assert state_out.get_container(container.name).plan == expected_plan
    # Check the unit status is blocked
    assert state_out.unit_status == testing.BlockedStatus('certificates integration not created')

def test_pebble_layer_with_certificates_relation():
    ctx = testing.Context(ClusterManagerCharm)
    container = testing.Container(name='microcloud-cluster-manager', can_connect=True)
    certificate_relation = testing.Relation('certificates', remote_app_data={'1': '2'})
    state_in = testing.State(
        containers={container},
        leader=True,
        relations=[certificate_relation]
    )
    state_out = ctx.run(ctx.on.pebble_ready(container), state_in)
    # Expected plan after Pebble ready with default config
    expected_plan = {}

    # Check that we have the plan we expected:
    assert state_out.get_container(container.name).plan == expected_plan
    assert state_out.unit_status == testing.BlockedStatus('Waiting for database relation')

def test_pebble_layer_with_database_relation():
    ctx = testing.Context(ClusterManagerCharm)
    container = testing.Container(name='microcloud-cluster-manager', can_connect=True)
    certificate_relation = testing.Relation('certificates', remote_app_data={'1': '2'})
    database_relation = testing.Relation('database', remote_app_data={'1': '2'})
    state_in = testing.State(
        containers={container},
        leader=True,
        relations=[certificate_relation, database_relation]
    )
    state_out = ctx.run(ctx.on.pebble_ready(container), state_in)
    # Expected plan after Pebble ready with default config
    expected_plan = {}

    # Check that we have the plan we expected:
    assert state_out.get_container(container.name).plan == expected_plan
    assert state_out.unit_status == testing.MaintenanceStatus('Waiting for Pebble in workload container')
