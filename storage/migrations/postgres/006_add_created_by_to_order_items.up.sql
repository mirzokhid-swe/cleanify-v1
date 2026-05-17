ALTER TABLE "order_items"
ADD COLUMN IF NOT EXISTS "created_by" BIGINT;
