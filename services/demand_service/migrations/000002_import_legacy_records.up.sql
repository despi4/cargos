-- Preserve data from the previous generic-record prototype, when present.
DO $$ BEGIN
IF to_regclass('records') IS NOT NULL THEN
INSERT INTO demand_records(id,created_at,updated_at,product_id,quantity,occurred_at,note,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'product_id')::UUID,(payload->>'quantity')::BIGINT,(payload->>'occurred_at')::TIMESTAMPTZ,(payload->>'note')::TEXT,owner FROM records WHERE kind='demand' ON CONFLICT DO NOTHING;
END IF;
END $$;
