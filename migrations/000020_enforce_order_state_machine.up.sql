-- Migration 000020: Enforce Order State Machine and Business Rules
-- Prevents illegal state transitions and validates driver assignment

-- Step 1: Create function to validate state transitions
CREATE OR REPLACE FUNCTION validate_order_state_transition()
RETURNS TRIGGER AS $$
BEGIN
    -- Define valid state transitions
    -- searching_driver -> assigned, no_driver_available, cancelled
    -- assigned -> accepted, cancelled, no_driver_available
    -- accepted -> picked_up, cancelled
    -- picked_up -> in_transit, cancelled
    -- in_transit -> delivered, cancelled
    -- delivered -> (terminal state, no transitions)
    -- cancelled -> (terminal state, no transitions)

    -- Allow any transition for new inserts
    IF TG_OP = 'INSERT' THEN
        RETURN NEW;
    END IF;

    -- If status hasn't changed, allow update
    IF OLD.status = NEW.status THEN
        RETURN NEW;
    END IF;

    -- Validate state transitions
    IF OLD.status = 'searching_driver' AND NEW.status NOT IN ('assigned', 'no_driver_available', 'cancelled') THEN
        RAISE EXCEPTION 'Invalid state transition from searching_driver to %', NEW.status;
    END IF;

    IF OLD.status = 'assigned' AND NEW.status NOT IN ('accepted', 'cancelled', 'no_driver_available') THEN
        RAISE EXCEPTION 'Invalid state transition from assigned to %', NEW.status;
    END IF;

    IF OLD.status = 'accepted' AND NEW.status NOT IN ('picked_up', 'cancelled') THEN
        RAISE EXCEPTION 'Invalid state transition from accepted to %', NEW.status;
    END IF;

    IF OLD.status = 'picked_up' AND NEW.status NOT IN ('in_transit', 'cancelled') THEN
        RAISE EXCEPTION 'Invalid state transition from picked_up to %', NEW.status;
    END IF;

    IF OLD.status = 'in_transit' AND NEW.status NOT IN ('delivered', 'cancelled') THEN
        RAISE EXCEPTION 'Invalid state transition from in_transit to %', NEW.status;
    END IF;

    -- Terminal states cannot transition
    IF OLD.status IN ('delivered', 'cancelled') THEN
        RAISE EXCEPTION 'Cannot change status from terminal state %', OLD.status;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 2: Create trigger for state validation
CREATE TRIGGER trigger_validate_order_state_transition
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION validate_order_state_transition();

-- Step 3: Create function to validate driver assignment
CREATE OR REPLACE FUNCTION validate_driver_assignment()
RETURNS TRIGGER AS $$
DECLARE
    driver_record RECORD;
BEGIN
    -- If driver_id is NULL, no validation needed
    IF NEW.driver_id IS NULL THEN
        RETURN NEW;
    END IF;

    -- If driver hasn't changed, skip validation
    IF TG_OP = 'UPDATE' AND OLD.driver_id = NEW.driver_id THEN
        RETURN NEW;
    END IF;

    -- Validate driver exists and has correct role
    SELECT id, role, phone_verified, account_status
    INTO driver_record
    FROM users
    WHERE id = NEW.driver_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Driver with id % does not exist', NEW.driver_id;
    END IF;

    IF driver_record.role != 'driver' THEN
        RAISE EXCEPTION 'User % is not a driver (role: %)', NEW.driver_id, driver_record.role;
    END IF;

    IF driver_record.phone_verified != true THEN
        RAISE EXCEPTION 'Driver % phone is not verified', NEW.driver_id;
    END IF;

    IF driver_record.account_status != 'active' THEN
        RAISE EXCEPTION 'Driver % account is not active (status: %)', NEW.driver_id, driver_record.account_status;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 4: Create trigger for driver validation
CREATE TRIGGER trigger_validate_driver_assignment
    BEFORE INSERT OR UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION validate_driver_assignment();

-- Step 5: Create function to auto-update timestamps on status change
CREATE OR REPLACE FUNCTION update_order_status_timestamps()
RETURNS TRIGGER AS $$
BEGIN
    -- Only run on status changes
    IF TG_OP = 'INSERT' OR OLD.status != NEW.status THEN
        CASE NEW.status
            WHEN 'assigned' THEN
                NEW.assigned_at = NOW();
            WHEN 'accepted' THEN
                NEW.accepted_at = NOW();
            WHEN 'picked_up' THEN
                NEW.picked_up_at = NOW();
            WHEN 'delivered' THEN
                NEW.delivered_at = NOW();
            WHEN 'cancelled' THEN
                NEW.cancelled_at = NOW();
            ELSE
                -- No special timestamp for other states
        END CASE;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 6: Create trigger for auto-updating timestamps
CREATE TRIGGER trigger_update_order_status_timestamps
    BEFORE INSERT OR UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION update_order_status_timestamps();

-- Step 7: Add comments
COMMENT ON FUNCTION validate_order_state_transition IS 'Enforces valid order state transitions (prevents illegal state jumps)';
COMMENT ON FUNCTION validate_driver_assignment IS 'Validates driver is eligible: role=driver, phone_verified=true, account_status=active';
COMMENT ON FUNCTION update_order_status_timestamps IS 'Auto-updates status timestamps (assigned_at, accepted_at, etc.) on state changes';
