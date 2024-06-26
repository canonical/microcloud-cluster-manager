#! /usr/bin/env bash
set -e

go run ./cmd/lxd-site-mgr --state-dir ./state/dir1 sql "
    INSERT INTO core_sites (name, site_certificate) VALUES ('site1', 'a');
    INSERT INTO core_sites (name, site_certificate) VALUES ('site2', 'as');
    INSERT INTO core_sites (name, site_certificate) VALUES ('site3', 'asd');

    INSERT INTO site_details (core_site_id, status, instance_statuses) VALUES (1, 'PENDING_APPROVAL', '{}');
    INSERT INTO site_details (core_site_id, status, instance_statuses) VALUES (2, 'ACTIVE', '{}');
    INSERT INTO site_details (core_site_id, status, instance_statuses) VALUES (3, 'ACTIVE', '{}');

    INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
        VALUES (1, 'member1', '127.0.0.1:9001', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
    INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
        VALUES (1, 'member2', '127.0.0.1:9002', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
    INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
        VALUES (2, 'member1', '127.0.0.1:9001', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
    INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
        VALUES (2, 'member2', '127.0.0.1:9002', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
    INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
        VALUES (3, 'member1', '127.0.0.1:9001', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
    INSERT INTO site_member_statuses (core_site_id, member_name, address, architecture, role, usage_cpu, usage_memory, usage_disk, status) 
        VALUES (3, 'member2', '127.0.0.1:9002', 'x86_64', 'controller', 0.0, 0.0, 0.0, 'ACTIVE');
"
