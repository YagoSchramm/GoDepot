FROM postgres:16-alpine

COPY script/migrate/001-create-tables.up.sql /docker-entrypoint-initdb.d/001-create-tables.sql
