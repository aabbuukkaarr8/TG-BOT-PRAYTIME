CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       chat_id BIGINT NOT NULL UNIQUE,
                       fajr TEXT DEFAULT TRUE,
                       zuhr TEXT DEFAULT TRUE,
                       asr TEXT DEFAULT TRUE,
                       maghrib TEXT DEFAULT TRUE,
                       isha TEXT DEFAULT TRUE,
                       status TEXT DEFAULT 'on'
);
