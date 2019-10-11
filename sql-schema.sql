DROP TABLE IF EXISTS `filters`;
DROP TABLE IF EXISTS `subscriptions`;

CREATE TABLE IF NOT EXISTS subscriptions (
	id SERIAL PRIMARY KEY,
	failures int NOT NULL,
	callback VARCHAR(2083) NOT NULL,
	callback_type VARCHAR(25) NOT NULL,
	status VARCHAR(20) NOT NULL,
	info VARCHAR(200),
	access_token VARCHAR(255),
	username VARCHAR(255),
	password VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS filters (
	id SERIAL PRIMARY KEY,
	subscription BIGINT(20) REFERENCES subscriptions(id),
	event_type VARCHAR(25) NOT NULL,
	filtering TEXT
);
