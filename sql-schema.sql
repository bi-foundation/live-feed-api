CREATE TABLE IF NOT EXISTS subscriptions (
	id SERIAL PRIMARY KEY,
	callback VARCHAR(2083) NOT NULL,
	callback_type VARCHAR(20) NOT NULL
);

CREATE TABLE IF NOT EXISTS filters (
	id SERIAL PRIMARY KEY,
	subscription BIGINT(20) REFERENCES subscriptions(id),
	event_type VARCHAR(20) NOT NULL,
	filtering TEXT
);
