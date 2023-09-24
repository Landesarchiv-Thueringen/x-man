create user lath_xdomea;
drop database if exists lath_xdomea;
create database lath_xdomea owner lath_xdomea;
grant all privileges on database lath_xdomea to postgres;
\c lath_xdomea
create extension if not exists "uuid-ossp";