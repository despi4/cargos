-- Preserve data from the previous generic-record prototype, when present.
DO $$ BEGIN
IF to_regclass('records') IS NOT NULL THEN
INSERT INTO products(id,created_at,updated_at,sku,name,unit,stock,min_stock,target_stock,active,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'sku')::TEXT,(payload->>'name')::TEXT,(payload->>'unit')::TEXT,(payload->>'stock')::BIGINT,(payload->>'min_stock')::BIGINT,(payload->>'target_stock')::BIGINT,(payload->>'active')::BOOLEAN,owner FROM records WHERE kind='product' ON CONFLICT DO NOTHING;
INSERT INTO movements(id,created_at,updated_at,product_id,quantity,reason,reference,balance,owner_id) SELECT id,(payload->>'created_at')::timestamptz,(payload->>'updated_at')::timestamptz,(payload->>'product_id')::UUID,(payload->>'quantity')::BIGINT,(payload->>'reason')::TEXT,(payload->>'reference')::TEXT,(payload->>'balance')::BIGINT,owner FROM records WHERE kind='movement' ON CONFLICT DO NOTHING;
END IF;
END $$;
