package main

var schema = `
CREATE TABLE addresses (
	address	varchar PRIMARY KEY,	
	analyzed BOOLEAN NOT NULL,
	email varchar
);
`
