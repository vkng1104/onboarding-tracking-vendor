CREATE TABLE vendors (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL CHECK (length(btrim(name)) > 0),
    region TEXT NOT NULL CHECK (length(btrim(region)) > 0),
    notes TEXT NOT NULL DEFAULT '',
    current_stage TEXT NOT NULL,
    stage_entered_at TIMESTAMPTZ NOT NULL,
    assigned_coordinator_id UUID NOT NULL REFERENCES coordinators(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE vendor_stage_transitions (
    id UUID PRIMARY KEY,
    vendor_id UUID NOT NULL REFERENCES vendors(id) ON DELETE CASCADE,
    previous_stage TEXT NOT NULL,
    new_stage TEXT NOT NULL,
    changed_by_coordinator_id UUID NOT NULL REFERENCES coordinators(id),
    changed_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT vendor_stage_transitions_stage_changed CHECK (previous_stage <> new_stage)
);

CREATE INDEX vendor_stage_transitions_vendor_changed_at_idx
    ON vendor_stage_transitions (vendor_id, changed_at DESC);
