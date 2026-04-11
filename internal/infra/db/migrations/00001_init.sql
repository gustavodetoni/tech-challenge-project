-- +goose Up
-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM (
  'ADMIN',
  'MANAGER',
  'MECHANIC',
  'ATTENDANT',
  'VIEWER'
);

CREATE TYPE client_document_type AS ENUM (
  'CPF',
  'CNPJ'
);

CREATE TYPE service_order_status AS ENUM (
  'RECEIVED',
  'IN_DIAGNOSIS',
  'WAITING_APPROVAL',
  'IN_PROGRESS',
  'FINISHED',
  'DELIVERED',
  'CANCELED'
);

CREATE TYPE budget_status AS ENUM (
  'DRAFT',
  'SENT',
  'APPROVED',
  'REJECTED',
  'EXPIRED',
  'CANCELED'
);

CREATE TYPE service_order_item_status AS ENUM (
  'PENDING',
  'APPROVED',
  'IN_PROGRESS',
  'COMPLETED',
  'CONSUMED',
  'CANCELED'
);

CREATE TYPE stock_movement_type AS ENUM (
  'IN',
  'OUT',
  'ADJUSTMENT'
);

CREATE TYPE stock_reference_type AS ENUM (
  'SERVICE_ORDER',
  'MANUAL',
  'PURCHASE'
);

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(120) NOT NULL,
  email varchar(240) NOT NULL,
  password_hash varchar(255) NOT NULL,
  role user_role NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS users_email_uq
  ON users (lower(email))
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS clients (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  document_type client_document_type NOT NULL,
  document_number varchar(20) NOT NULL,
  name varchar(120) NOT NULL,
  email varchar(240),
  phone varchar(40),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS clients_document_number_uq
  ON clients (document_number)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS clients_name_idx
  ON clients (name)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS vehicles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id uuid NOT NULL REFERENCES clients(id),
  plate varchar(10) NOT NULL,
  brand varchar(60) NOT NULL,
  model varchar(60) NOT NULL,
  manufacture_year smallint,
  model_year smallint NOT NULL,
  color varchar(30),
  mileage int,
  chassis varchar(40),
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT vehicles_mileage_ck CHECK (mileage IS NULL OR mileage >= 0),
  CONSTRAINT vehicles_manufacture_year_ck CHECK (
    manufacture_year IS NULL OR manufacture_year BETWEEN 1900 AND 2100
  ),
  CONSTRAINT vehicles_model_year_ck CHECK (
    model_year BETWEEN 1900 AND 2100
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS vehicles_plate_uq
  ON vehicles (plate)
  WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS vehicles_chassis_uq
  ON vehicles (chassis)
  WHERE deleted_at IS NULL AND chassis IS NOT NULL;

CREATE INDEX IF NOT EXISTS vehicles_client_id_idx
  ON vehicles (client_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS services (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(120) NOT NULL,
  description text,
  base_price_cents bigint NOT NULL DEFAULT 0,
  estimated_minutes int NOT NULL DEFAULT 0,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT services_base_price_cents_ck CHECK (base_price_cents >= 0),
  CONSTRAINT services_estimated_minutes_ck CHECK (estimated_minutes >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS services_name_uq
  ON services (name)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS parts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  sku varchar(60) NOT NULL,
  name varchar(120) NOT NULL,
  description text,
  unit_price_cents bigint NOT NULL DEFAULT 0,
  stock_quantity int NOT NULL DEFAULT 0,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT parts_unit_price_cents_ck CHECK (unit_price_cents >= 0),
  CONSTRAINT parts_stock_quantity_ck CHECK (stock_quantity >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS parts_sku_uq
  ON parts (sku)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS parts_name_idx
  ON parts (name)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS service_orders (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  code varchar(30) NOT NULL,
  client_id uuid NOT NULL REFERENCES clients(id),
  vehicle_id uuid NOT NULL REFERENCES vehicles(id),
  assigned_user_id uuid REFERENCES users(id),
  status service_order_status NOT NULL,
  customer_complaint text,
  diagnosis_notes text,
  internal_notes text,
  mileage_in int,
  mileage_out int,
  opened_at timestamptz NOT NULL DEFAULT now(),
  diagnosis_started_at timestamptz,
  waiting_approval_at timestamptz,
  execution_started_at timestamptz,
  finished_at timestamptz,
  delivered_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT service_orders_mileage_in_ck CHECK (mileage_in IS NULL OR mileage_in >= 0),
  CONSTRAINT service_orders_mileage_out_ck CHECK (mileage_out IS NULL OR mileage_out >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS service_orders_code_uq
  ON service_orders (code)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS service_orders_client_id_idx
  ON service_orders (client_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS service_orders_vehicle_id_idx
  ON service_orders (vehicle_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS service_orders_status_idx
  ON service_orders (status)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS service_orders_assigned_user_id_idx
  ON service_orders (assigned_user_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS service_order_status_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  service_order_id uuid NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
  from_status service_order_status,
  to_status service_order_status NOT NULL,
  changed_by_user_id uuid REFERENCES users(id),
  reason text,
  changed_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS so_status_history_service_order_id_idx
  ON service_order_status_history (service_order_id);

CREATE INDEX IF NOT EXISTS so_status_history_changed_at_idx
  ON service_order_status_history (changed_at);

CREATE TABLE IF NOT EXISTS budgets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  service_order_id uuid NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
  version int NOT NULL DEFAULT 1,
  status budget_status NOT NULL,
  total_amount_cents bigint NOT NULL DEFAULT 0,
  notes text,
  sent_at timestamptz,
  decided_at timestamptz,
  approved_at timestamptz,
  rejected_at timestamptz,
  approved_by_name varchar(120),
  rejection_reason text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT budgets_version_ck CHECK (version >= 1),
  CONSTRAINT budgets_total_amount_cents_ck CHECK (total_amount_cents >= 0)
);

CREATE UNIQUE INDEX IF NOT EXISTS budgets_service_order_version_uq
  ON budgets (service_order_id, version)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS budgets_service_order_id_idx
  ON budgets (service_order_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS budgets_status_idx
  ON budgets (status)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS budget_services (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  budget_id uuid NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
  service_id uuid REFERENCES services(id),
  description varchar(255) NOT NULL,
  quantity int NOT NULL DEFAULT 1,
  unit_price_cents bigint NOT NULL DEFAULT 0,
  total_price_cents bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT budget_services_quantity_ck CHECK (quantity > 0),
  CONSTRAINT budget_services_unit_price_cents_ck CHECK (unit_price_cents >= 0),
  CONSTRAINT budget_services_total_price_cents_ck CHECK (total_price_cents >= 0)
);

CREATE INDEX IF NOT EXISTS budget_services_budget_id_idx
  ON budget_services (budget_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS budget_services_service_id_idx
  ON budget_services (service_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS budget_parts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  budget_id uuid NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
  part_id uuid REFERENCES parts(id),
  description varchar(255) NOT NULL,
  quantity int NOT NULL DEFAULT 1,
  unit_price_cents bigint NOT NULL DEFAULT 0,
  total_price_cents bigint NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT budget_parts_quantity_ck CHECK (quantity > 0),
  CONSTRAINT budget_parts_unit_price_cents_ck CHECK (unit_price_cents >= 0),
  CONSTRAINT budget_parts_total_price_cents_ck CHECK (total_price_cents >= 0)
);

CREATE INDEX IF NOT EXISTS budget_parts_budget_id_idx
  ON budget_parts (budget_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS budget_parts_part_id_idx
  ON budget_parts (part_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS service_order_services (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  service_order_id uuid NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
  service_id uuid REFERENCES services(id),
  budget_service_id uuid REFERENCES budget_services(id),
  description varchar(255) NOT NULL,
  quantity int NOT NULL DEFAULT 1,
  unit_price_cents bigint NOT NULL DEFAULT 0,
  total_price_cents bigint NOT NULL DEFAULT 0,
  status service_order_item_status NOT NULL DEFAULT 'PENDING',
  assigned_user_id uuid REFERENCES users(id),
  started_at timestamptz,
  completed_at timestamptz,
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT sos_quantity_ck CHECK (quantity > 0),
  CONSTRAINT sos_unit_price_cents_ck CHECK (unit_price_cents >= 0),
  CONSTRAINT sos_total_price_cents_ck CHECK (total_price_cents >= 0)
);

CREATE INDEX IF NOT EXISTS sos_service_order_id_idx
  ON service_order_services (service_order_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS sos_service_id_idx
  ON service_order_services (service_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS sos_status_idx
  ON service_order_services (status)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS sos_assigned_user_id_idx
  ON service_order_services (assigned_user_id)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS service_order_parts (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  service_order_id uuid NOT NULL REFERENCES service_orders(id) ON DELETE CASCADE,
  part_id uuid REFERENCES parts(id),
  budget_part_id uuid REFERENCES budget_parts(id),
  description varchar(255) NOT NULL,
  quantity int NOT NULL DEFAULT 1,
  unit_price_cents bigint NOT NULL DEFAULT 0,
  total_price_cents bigint NOT NULL DEFAULT 0,
  status service_order_item_status NOT NULL DEFAULT 'PENDING',
  notes text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT sop_quantity_ck CHECK (quantity > 0),
  CONSTRAINT sop_unit_price_cents_ck CHECK (unit_price_cents >= 0),
  CONSTRAINT sop_total_price_cents_ck CHECK (total_price_cents >= 0)
);

CREATE INDEX IF NOT EXISTS sop_service_order_id_idx
  ON service_order_parts (service_order_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS sop_part_id_idx
  ON service_order_parts (part_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS sop_status_idx
  ON service_order_parts (status)
  WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS stock_movements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  part_id uuid NOT NULL REFERENCES parts(id),
  movement_type stock_movement_type NOT NULL,
  quantity int NOT NULL,
  unit_cost_cents bigint,
  reference_type stock_reference_type,
  reference_id uuid,
  notes text,
  created_by_user_id uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  CONSTRAINT stock_movements_quantity_ck CHECK (quantity > 0),
  CONSTRAINT stock_movements_unit_cost_cents_ck CHECK (
    unit_cost_cents IS NULL OR unit_cost_cents >= 0
  )
);

CREATE INDEX IF NOT EXISTS stock_movements_part_id_idx
  ON stock_movements (part_id)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS stock_movements_created_at_idx
  ON stock_movements (created_at)
  WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS stock_movements_reference_idx
  ON stock_movements (reference_type, reference_id)
  WHERE deleted_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS service_order_parts;
DROP TABLE IF EXISTS service_order_services;
DROP TABLE IF EXISTS budget_parts;
DROP TABLE IF EXISTS budget_services;
DROP TABLE IF EXISTS budgets;
DROP TABLE IF EXISTS service_order_status_history;
DROP TABLE IF EXISTS service_orders;
DROP TABLE IF EXISTS parts;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS vehicles;
DROP TABLE IF EXISTS clients;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS stock_reference_type;
DROP TYPE IF EXISTS stock_movement_type;
DROP TYPE IF EXISTS service_order_item_status;
DROP TYPE IF EXISTS budget_status;
DROP TYPE IF EXISTS service_order_status;
DROP TYPE IF EXISTS client_document_type;
DROP TYPE IF EXISTS user_role;

-- +goose StatementEnd