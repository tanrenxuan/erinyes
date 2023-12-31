#!/bin/bash

DB_USER="root"
DB_PASSWORD="123456"
DB_HOST="localhost"
DB_PORT="3306"
DB_NAME="erinyes"

DB_CONNECTION_STRING="-u$DB_USER -p$DB_PASSWORD -h$DB_HOST -P$DB_PORT $DB_NAME"
TABLES=$(mysql $DB_CONNECTION_STRING -e "SHOW TABLES" | tail -n +2)

# 清空每个表的数据
for TABLE in $TABLES; do
    echo "Truncating table: $TABLE"
    mysql $DB_CONNECTION_STRING -e "TRUNCATE TABLE $TABLE"
done

echo "Done"