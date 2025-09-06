CREATE TYPE user_state AS ENUM ('waiting_for_channel', 'waiting_for_message');

CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  telegram_id BIGINT UNIQUE NOT NULL,
  user_state user_state DEFAULT 'waiting_for_channel',
  channelIDs BIGINT[] DEFAULT []
);

CREATE TABLE IF NOT EXISTS scheduled_posts (
  id SERIAL PRIMARY KEY,
  chat_id BIGINT NOT NULL,
  from_chat_id BIGINT NOT NULL,
  message_id INT NOT NULL,
  send_at TIMESTAMP NOT NULL
)