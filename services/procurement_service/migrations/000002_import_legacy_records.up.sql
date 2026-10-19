-- Preserve data from the previous generic-record prototype, when present.
DO $$ BEGIN
IF to_regclass('records') IS NOT NULL THEN
INSERT INTO suppliers(id,created_at,updated_at,name,email,lead_days,active,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'name')::TEXT,(payload->>'email')::TEXT,(payload->>'lead_days')::INTEGER,(payload->>'active')::BOOLEAN,owner FROM records WHERE kind='supplier' ON CONFLICT DO NOTHING;
INSERT INTO orders(id,created_at,updated_at,product_id,supplier_id,quantity,unit_price,currency,status,expected_at,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'product_id')::UUID,(payload->>'supplier_id')::UUID,(payload->>'quantity')::BIGINT,(payload->>'unit_price')::BIGINT,(payload->>'currency')::TEXT,(payload->>'status')::TEXT,(payload->>'expected_at')::TIMESTAMPTZ,owner FROM records WHERE kind='order' ON CONFLICT DO NOTHING;
END IF;
END $$;
