#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/helper/common.bash
load $BATS_TEST_DIRNAME/helper/query-server-common.bash

setup() {
    skiponwindows "Has dependencies that are missing on the Jenkins Windows installation."

    setup_common
    start_sql_server
}

teardown() {
    skiponwindows "Has dependencies that are missing on the Jenkins Windows installation."

    stop_sql_server
    teardown_common
}

@test "test basic querying via dolt sql-server" {
    skiponwindows "Has dependencies that are missing on the Jenkins Windows installation."

    server_query 1 "SHOW tables" ""
    server_query 1 "CREATE TABLE one_pk (
        pk BIGINT NOT NULL COMMENT 'tag:0',
        c1 BIGINT COMMENT 'tag:1',
        c2 BIGINT COMMENT 'tag:2',
        PRIMARY KEY (pk)
    )" ""
    server_query 1 "SHOW tables" "Table\none_pk"
    insert_query 1 "INSERT INTO one_pk (pk) values (0)"
    server_query 1 "SELECT * from one_pk ORDER BY pk" "pk,c1,c2\n0,None,None"
    insert_query 1 "INSERT INTO one_pk (pk,c1) values (1,1)"
    insert_query 1 "INSERT INTO one_pk (pk,c1,c2) values (2,2,2),(3,3,3)"
    server_query 1 "SELECT * from one_pk ORDER by pk" "pk,c1,c2\n0,None,None\n1,1,None\n2,2,2\n3,3,3"
    update_query 1 "UPDATE one_pk SET c2=c1 WHERE c2 is NULL and c1 IS NOT NULL"
}