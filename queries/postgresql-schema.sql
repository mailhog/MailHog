SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

CREATE TABLE IF NOT EXISTS messages (
    message jsonb
);

CREATE INDEX IF NOT EXISTS messages_expr_idx ON messages USING btree (((message -> 'Created'::text)));
CREATE INDEX IF NOT EXISTS messages_expr_idx2 ON messages USING btree (((message ->> 'ID'::text)));

-- TODO: can we make this indexes better by using "Content" (parsed body) rather than "Raw" (SMTP wire protocol)?
CREATE INDEX IF NOT EXISTS messages_expr_idx3 ON messages USING gin (to_tsvector('english'::regconfig, ((message -> 'Raw'::text) ->> 'Data'::text)));
CREATE INDEX IF NOT EXISTS messages_expr_idx4 ON messages USING gin (to_tsvector('english'::regconfig, ((message -> 'Raw'::text) ->> 'From'::text)));
CREATE INDEX IF NOT EXISTS messages_expr_idx5 ON messages USING gin (to_tsvector('english'::regconfig, ((message -> 'Raw'::text) ->> 'To'::text)));
