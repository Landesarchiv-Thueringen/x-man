create user xman WITH PASSWORD 'test1234';;
drop database if exists xman;
create database xman owner xman;
\c xman
create extension if not exists "uuid-ossp";