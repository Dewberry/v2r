-- Create the database
CREATE DATABASE dev;

-- Add extensions
CREATE EXTENSION postgis;

-- Create the schema
CREATE SCHEMA IF NOT EXISTS sandbox;

-- DROP TABLE IF EXISTS dev.sandbox;
CREATE TABLE IF NOT EXISTS sandbox.location_1(
    uid SERIAL PRIMARY KEY,
    elevation float NOT NULL,
    geom geometry(Point, 2284)
);

-- Insert sample data
INSERT INTO sandbox.location_1(elevation, geom) 
VALUES 
    (0.0, ST_GeomFromText('POINT(12031543.049036839976907 3953694.544257536064833)', 2284)),
    (0.0, ST_GeomFromText('POINT(12039094.668343752622604 3942453.752100885380059)', 2284)),
    (0.0, ST_GeomFromText('POINT(12052295.156865464523435 3931480.047175889834762)', 2284)),
    (0.0, ST_GeomFromText('POINT(12076264.960092723369598 3917394.837044359184802)', 2284)),
    (0.0, ST_GeomFromText('POINT(12091339.669621860608459 3899012.514405147172511)', 2284)),
    (2.5, ST_GeomFromText('POINT(12020401.943588392809033 3946331.704970045480877)', 2284)),
    (2.9, ST_GeomFromText('POINT(12025601.603362886235118 3938077.554017631337047)', 2284)),
    (7.0, ST_GeomFromText('POINT(12034957.244500949978828 3920782.799525749869645)', 2284)),
    (6.3, ST_GeomFromText('POINT(12055886.38144482113421 3898699.015764322131872)' , 2284)),
    (2.0, ST_GeomFromText('POINT( 12074513.648989820852876 3885548.111378168221563)', 2284)),
    (3.1, ST_GeomFromText('POINT( 12049260.510355683043599 3952539.493517972063273)', 2284)),
    (2.8, ST_GeomFromText('POINT( 12051951.205032240599394 3947886.314968572929502)', 2284)),
    (3.9, ST_GeomFromText('POINT( 12096825.067172093316913 3926824.496812131721526)', 2284)),
    (4.6, ST_GeomFromText('POINT( 12080323.349286442622542 3941202.122561900410801)', 2284));





SELECT elevation, ST_X(geom), ST_Y(geom)
FROM sandbox.location_1;
