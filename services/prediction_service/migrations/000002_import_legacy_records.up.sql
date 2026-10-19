-- Preserve data from the previous generic-record prototype, when present.
DO $$ BEGIN
IF to_regclass('records') IS NOT NULL THEN
INSERT INTO predictions(id,created_at,updated_at,window_days,recommendations,notification_sent,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'window_days')::INTEGER,payload->'recommendations',(payload->>'notification_sent')::BOOLEAN,owner FROM records WHERE kind='prediction' ON CONFLICT DO NOTHING;
END IF;
END $$;
