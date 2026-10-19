-- Preserve data from the previous generic-record prototype, when present.
DO $$ BEGIN
IF to_regclass('records') IS NOT NULL THEN
INSERT INTO notifications(id,created_at,updated_at,event_id,title,message,is_read,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'event_id')::UUID,(payload->>'title')::TEXT,(payload->>'message')::TEXT,(payload->>'read')::BOOLEAN,owner FROM records WHERE kind='notification' ON CONFLICT DO NOTHING;
END IF;
END $$;
